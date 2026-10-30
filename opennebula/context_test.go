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

func TestFlattenContextClearsMissingContext(t *testing.T) {
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"context": contextSchema(),
	}, map[string]interface{}{
		"context": map[string]interface{}{"NETWORK": "YES"},
	})
	if err := flattenContext(d, vm.NewTemplate()); err != nil {
		t.Fatal(err)
	}
	if got := d.Get("context").(map[string]interface{}); len(got) != 0 {
		t.Fatalf("context = %#v, want empty", got)
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

func TestFlattenContextKeepsConfiguredPassword(t *testing.T) {
	const configured = "Secret123"
	// Ciphertext shape returned by OpenNebula for CONTEXT/PASSWORD.
	const remote = "mU18gesj9B80f/iOrZrg8A=="

	tpl := vm.NewTemplate()
	if err := tpl.AddCtx(vmk.Context("PASSWORD"), remote); err != nil {
		t.Fatal(err)
	}
	if err := tpl.AddCtx(vmk.Context("NETWORK"), "YES"); err != nil {
		t.Fatal(err)
	}
	if err := tpl.AddCtx(vmk.Context("USERNAME"), "Administrator"); err != nil {
		t.Fatal(err)
	}

	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"context": contextSchema(),
	}, map[string]interface{}{
		"context": map[string]interface{}{
			"PASSWORD": "Secret123",
			"NETWORK":  "YES",
			"USERNAME": "Administrator",
		},
	})

	if err := flattenContext(d, tpl); err != nil {
		t.Fatal(err)
	}

	got := d.Get("context").(map[string]interface{})
	if got["PASSWORD"] != configured {
		t.Fatalf("PASSWORD = %q, want configured value", got["PASSWORD"])
	}
	if got["NETWORK"] != "YES" || got["USERNAME"] != "Administrator" {
		t.Fatalf("context = %#v", got)
	}
}

func TestFlattenContextReadsUnencryptedChanges(t *testing.T) {
	tpl := vm.NewTemplate()
	if err := tpl.AddCtx(vmk.Context("PASSWORD"), "mU18gesj9B80f/iOrZrg8A=="); err != nil {
		t.Fatal(err)
	}
	if err := tpl.AddCtx(vmk.Context("NETWORK"), "NO"); err != nil {
		t.Fatal(err)
	}

	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"context": contextSchema(),
	}, map[string]interface{}{
		"context": map[string]interface{}{
			"PASSWORD": "Secret123",
			"NETWORK":  "YES",
		},
	})

	if err := flattenContext(d, tpl); err != nil {
		t.Fatal(err)
	}

	got := d.Get("context").(map[string]interface{})
	if got["NETWORK"] != "NO" {
		t.Fatalf("NETWORK = %#v, want remote NO", got["NETWORK"])
	}
	if got["PASSWORD"] != "Secret123" {
		t.Fatalf("PASSWORD = %q, want configured value", got["PASSWORD"])
	}
}

func TestFlattenContextAdoptsUnencryptedContextValues(t *testing.T) {
	remote := map[string]string{
		"PASSWORD_BASE64":         "Y2hhbmdlZA==",
		"CRYPTED_PASSWORD_BASE64": "Y2hhbmdlZA==",
		"USERNAME":                "Administrator",
		"SSH_PUBLIC_KEY":          "ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIexample",
		"START_SCRIPT_BASE64":     "Y2hhbmdlZA==",
	}
	configured := map[string]interface{}{
		"PASSWORD_BASE64":         "c2VjcmV0",
		"CRYPTED_PASSWORD_BASE64": "c2VjcmV0",
		"USERNAME":                "root",
		"SSH_PUBLIC_KEY":          "$USER[SSH_PUBLIC_KEY]",
		"START_SCRIPT_BASE64":     "c2VjcmV0",
	}

	tpl := vm.NewTemplate()
	for key, value := range remote {
		if err := tpl.AddCtx(vmk.Context(key), value); err != nil {
			t.Fatal(err)
		}
	}

	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"context": contextSchema(),
	}, map[string]interface{}{
		"context": configured,
	})
	if err := flattenContext(d, tpl); err != nil {
		t.Fatal(err)
	}

	got := d.Get("context").(map[string]interface{})
	for key, value := range remote {
		if got[key] != value {
			t.Fatalf("%s = %#v, want remote %q", key, got[key], value)
		}
	}
}

func TestTemplateReplaceSendsConfiguredPassword(t *testing.T) {
	const configured = "Secret123"
	const remote = "mU18gesj9B80f/iOrZrg8A=="

	prior := baseTemplateConfig()
	prior["context"] = map[string]interface{}{
		"NETWORK":  "YES",
		"PASSWORD": configured,
	}
	next := withChange(prior, "description", "updated")

	tpl := createdTemplate(t, prior)
	// Simulate the ciphertext OpenNebula returns in place of the configured password.
	vec, err := tpl.GetVector(string(vmk.ContextVec))
	if err != nil {
		t.Fatal(err)
	}
	replaced := false
	for i := range vec.Pairs {
		if vec.Pairs[i].Key() == "PASSWORD" {
			vec.Pairs[i].Value = remote
			replaced = true
		}
	}
	if !replaced {
		t.Fatal("created template has no PASSWORD")
	}

	d := templateUpdateData(t, prior, next)
	updated, diags := applyTemplateContentChanges(d, tpl)
	if len(diags) > 0 {
		t.Fatal(diags)
	}
	if !updated {
		t.Fatal("description change was not marked for replace")
	}
	got, err := tpl.GetStrFromVec(string(vmk.ContextVec), "PASSWORD")
	if err != nil {
		t.Fatal(err)
	}
	if got != configured {
		t.Fatalf("PASSWORD sent to OpenNebula = %q, want configured value", got)
	}
}

func TestFlattenTagsKeepsOnePassword(t *testing.T) {
	tpl := vm.NewTemplate()
	tpl.AddPair("ONE_PASSWORD", "mU18gesj9B80f/iOrZrg8A==")
	tpl.AddPair("LOGO", "windows8.png")

	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"tags":         tagsSchema(),
		"default_tags": defaultTagsSchemaComputed(),
		"tags_all":     tagsSchemaComputed(),
	}, map[string]interface{}{
		"tags": map[string]interface{}{
			"one_password": "plain-secret",
			"logo":         "ubuntu.png",
		},
	})

	if diags := flattenTemplateTags(d, &Configuration{}, &tpl.Template); len(diags) > 0 {
		t.Fatal(diags)
	}

	tags := d.Get("tags").(map[string]interface{})
	if tags["one_password"] != "plain-secret" {
		t.Fatalf("one_password = %#v, want configured value", tags["one_password"])
	}
	if tags["logo"] != "windows8.png" {
		t.Fatalf("logo = %#v, want remote windows8.png", tags["logo"])
	}
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
