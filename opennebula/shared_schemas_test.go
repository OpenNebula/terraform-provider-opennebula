package opennebula

import (
	"strings"
	"testing"

	vm "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
)

// TestAddOS_FirmwareSecure_storesYES proves that firmware_secure=true must
// produce FIRMWARE_SECURE="YES" in the OpenNebula template sent to the API.
//
// OpenNebula's Template.cc:345 only recognises "YES" (case-insensitive) as
// boolean true. strconv.FormatBool(true) produces "true", which OpenNebula
// silently treats as false — disabling SecureBoot despite firmware_secure=true.
func TestAddOS_FirmwareSecure_storesYES(t *testing.T) {
	tpl := vm.NewTemplate()

	addOS(tpl, []interface{}{
		map[string]interface{}{
			"arch": "x86_64", "machine": "q35", "boot": "disk",
			"kernel": "", "kernel_ds": "", "initrd": "", "initrd_ds": "",
			"root": "", "kernel_cmd": "", "bootloader": "", "sd_disk_bus": "",
			"uuid": "", "firmware": "/usr/share/OVMF/OVMF_CODE.secboot.fd",
			"firmware_secure": true,
		},
	})

	tplStr := tpl.String()

	if !strings.Contains(tplStr, `FIRMWARE_SECURE="YES"`) {
		t.Errorf("template sent to OpenNebula API:\n%s\nwant FIRMWARE_SECURE=\"YES\", OpenNebula ignores any other value", tplStr)
	}
}
