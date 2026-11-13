package opennebula

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/securitygroup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type securityGroupUpdateTransport func(*http.Request) (*http.Response, error)

func (f securityGroupUpdateTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestSecurityGroupUpdateRegression(t *testing.T) {
	allowAll := []map[string]string{
		{"PROTOCOL": "ALL", "RULE_TYPE": "INBOUND"},
		{"PROTOCOL": "ALL", "RULE_TYPE": "OUTBOUND"},
	}
	loopback := []map[string]string{
		{"PROTOCOL": "TCP", "RULE_TYPE": "INBOUND", "IP": "127.0.0.1", "RANGE": "9", "SIZE": "1"},
	}
	routerOld := []map[string]string{
		{"PROTOCOL": "ALL", "RULE_TYPE": "INBOUND", "IP": "224.0.0.0", "SIZE": "256"},
	}
	routerNew := []map[string]string{
		{"PROTOCOL": "ALL", "RULE_TYPE": "INBOUND", "IP": "224.0.0.18", "SIZE": "1"},
		{"PROTOCOL": "ALL", "RULE_TYPE": "OUTBOUND"},
	}
	for _, tc := range []struct {
		name        string
		id          int
		oldRules    []map[string]string
		newRules    []map[string]string
		description string
		commit      bool
		reject      bool
	}{
		{name: "default_rules_commit", oldRules: allowAll, newRules: loopback, description: "old", commit: true},
		{name: "default_rules_no_commit", oldRules: allowAll, newRules: loopback, description: "old"},
		{name: "router_replace_and_add", id: 101, oldRules: routerOld, newRules: routerNew, description: "old", commit: true},
		{name: "description_only", oldRules: allowAll, newRules: allowAll, description: "new", commit: true},
		{name: "description_clear", oldRules: allowAll, newRules: allowAll, description: "", commit: true},
		{name: "description_and_rules", oldRules: allowAll, newRules: loopback, description: "new", commit: true},
		{name: "rules_update_rejected", oldRules: allowAll, newRules: loopback, description: "old", commit: true, reject: true},
		{name: "description_update_rejected", oldRules: allowAll, newRules: allowAll, description: "new", commit: true, reject: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rulesChanged := !reflect.DeepEqual(tc.oldRules, tc.newRules)
			backendDescription := "old"
			backendRules := tc.oldRules
			var mutations []string
			updates := 0
			transport := securityGroupUpdateTransport(func(req *http.Request) (*http.Response, error) {
				var call struct {
					Method string `xml:"methodName"`
					Params []struct {
						Value struct {
							String  string `xml:"string"`
							Int     int    `xml:"int"`
							Boolean string `xml:"boolean"`
						} `xml:"value"`
					} `xml:"params>param"`
				}
				if err := xml.NewDecoder(req.Body).Decode(&call); err != nil {
					return nil, err
				}
				if len(call.Params) < 2 || call.Params[1].Value.Int != tc.id {
					return nil, fmt.Errorf("%s: expected security group ID %d", call.Method, tc.id)
				}
				success, value := "1", "<int>0</int>"
				switch call.Method {
				case "one.secgroup.info":
					info := fmt.Sprintf(`<SECURITY_GROUP><ID>%d</ID><NAME>test-sg</NAME><UID>0</UID><GID>0</GID><UNAME>oneadmin</UNAME><GNAME>oneadmin</GNAME><PERMISSIONS><OWNER_U>1</OWNER_U><OWNER_M>1</OWNER_M><GROUP_U>1</GROUP_U></PERMISSIONS><TEMPLATE><DESCRIPTION>%s</DESCRIPTION><CUSTOM_TAG>keep</CUSTOM_TAG><CUSTOM_VECTOR><VALUE>untouched</VALUE></CUSTOM_VECTOR>%s</TEMPLATE></SECURITY_GROUP>`, tc.id, backendDescription, securityGroupRulesXML(backendRules))
					var escaped bytes.Buffer
					if err := xml.EscapeText(&escaped, []byte(info)); err != nil {
						return nil, err
					}
					value = "<string>" + escaped.String() + "</string>"
				case "one.secgroup.update":
					updates++
					mutations = append(mutations, call.Method)
					if len(call.Params) != 4 || call.Params[3].Value.Int != 0 {
						return nil, fmt.Errorf("expected replacement update with mode 0")
					}
					sent := call.Params[2].Value.String
					for _, expected := range []string{`CUSTOM_TAG="keep"`, `CUSTOM_VECTOR=[`, `VALUE="untouched"`} {
						if !strings.Contains(sent, expected) {
							t.Errorf("update missing %s: %s", expected, sent)
						}
					}
					rules := securityGroupSentRules(t, sent)
					if !reflect.DeepEqual(rules, tc.newRules) {
						t.Errorf("updated rules = %#v, want %#v; template: %s", rules, tc.newRules, sent)
					}
					descriptions := regexp.MustCompile(`DESCRIPTION="([^"]*)"`).FindAllStringSubmatch(sent, -1)
					description := ""
					if len(descriptions) > 1 {
						t.Errorf("duplicate description in template: %s", sent)
					}
					if len(descriptions) > 0 {
						description = descriptions[0][1]
					}
					if description != tc.description {
						t.Errorf("updated description = %q, want %q", description, tc.description)
					}
					if tc.reject {
						success, value = "0", "<string>Security group update rejected</string>"
					} else {
						backendDescription, backendRules = description, rules
					}
				case "one.secgroup.commit":
					mutations = append(mutations, call.Method)
					if len(call.Params) != 3 || call.Params[2].Value.Boolean != "1" {
						t.Errorf("commit must update only outdated VMs")
					}
				default:
					return nil, fmt.Errorf("unexpected XML-RPC method %s", call.Method)
				}
				body := `<methodResponse><params><param><value><array><data><value><boolean>` + success + `</boolean></value><value>` + value + `</value><value><int>0</int></value></data></array></value></param></params></methodResponse>`
				return &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}, nil
			})
			cfg := &Configuration{Controller: goca.NewController(goca.NewClient(goca.NewConfig("test", "test", "http://one.example/RPC2"), &http.Client{Transport: transport}))}
			r := resourceOpennebulaSecurityGroup()
			old := schema.TestResourceDataRaw(t, r.Schema, map[string]interface{}{
				"name": "test-sg", "description": "old", "rule": securityGroupRuleConfig(tc.oldRules), "commit": tc.commit,
			})
			old.SetId(strconv.Itoa(tc.id))
			raw := map[string]interface{}{
				"name": "test-sg", "description": tc.description, "rule": securityGroupRuleConfig(tc.newRules), "commit": tc.commit,
			}
			diff, err := r.Diff(context.Background(), old.State(), terraform.NewResourceConfigRaw(raw), cfg)
			if err != nil {
				t.Fatal(err)
			}
			if diff == nil {
				t.Fatal("expected an update diff")
			}
			// Computed tags absent from the fixture must not trigger an unrelated template update.
			for key := range diff.Attributes {
				if key != "description" && !strings.HasPrefix(key, "rule.") {
					delete(diff.Attributes, key)
				}
			}
			if tc.description == "" {
				// Optional+Computed strings treat an empty configuration as unset in the
				// legacy SDK. Exercise clearing at the Apply boundary with an explicit diff.
				diff.Attributes["description"] = &terraform.ResourceAttrDiff{Old: "old", New: ""}
			}
			if len(diff.Attributes) == 0 || diff.RequiresNew() {
				t.Fatalf("expected a nonempty in-place update: %#v", diff)
			}
			state, diags := r.Apply(context.Background(), old.State(), diff, cfg)
			if diags.HasError() != tc.reject {
				t.Errorf("diagnostics = %v, want error=%t", diags, tc.reject)
			}
			if tc.reject && !strings.Contains(fmt.Sprint(diags), "Security group update rejected") {
				t.Errorf("diagnostics must include backend failure: %v", diags)
			}
			if updates != 1 {
				t.Errorf("got %d updates, want 1", updates)
			}
			wantMutations := []string{"one.secgroup.update"}
			if rulesChanged && tc.commit && !tc.reject {
				wantMutations = append(wantMutations, "one.secgroup.commit")
			}
			if !reflect.DeepEqual(mutations, wantMutations) {
				t.Errorf("mutation order = %v, want %v", mutations, wantMutations)
			}
			if !tc.reject {
				if state == nil || state.ID != strconv.Itoa(tc.id) {
					t.Fatalf("unexpected state: %#v", state)
				}
				if state.Attributes["description"] != tc.description {
					t.Errorf("refreshed description = %q, want %q", state.Attributes["description"], tc.description)
				}
				if state.Attributes["rule.#"] != strconv.Itoa(len(tc.newRules)) {
					t.Errorf("refreshed rule count = %q", state.Attributes["rule.#"])
				}
				for i, rule := range tc.newRules {
					for key, want := range rule {
						key = fmt.Sprintf("rule.%d.%s", i, strings.ToLower(key))
						if state.Attributes[key] != want {
							t.Errorf("refreshed %s = %q, want %q", key, state.Attributes[key], want)
						}
					}
				}
			}
		})
	}
}

