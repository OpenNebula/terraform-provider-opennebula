package opennebula

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var (
	xmlrpcMethodRe = regexp.MustCompile(`<methodName>([^<]+)</methodName>`)
	xmlrpcStringRe = regexp.MustCompile(`<string>([^<]*)</string>`)
)

// fakeOned serves one.vm.info and one.vm.action for a single VM, returning
// the queued states in order and repeating the last one.
type fakeOned struct {
	mu      sync.Mutex
	states  [][2]int
	after   [][2]int
	actions []string
}

func (f *fakeOned) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	method := xmlrpcMethodRe.FindSubmatch(body)
	if method == nil {
		http.Error(w, "no methodName", http.StatusBadRequest)
		return
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	var out string
	switch string(method[1]) {
	case "one.vm.info":
		s := f.states[0]
		if len(f.states) > 1 {
			f.states = f.states[1:]
		}
		out = fmt.Sprintf("<VM><ID>42</ID><NAME>failed</NAME><STATE>%d</STATE><LCM_STATE>%d</LCM_STATE>"+
			"<USER_TEMPLATE><ERROR>DEPLOY: driver error</ERROR></USER_TEMPLATE></VM>", s[0], s[1])
	case "one.vm.action":
		// args are the session token, then the action name
		args := xmlrpcStringRe.FindAllSubmatch(body, -1)
		if len(args) < 2 {
			http.Error(w, "no action argument", http.StatusBadRequest)
			return
		}
		f.actions = append(f.actions, string(args[1][1]))
		f.states = f.after
		out = "42"
	default:
		http.Error(w, "unexpected method", http.StatusBadRequest)
		return
	}

	var escaped bytes.Buffer
	_ = xml.EscapeText(&escaped, []byte(out))
	fmt.Fprintf(w, "<?xml version=\"1.0\"?><methodResponse><params><param><value><array><data>"+
		"<value><boolean>1</boolean></value><value><string>%s</string></value><value><i4>0</i4></value>"+
		"</data></array></value></param></params></methodResponse>", escaped.String())
}

func newFakeOnedConfig(t *testing.T, f *fakeOned) *Configuration {
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)

	client, err := goca.NewClientFromConfig(goca.NewConfig("oneadmin", "pass", srv.URL))
	if err != nil {
		t.Fatal(err)
	}
	return &Configuration{Controller: goca.NewController(client)}
}

func TestVirtualMachineDeleteFromFailureState(t *testing.T) {
	failed := [2]int{int(vm.Active), int(vm.BootFailure)}
	f := &fakeOned{
		states: [][2]int{failed},
		// terminate is asynchronous, so the failure state is seen once more
		after: [][2]int{failed, {int(vm.Done), 0}},
	}

	d := schema.TestResourceDataRaw(t, resourceOpennebulaVirtualMachine().Schema, map[string]interface{}{})
	d.SetId("42")

	diags := resourceOpennebulaVirtualMachineDelete(context.Background(), d, newFakeOnedConfig(t, f))
	if diags.HasError() {
		t.Fatalf("delete failed: %+v", diags)
	}
	if len(f.actions) != 1 || f.actions[0] != "terminate" {
		t.Fatalf("expected one terminate action, got %v", f.actions)
	}
}

func TestWaitForVMStatesUnexpectedFailure(t *testing.T) {
	f := &fakeOned{states: [][2]int{{int(vm.Active), int(vm.BootFailure)}}}
	vmc := newFakeOnedConfig(t, f).Controller.VM(42)

	conf := NewVMUpdateStateConf(defaultVMTimeout, []string{}, NewVMLCMState(vm.Running).ToStrings())
	_, err := waitForVMStates(context.Background(), vmc, conf)
	if err == nil || !strings.Contains(err.Error(), "entered fail state") {
		t.Fatalf("expected a fail state error, got %v", err)
	}
}
