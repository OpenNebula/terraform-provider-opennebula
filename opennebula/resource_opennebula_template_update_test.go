package opennebula

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

func baseTemplateConfig() map[string]interface{} {
	return map[string]interface{}{
		"name":        "test-template",
		"cpu":         1.0,
		"vcpu":        1,
		"memory":      512,
		"description": "created",
		"context": map[string]interface{}{
			"NETWORK": "YES",
		},
		"features": []interface{}{
			map[string]interface{}{
				"acpi": "YES",
			},
		},
		"cpumodel": []interface{}{
			map[string]interface{}{
				"model": "host-passthrough",
			},
		},
		"graphics": []interface{}{
			map[string]interface{}{
				"listen": "0.0.0.0",
				"type":   "VNC",
			},
		},
		"os": []interface{}{
			map[string]interface{}{
				"arch":            "x86_64",
				"boot":            "disk0",
				"machine":         "q35",
				"firmware":        "UEFI",
				"firmware_secure": false,
			},
		},
		"disk": []interface{}{
			map[string]interface{}{
				"image_id": 10,
				"size":     20480,
			},
		},
		"user_inputs": map[string]interface{}{
			"CLUSTER_ID": "M|text|Cluster| |",
		},
		"sched_requirements":    "HYPERVISOR=kvm",
		"sched_ds_requirements": "NAME=default",
		"tags": map[string]interface{}{
			"logo": "ubuntu.png",
		},
		"raw": []interface{}{
			map[string]interface{}{
				"type": "kvm",
				"data": "<clock offset='utc'/>",
			},
		},
		"template_section": []interface{}{
			map[string]interface{}{
				"name": "input",
				"elements": map[string]interface{}{
					"TYPE": "tablet",
					"BUS":  "usb",
				},
			},
		},
	}
}

func withChange(base map[string]interface{}, key string, value interface{}) map[string]interface{} {
	next := make(map[string]interface{}, len(base)+1)
	for k, v := range base {
		next[k] = v
	}
	if value == nil {
		delete(next, key)
	} else {
		next[key] = value
	}
	return next
}

func templateUpdateData(t *testing.T, prior, next map[string]interface{}) *schema.ResourceData {
	t.Helper()

	ctx := context.Background()
	res := resourceOpennebulaTemplate()
	meta := &Configuration{}

	createDiff, err := res.Diff(ctx, nil, terraform.NewResourceConfigRaw(prior), meta)
	if err != nil {
		t.Fatalf("prior diff: %s", err)
	}
	state := (&terraform.InstanceState{ID: "1"}).MergeDiff(createDiff)
	// Computed attributes are stored as Terraform's unknown sentinel. A later
	// diff tries to parse that sentinel as a number.
	for key, value := range state.Attributes {
		if value == "74D93920-ED26-11E3-AC10-0800200C9A66" {
			delete(state.Attributes, key)
		}
	}

	nextDiff, err := res.Diff(ctx, state, terraform.NewResourceConfigRaw(next), meta)
	if err != nil {
		t.Fatalf("update diff: %s", err)
	}

	d, err := schema.InternalMap(res.Schema).Data(state, nextDiff)
	if err != nil {
		t.Fatal(err)
	}
	d.SetId("1")
	return d
}

func createdTemplate(t *testing.T, config map[string]interface{}) *vm.Template {
	t.Helper()

	d := schema.TestResourceDataRaw(t, resourceOpennebulaTemplate().Schema, config)
	tpl, err := generateTemplate(d, &Configuration{})
	if err != nil {
		t.Fatal(err)
	}
	addNICs(d, tpl)
	addNICAliases(d, tpl)
	return tpl
}

