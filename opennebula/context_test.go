package opennebula

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

// oneContextStartScript is a Windows START_SCRIPT as executed by one-context.
// The quotes and backslashes are what OpenNebula escapes in CONTEXT values.
const oneContextStartScript = `Enable-NetFirewallRule -DisplayGroup "Remote Desktop"
Set-ItemProperty -Path "HKLM:\System\CurrentControlSet\Control\Terminal Server" -Name "fDenyTSConnections" -Value 0`

func TestAddContextReplacesContextVector(t *testing.T) {
	tpl := vm.NewTemplate()
	if err := tpl.AddCtx(vmk.Context("EJECT_CDROM"), "YES"); err != nil {
		t.Fatal(err)
	}
	if err := tpl.AddCtx(vmk.Context("NETWORK"), "YES"); err != nil {
		t.Fatal(err)
	}

	addContext(tpl, map[string]interface{}{
		"EJECT_CDROM": "YS",
		"network":     "YES",
	})

	if got, _ := tpl.GetStrFromVec(string(vmk.ContextVec), "EJECT_CDROM"); got != "YS" {
		t.Fatalf("EJECT_CDROM = %q, want YS", got)
	}
	if got, _ := tpl.GetStrFromVec(string(vmk.ContextVec), "NETWORK"); got != "YES" {
		t.Fatalf("NETWORK = %q, want YES", got)
	}
	if strings.Count(tpl.String(), "CONTEXT=") != 1 {
		t.Fatalf("template = %s", tpl.String())
	}
}

func TestFlattenContextReadsRemoteValue(t *testing.T) {
	tpl := vm.NewTemplate()
	if err := tpl.AddCtx(vmk.Context("EJECT_CDROM"), "YES"); err != nil {
		t.Fatal(err)
	}

	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"context": contextSchema(),
	}, map[string]interface{}{
		"context": map[string]interface{}{
			"EJECT_CDROM": "YS",
		},
	})

	if err := flattenContext(d, tpl); err != nil {
		t.Fatal(err)
	}

	got := d.Get("context").(map[string]interface{})
	if got["EJECT_CDROM"] != "YES" {
		t.Fatalf("EJECT_CDROM = %#v, want remote YES", got["EJECT_CDROM"])
	}
}

func TestFlattenContextIgnoresOpenNebulaEscaping(t *testing.T) {
	configured := oneContextStartScript + "\n"
	remote := escapeTemplateValue(strings.TrimSpace(configured))
	if unescapeTemplateValue(remote) != strings.TrimSpace(configured) {
		t.Fatalf("unescape(escape(config)) = %q", unescapeTemplateValue(remote))
	}

	tpl := vm.NewTemplate()
	if err := tpl.AddCtx(vmk.Context("START_SCRIPT"), remote); err != nil {
		t.Fatal(err)
	}

	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"context": contextSchema(),
	}, map[string]interface{}{
		"context": map[string]interface{}{
			"START_SCRIPT": configured,
		},
	})

	if err := flattenContext(d, tpl); err != nil {
		t.Fatal(err)
	}

	got := d.Get("context").(map[string]interface{})
	if got["START_SCRIPT"] != configured {
		t.Fatalf("START_SCRIPT changed:\n got %q\nwant %q", got["START_SCRIPT"], configured)
	}
}

func TestDecodeContextValuesBeforeReserialize(t *testing.T) {
	configured := strings.TrimSpace(oneContextStartScript)

	created := vm.NewTemplate()
	addContext(created, map[string]interface{}{
		"NETWORK":      "YES",
		"START_SCRIPT": configured,
	})

	fetched := vm.NewTemplate()
	if err := fetched.AddCtx(vmk.Context("NETWORK"), "YES"); err != nil {
		t.Fatal(err)
	}
	if err := fetched.AddCtx(vmk.Context("START_SCRIPT"), escapeTemplateValue(configured)); err != nil {
		t.Fatal(err)
	}

	if got, want := contextPairString(fetched, "START_SCRIPT"), contextPairString(created, "START_SCRIPT"); got == want {
		t.Fatal("expected fetched START_SCRIPT to gain an extra escape layer")
	}

	decodeContextValues(fetched)

	if got, want := contextPairString(fetched, "START_SCRIPT"), contextPairString(created, "START_SCRIPT"); got != want {
		t.Fatalf("START_SCRIPT serialization:\n got %s\nwant %s", got, want)
	}
	if got, _ := fetched.GetStrFromVec(string(vmk.ContextVec), "NETWORK"); got != "YES" {
		t.Fatalf("NETWORK = %q", got)
	}
	if got, _ := fetched.GetStrFromVec(string(vmk.ContextVec), "START_SCRIPT"); got != configured {
		t.Fatalf("START_SCRIPT = %q, want configured value", got)
	}
}

func contextPairString(tpl *vm.Template, key string) string {
	vec, err := tpl.GetVector(string(vmk.ContextVec))
	if err != nil {
		return ""
	}
	for i := range vec.Pairs {
		if vec.Pairs[i].Key() == key {
			return vec.Pairs[i].String()
		}
	}
	return ""
}

func TestFlattenContextIgnoresTrimmedWhitespace(t *testing.T) {
	configured := "Enable-NetFirewallRule -DisplayGroup \"Remote Desktop\"\n"
	tpl := vm.NewTemplate()
	if err := tpl.AddCtx(vmk.Context("START_SCRIPT"), strings.TrimSpace(configured)); err != nil {
		t.Fatal(err)
	}

	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"context": contextSchema(),
	}, map[string]interface{}{
		"context": map[string]interface{}{
			"START_SCRIPT": configured,
		},
	})

	if err := flattenContext(d, tpl); err != nil {
		t.Fatal(err)
	}

	got := d.Get("context").(map[string]interface{})
	if got["START_SCRIPT"] != configured {
		t.Fatalf("START_SCRIPT = %q, want configured value", got["START_SCRIPT"])
	}
}
