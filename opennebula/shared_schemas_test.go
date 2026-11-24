package opennebula

import (
	"strings"
	"testing"

	vm "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
)

func testOSTemplateConfig(secure bool) []interface{} {
	return []interface{}{
		map[string]interface{}{
			"arch": "x86_64", "machine": "q35", "boot": "disk",
			"kernel": "", "kernel_ds": "", "initrd": "", "initrd_ds": "",
			"root": "", "kernel_cmd": "", "bootloader": "", "sd_disk_bus": "",
			"uuid": "", "firmware": "/usr/share/OVMF/OVMF_CODE.secboot.fd",
			"firmware_secure": secure,
		},
	}
}

// TestAddOS_FirmwareSecure_storesYES proves that firmware_secure=true must
// produce FIRMWARE_SECURE="YES" in the OpenNebula template sent to the API.
//
// OpenNebula's Template.cc:341 only recognises "YES" (case-insensitive) as
// boolean true. strconv.FormatBool(true) produces "true", which OpenNebula
// silently treats as false — disabling SecureBoot despite firmware_secure=true.
func TestAddOS_FirmwareSecure_storesYES(t *testing.T) {
	tpl := vm.NewTemplate()
	addOS(tpl, testOSTemplateConfig(true))

	tplStr := tpl.String()
	if !strings.Contains(tplStr, `FIRMWARE_SECURE="YES"`) {
		t.Errorf("template sent to OpenNebula API:\n%s\nwant FIRMWARE_SECURE=\"YES\", OpenNebula ignores any other value", tplStr)
	}
}

// TestAddOS_FirmwareSecure_storesNO pins the inverted branch: firmware_secure=false
// must serialise to "NO" — the branch most prone to logic inversion.
func TestAddOS_FirmwareSecure_storesNO(t *testing.T) {
	tpl := vm.NewTemplate()
	addOS(tpl, testOSTemplateConfig(false))

	tplStr := tpl.String()
	if !strings.Contains(tplStr, `FIRMWARE_SECURE="NO"`) {
		t.Errorf("template sent to OpenNebula API:\n%s\nwant FIRMWARE_SECURE=\"NO\"", tplStr)
	}
	if strings.Contains(tplStr, `FIRMWARE_SECURE="true"`) || strings.Contains(tplStr, `FIRMWARE_SECURE="false"`) {
		t.Errorf("template must never contain Go-bool forms, got:\n%s", tplStr)
	}
}

// TestFirmwareSecure_roundTrip verifies the flatten side: what addOS writes,
// flattenTemplate must read back as the same bool. Covers YES, NO and
// case-insensitivity of the parser.
func TestFirmwareSecure_roundTrip(t *testing.T) {
	cases := []struct {
		stored string
		want   bool
	}{
		{"YES", true},
		{"yes", true},
		{"Yes", true},
		{"NO", false},
		{"no", false},
	}
	for _, tc := range cases {
		tpl := vm.NewTemplate()
		addOS(tpl, testOSTemplateConfig(tc.want))

		// Rebuild a template with the raw stored value, as OpenNebula would return it.
		fresh := vm.NewTemplate()
		osVec := fresh.AddVector("OS")
		osVec.AddPair("ARCH", "x86_64")
		osVec.AddPair("MACHINE", "q35")
		osVec.AddPair("BOOT", "disk")
		osVec.AddPair("FIRMWARE", "/usr/share/OVMF/OVMF_CODE.secboot.fd")
		osVec.AddPair("FIRMWARE_SECURE", tc.stored)

		got, err := firmwareSecureFromTemplate(fresh)
		if err != nil {
			t.Fatalf("stored %q: unexpected error: %v", tc.stored, err)
		}
		if got != tc.want {
			t.Errorf("stored %q: got firmware_secure=%v, want %v", tc.stored, got, tc.want)
		}
	}
}

// TestFirmwareSecure_absentPair pins the zero-value path: when the OS block
// exists but carries no FIRMWARE_SECURE pair, flatten must report false without
// error.
func TestFirmwareSecure_absentPair(t *testing.T) {
	fresh := vm.NewTemplate()
	osVec := fresh.AddVector("OS")
	osVec.AddPair("ARCH", "x86_64")
	osVec.AddPair("BOOT", "disk")

	got, err := firmwareSecureFromTemplate(fresh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != false {
		t.Errorf("absent FIRMWARE_SECURE: got %v, want false", got)
	}
}

// TestUpdateVMTemplateVec_FirmwareSecure guards the third write path: the
// generic vector-update used on every VM OS change must serialise bools as
// YES/NO, not Go-bool "true"/"false". Regression test for the update path
// bypassing addOS.
func TestUpdateVMTemplateVec_FirmwareSecure(t *testing.T) {
	tpl := vm.NewTemplate()
	osVec := tpl.AddVector("OS")
	osVec.AddPair("ARCH", "x86_64")
	osVec.AddPair("BOOT", "disk")
	osVec.AddPair("FIRMWARE_SECURE", "NO")

	applied := map[string]interface{}{"arch": "x86_64", "boot": "disk", "firmware_secure": false}
	newCfg := map[string]interface{}{"arch": "x86_64", "boot": "disk", "firmware_secure": true}

	if err := updateVMTemplateVec(tpl, "OS", applied, newCfg); err != nil {
		t.Fatalf("updateVMTemplateVec: %v", err)
	}

	tplStr := tpl.String()
	if !strings.Contains(tplStr, `FIRMWARE_SECURE="YES"`) {
		t.Errorf("OS update path must write YES, template now:\n%s", tplStr)
	}
	if strings.Contains(tplStr, `FIRMWARE_SECURE="true"`) {
		t.Errorf("OS update path leaked Go-bool form, template now:\n%s", tplStr)
	}
	if strings.Count(tplStr, "FIRMWARE_SECURE") != 1 {
		t.Errorf("expected exactly one FIRMWARE_SECURE pair, template now:\n%s", tplStr)
	}
}