func applyCreatedTemplateUpdate(t *testing.T, prior, next map[string]interface{}) (*vm.Template, bool) {
	t.Helper()

	tpl := createdTemplate(t, prior)
	d := templateUpdateData(t, prior, next)
	updated, diags := applyTemplateContentChanges(d, tpl)
	if len(diags) > 0 {
		t.Fatalf("update diags: %v", diags)
	}
	if !updated {
		t.Fatalf("attribute change was not marked for replace; OpenNebula would keep the previous template")
	}
	return tpl, updated
}

func TestTemplateUpdateNoChange(t *testing.T) {
	cfg := baseTemplateConfig()
	tpl := createdTemplate(t, cfg)
	d := templateUpdateData(t, cfg, cfg)
	updated, diags := applyTemplateContentChanges(d, tpl)
	if len(diags) > 0 {
		t.Fatal(diags)
	}
	if updated {
		t.Fatal("unchanged template was marked for replace")
	}
}

func TestTemplateUpdateAddsNIC(t *testing.T) {
	prior := baseTemplateConfig()
	next := withChange(prior, "nic", []interface{}{
		map[string]interface{}{
			"network_id": 42,
			"model":      "virtio",
		},
	})

	tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
	body := tpl.String()
	if strings.Count(body, "NIC=[") != 1 || !strings.Contains(body, `NETWORK_ID="42"`) || !strings.Contains(body, `MODEL="virtio"`) {
		t.Fatalf("updated template missing NIC:\n%s", body)
	}
}

func TestTemplateUpdateChangesNIC(t *testing.T) {
	prior := withChange(baseTemplateConfig(), "nic", []interface{}{
		map[string]interface{}{
			"network_id": 7,
			"ip":         "10.0.0.4",
		},
	})
	next := withChange(baseTemplateConfig(), "nic", []interface{}{
		map[string]interface{}{
			"network_id": 7,
			"ip":         "10.0.0.9",
		},
	})

	tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
	body := tpl.String()
	if strings.Count(body, "NIC=[") != 1 || !strings.Contains(body, `IP="10.0.0.9"`) || strings.Contains(body, `IP="10.0.0.4"`) {
		t.Fatalf("NIC IP was not replaced:\n%s", body)
	}
}

func TestTemplateUpdateRemovesNIC(t *testing.T) {
	prior := withChange(baseTemplateConfig(), "nic", []interface{}{
		map[string]interface{}{
			"network_id": 7,
		},
	})
	next := baseTemplateConfig()

	tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
	if strings.Contains(tpl.String(), "NIC=[") {
		t.Fatalf("NIC was not removed:\n%s", tpl.String())
	}
}

func TestTemplateUpdateAddsNICAlias(t *testing.T) {
	prior := withChange(baseTemplateConfig(), "nic", []interface{}{
		map[string]interface{}{
			"name":       "parent0",
			"network_id": 7,
		},
	})
	next := withChange(prior, "nic_alias", []interface{}{
		map[string]interface{}{
			"parent":     "parent0",
			"network_id": 8,
			"ip":         "10.1.0.5",
		},
	})

	tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
	body := tpl.String()
	if strings.Count(body, "NIC=[") != 1 || strings.Count(body, "NIC_ALIAS=[") != 1 || !strings.Contains(body, `IP="10.1.0.5"`) || !strings.Contains(body, `PARENT="parent0"`) {
		t.Fatalf("NIC alias was not added:\n%s", body)
	}
}

func osWithFirmwareSecure(enabled bool) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"arch":            "x86_64",
			"boot":            "disk0",
			"machine":         "q35",
			"firmware":        "UEFI",
			"firmware_secure": enabled,
		},
	}
}

func TestOSFirmwareSecureYesNo(t *testing.T) {
	for _, tc := range []struct {
		value bool
		want  string
	}{
		{true, "YES"},
		{false, "NO"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			tpl := createdTemplate(t, withChange(baseTemplateConfig(), "os", osWithFirmwareSecure(tc.value)))
			got, err := tpl.GetOS(vmk.FirmwareSecure)
			if err != nil || got != tc.want {
				t.Fatalf("firmware_secure %v: got %q err=%v, want %s", tc.value, got, err, tc.want)
			}
		})
	}
}

