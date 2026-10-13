package opennebula

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestGenerateHookTemplateAPI(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceOpennebulaHook().Schema, map[string]interface{}{
		"name":            "hook-api",
		"type":            "api",
		"command":         "log_new_user.rb",
		"arguments":       "$API",
		"arguments_stdin": true,
		"timeout":         30,
		"call":            "one.user.allocate",
		"tags": map[string]interface{}{
			"environment": "test",
		},
	})

	tpl, err := generateHookTemplate(d, &Configuration{
		defaultTags: map[string]interface{}{
			"owner": "terraform",
		},
	})
	if err != nil {
		t.Fatalf("generateHookTemplate returned an error: %s", err)
	}

	assertHookTemplateValue(t, tpl, "NAME", "hook-api")
	assertHookTemplateValue(t, tpl, "TYPE", "api")
	assertHookTemplateValue(t, tpl, "COMMAND", "log_new_user.rb")
	assertHookTemplateValue(t, tpl, "ARGUMENTS", "$API")
	assertHookTemplateValue(t, tpl, "ARGUMENTS_STDIN", "YES")
	assertHookTemplateValue(t, tpl, "TIMEOUT", "30")
	assertHookTemplateValue(t, tpl, "CALL", "one.user.allocate")
	assertHookTemplateValue(t, tpl, "ENVIRONMENT", "test")
	assertHookTemplateValue(t, tpl, "OWNER", "terraform")
}

func TestGenerateHookTemplateVMState(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceOpennebulaHook().Schema, map[string]interface{}{
		"name":      "hook-vm",
		"type":      "state",
		"command":   "vm-pending.rb",
		"arguments": "$TEMPLATE pending",
		"resource":  "vm",
		"on":        "custom",
		"state":     "pending",
		"lcm_state": "lcm_init",
		"remote":    true,
	})

	tpl, err := generateHookTemplate(d, &Configuration{})
	if err != nil {
		t.Fatalf("generateHookTemplate returned an error: %s", err)
	}

	assertHookTemplateValue(t, tpl, "TYPE", "state")
	assertHookTemplateValue(t, tpl, "RESOURCE", "VM")
	assertHookTemplateValue(t, tpl, "ON", "CUSTOM")
	assertHookTemplateValue(t, tpl, "STATE", "PENDING")
	assertHookTemplateValue(t, tpl, "LCM_STATE", "LCM_INIT")
	assertHookTemplateValue(t, tpl, "REMOTE", "YES")
}

func TestGenerateHookTemplateRejectsReservedTags(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceOpennebulaHook().Schema, map[string]interface{}{
		"name":    "hook-api",
		"type":    "api",
		"command": "log_new_user.rb",
		"call":    "one.user.allocate",
		"tags": map[string]interface{}{
			"call": "reserved",
		},
	})

	if _, err := generateHookTemplate(d, &Configuration{}); err == nil {
		t.Fatalf("generateHookTemplate did not reject reserved tag key")
	}
}

func TestHookTemplateTagsSkipsHookAttributes(t *testing.T) {
	d := schema.TestResourceDataRaw(t, resourceOpennebulaHook().Schema, map[string]interface{}{
		"name":    "hook-api",
		"type":    "api",
		"command": "log_new_user.rb",
		"call":    "one.user.allocate",
		"tags": map[string]interface{}{
			"environment": "test",
		},
	})

	tpl, err := generateHookTemplate(d, &Configuration{})
	if err != nil {
		t.Fatalf("generateHookTemplate returned an error: %s", err)
	}

	tags := hookTemplateTags(tpl.Template)
	if _, ok := tags["CALL"]; ok {
		t.Fatalf("CALL was returned as tag")
	}
	if tags["ENVIRONMENT"] != "test" {
		t.Fatalf("ENVIRONMENT tag mismatch: got %q", tags["ENVIRONMENT"])
	}
}

func assertHookTemplateValue(t *testing.T, tpl interface {
	GetStr(string) (string, error)
}, key string, want string) {
	t.Helper()

	got, err := tpl.GetStr(key)
	if err != nil {
		t.Fatalf("missing template key %s: %s", key, err)
	}
	if got != want {
		t.Fatalf("template key %s mismatch: got %q, want %q", key, got, want)
	}
}