func securityGroupRuleConfig(rules []map[string]string) []interface{} {
	result := make([]interface{}, 0, len(rules))
	for _, rule := range rules {
		config := make(map[string]interface{}, len(rule))
		for key, value := range rule {
			config[strings.ToLower(key)] = value
		}
		result = append(result, config)
	}
	return result
}

func securityGroupRulesXML(rules []map[string]string) string {
	var result strings.Builder
	for _, rule := range rules {
		result.WriteString("<RULE>")
		for key, value := range rule {
			fmt.Fprintf(&result, "<%s>%s</%s>", key, value, key)
		}
		result.WriteString("</RULE>")
	}
	return result.String()
}

func securityGroupSentRules(t *testing.T, sent string) []map[string]string {
	t.Helper()
	var rules []map[string]string
	for _, vector := range regexp.MustCompile(`(?s)\bRULE=\[(.*?)\]`).FindAllStringSubmatch(sent, -1) {
		rule := make(map[string]string)
		for _, pair := range regexp.MustCompile(`([A-Z_]+)="([^"]*)"`).FindAllStringSubmatch(vector[1], -1) {
			if _, exists := rule[pair[1]]; exists {
				t.Errorf("duplicate rule attribute %s in %s", pair[1], vector[0])
			}
			rule[pair[1]] = pair[2]
		}
		rules = append(rules, rule)
	}
	return rules
}

func TestSecurityGroupReadNormalizesRuleEnums(t *testing.T) {
	var template securitygroup.Template
	if err := xml.Unmarshal([]byte(`<TEMPLATE><RULE><PROTOCOL>All</PROTOCOL><RULE_TYPE>inbound</RULE_TYPE><IP>224.0.0.18</IP><SIZE>1</SIZE></RULE><RULE><PROTOCOL>tCp</PROTOCOL><RULE_TYPE>Outbound</RULE_TYPE><IP>127.0.0.1</IP><RANGE>9</RANGE><SIZE>1</SIZE></RULE></TEMPLATE>`), &template); err != nil {
		t.Fatal(err)
	}
	got := generateSecurityGroupMapFromStructs(template.GetRules())
	want := []map[string]interface{}{
		{"protocol": "ALL", "rule_type": "INBOUND", "ip": "224.0.0.18", "size": "1"},
		{"protocol": "TCP", "rule_type": "OUTBOUND", "ip": "127.0.0.1", "range": "9", "size": "1"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("flattened rules = %#v, want %#v", got, want)
	}
}