func TestFlattenOSFirmwareSecure(t *testing.T) {
	for _, tc := range []struct {
		stored string
		want   bool
	}{
		{"YES", true},
		{"NO", false},
		{"true", false},
		{"false", false},
	} {
		t.Run(tc.stored, func(t *testing.T) {
			readTpl := vm.NewTemplate()
			if err := readTpl.AddOS(vmk.Arch, "x86_64"); err != nil {
				t.Fatal(err)
			}
			if err := readTpl.AddOS(vmk.Boot, "disk0"); err != nil {
				t.Fatal(err)
			}
			if err := readTpl.AddOS(vmk.FirmwareSecure, tc.stored); err != nil {
				t.Fatal(err)
			}

			d := schema.TestResourceDataRaw(t, resourceOpennebulaTemplate().Schema, map[string]interface{}{
				"name": "firmware-secure",
			})
			if err := flattenTemplate(d, map[string]interface{}{}, readTpl); err != nil {
				t.Fatal(err)
			}
			osList := d.Get("os").([]interface{})
			if len(osList) != 1 {
				t.Fatalf("os length = %d, want 1", len(osList))
			}
			if got := osList[0].(map[string]interface{})["firmware_secure"]; got != tc.want {
				t.Fatalf("firmware_secure %q: got %#v, want %v", tc.stored, got, tc.want)
			}
		})
	}
}

func TestTemplateUpdateOSFirmwareSecure(t *testing.T) {
	for _, tc := range []struct {
		from bool
		to   bool
		want string
	}{
		{false, true, "YES"},
		{true, false, "NO"},
	} {
		t.Run(tc.want, func(t *testing.T) {
			prior := withChange(baseTemplateConfig(), "os", osWithFirmwareSecure(tc.from))
			next := withChange(baseTemplateConfig(), "os", osWithFirmwareSecure(tc.to))

			tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
			got, err := tpl.GetOS(vmk.FirmwareSecure)
			if err != nil || got != tc.want {
				t.Fatalf("firmware_secure = %q err=%v, want %s\n%s", got, err, tc.want, tpl.String())
			}
		})
	}
}

func TestUpdateVMOSFirmwareSecure(t *testing.T) {
	tpl := vm.NewTemplate()
	if err := tpl.AddOS(vmk.Arch, "x86_64"); err != nil {
		t.Fatal(err)
	}
	if err := tpl.AddOS(vmk.Firmware, "UEFI"); err != nil {
		t.Fatal(err)
	}
	if err := tpl.AddOS(vmk.FirmwareSecure, "NO"); err != nil {
		t.Fatal(err)
	}

	applied := map[string]interface{}{
		"arch":            "x86_64",
		"firmware":        "UEFI",
		"firmware_secure": false,
	}
	next := map[string]interface{}{
		"arch":            "x86_64",
		"firmware":        "UEFI",
		"firmware_secure": true,
	}
	if err := updateVMTemplateVec(tpl, "OS", osTemplateConfig(applied), osTemplateConfig(next)); err != nil {
		t.Fatal(err)
	}

	got, err := tpl.GetOS(vmk.FirmwareSecure)
	if err != nil || got != "YES" {
		t.Fatalf("firmware_secure = %q err=%v, want YES\n%s", got, err, tpl.String())
	}
	if strings.Contains(tpl.String(), "true") || strings.Contains(tpl.String(), "false") {
		t.Fatalf("OS update wrote a bool:\n%s", tpl.String())
	}
}

func TestTemplateUpdateOSFirmwareUEFIToBIOS(t *testing.T) {
	prior := baseTemplateConfig()
	next := withChange(prior, "os", []interface{}{
		map[string]interface{}{
			"arch":            "x86_64",
			"boot":            "disk0",
			"machine":         "q35",
			"firmware":        "BIOS",
			"firmware_secure": false,
		},
	})

	tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
	firmware, _ := tpl.GetOS(vmk.Firmware)
	if firmware != "BIOS" {
		t.Fatalf("firmware = %q, want BIOS", firmware)
	}
	arch, _ := tpl.GetOS(vmk.Arch)
	if arch != "x86_64" {
		t.Fatalf("arch = %q, want x86_64", arch)
	}
	firmwareSecure, _ := tpl.GetOS(vmk.FirmwareSecure)
	if firmwareSecure != "NO" {
		t.Fatalf("firmware_secure = %q, want NO", firmwareSecure)
	}
}

func TestTemplateUpdateClearsOSFirmware(t *testing.T) {
	prior := baseTemplateConfig()
	next := withChange(prior, "os", []interface{}{
		map[string]interface{}{
			"arch":            "x86_64",
			"boot":            "disk0",
			"machine":         "pc",
			"firmware":        "",
			"firmware_secure": false,
		},
	})

	tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
	firmware, err := tpl.GetOS(vmk.Firmware)
	if err == nil || firmware != "" {
		t.Fatalf("firmware = %q err=%v, want attribute removed", firmware, err)
	}
	machine, _ := tpl.GetOS(vmk.Machine)
	if machine != "pc" {
		t.Fatalf("machine = %q, want pc", machine)
	}
	firmwareSecure, _ := tpl.GetOS(vmk.FirmwareSecure)
	if firmwareSecure != "NO" {
		t.Fatalf("firmware_secure = %q, want NO", firmwareSecure)
	}
}

func TestTemplateUpdateGraphics(t *testing.T) {
	prior := baseTemplateConfig()
	next := withChange(prior, "graphics", []interface{}{
		map[string]interface{}{
			"listen": "127.0.0.1",
			"type":   "SPICE",
		},
	})

	tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
	graphicType, _ := tpl.GetIOGraphic(vmk.GraphicType)
	listen, _ := tpl.GetIOGraphic(vmk.Listen)
	if graphicType != "SPICE" || listen != "127.0.0.1" {
		t.Fatalf("graphics = type %q listen %q, want SPICE / 127.0.0.1", graphicType, listen)
	}
}

func TestTemplateUpdateCPUModel(t *testing.T) {
	prior := baseTemplateConfig()
	next := withChange(prior, "cpumodel", []interface{}{
		map[string]interface{}{
			"model": "host-model",
		},
	})

	tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
	model, _ := tpl.GetCPUModel(vmk.Model)
	if model != "host-model" {
		t.Fatalf("cpumodel = %q, want host-model", model)
	}
}

func TestTemplateUpdateDisk(t *testing.T) {
	prior := baseTemplateConfig()
	next := withChange(prior, "disk", []interface{}{
		map[string]interface{}{
			"image_id":   10,
			"size":       40960,
			"dev_prefix": "vd",
		},
	})

	tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
	body := tpl.String()
	if strings.Count(body, "DISK=[") != 1 || !strings.Contains(body, `SIZE="40960"`) || !strings.Contains(body, `DEV_PREFIX="vd"`) {
		t.Fatalf("disk was not replaced:\n%s", body)
	}
}

func TestTemplateUpdateCapacity(t *testing.T) {
	prior := baseTemplateConfig()
	next := withChange(prior, "cpu", 2.0)
	next = withChange(next, "vcpu", 4)
	next = withChange(next, "memory", 2048)

	tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
	cpu, _ := tpl.GetCPU()
	vcpu, _ := tpl.GetVCPU()
	memory, _ := tpl.GetMemory()
	if cpu != 2 {
		t.Fatalf("cpu = %v, want 2", cpu)
	}
	if vcpu != 4 {
		t.Fatalf("vcpu = %d, want 4", vcpu)
	}
	if memory != 2048 {
		t.Fatalf("memory = %d, want 2048", memory)
	}
}

func TestTemplateUpdateContextFeaturesRaw(t *testing.T) {
	prior := baseTemplateConfig()
	next := withChange(prior, "context", map[string]interface{}{
		"NETWORK":      "YES",
		"SET_HOSTNAME": "$NAME",
	})
	next = withChange(next, "features", []interface{}{
		map[string]interface{}{
			"acpi": "NO",
			"apic": "YES",
		},
	})
	next = withChange(next, "raw", []interface{}{
		map[string]interface{}{
			"type": "kvm",
			"data": "<clock offset='localtime'/>",
		},
	})

	tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
	hostname, _ := tpl.GetCtx(vmk.Context("SET_HOSTNAME"))
	if hostname != "$NAME" {
		t.Fatalf("SET_HOSTNAME = %q, want $NAME", hostname)
	}
	acpi, _ := tpl.GetFeature(vmk.ACPI)
	apic, _ := tpl.GetFeature(vmk.APIC)
	if acpi != "NO" || apic != "YES" {
		t.Fatalf("features acpi=%q apic=%q, want NO / YES", acpi, apic)
	}
	raw, err := tpl.GetVector("RAW")
	if err != nil {
		t.Fatal(err)
	}
	data, _ := raw.GetStr("DATA")
	if data != "<clock offset='localtime'/>" {
		t.Fatalf("raw data = %q", data)
	}
}

func TestTemplateUpdateUserInputsAndPlacement(t *testing.T) {
	prior := baseTemplateConfig()
	next := withChange(prior, "user_inputs", map[string]interface{}{
		"CLUSTER_ID": "M|text|Cluster| |",
		"KLANT_NR":   "M|number|Klant| |",
	})
	next = withChange(next, "description", "updated")
	next = withChange(next, "sched_requirements", "CLUSTER_ID=1")
	next = withChange(next, "sched_ds_requirements", "NAME=images")

	tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
	inputs, err := tpl.GetVector("USER_INPUTS")
	if err != nil {
		t.Fatal(err)
	}
	klant, _ := inputs.GetStr("KLANT_NR")
	if klant != "M|number|Klant| |" {
		t.Fatalf("KLANT_NR = %q", klant)
	}
	description, _ := tpl.GetStr(string(vmk.Description))
	if description != "updated" {
		t.Fatalf("description = %q", description)
	}
	sched, _ := tpl.GetStr("SCHED_REQUIREMENTS")
	if sched != "CLUSTER_ID=1" {
		t.Fatalf("sched_requirements = %q", sched)
	}
	schedDS, _ := tpl.GetStr("SCHED_DS_REQUIREMENTS")
	if schedDS != "NAME=images" {
		t.Fatalf("sched_ds_requirements = %q", schedDS)
	}
}

func TestTemplateUpdateTagsAndSection(t *testing.T) {
	prior := baseTemplateConfig()
	next := withChange(prior, "tags", map[string]interface{}{
		"logo": "windows8.png",
	})
	next = withChange(next, "template_section", []interface{}{
		map[string]interface{}{
			"name": "input",
			"elements": map[string]interface{}{
				"TYPE": "mouse",
				"BUS":  "usb",
			},
		},
	})

	tpl, _ := applyCreatedTemplateUpdate(t, prior, next)
	logo, _ := tpl.GetStr("LOGO")
	if logo != "windows8.png" {
		t.Fatalf("LOGO = %q, want windows8.png", logo)
	}
	section, err := tpl.GetVector("INPUT")
	if err != nil {
		t.Fatal(err)
	}
	inputType, _ := section.GetStr("TYPE")
	if inputType != "mouse" {
		t.Fatalf("input TYPE = %q, want mouse", inputType)
	}
}

func TestTemplateVMGroupIsRecreated(t *testing.T) {
	vmgroup := resourceOpennebulaTemplate().Schema["vmgroup"]
	if vmgroup == nil || !vmgroup.ForceNew {
		t.Fatal("vmgroup changes must force a new template")
	}
}
