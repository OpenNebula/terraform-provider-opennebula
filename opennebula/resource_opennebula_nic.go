package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

func resourceOpennebulaNIC() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaNICCreate,
		ReadContext:   resourceOpennebulaNICRead,
		Exists:        resourceOpennebulaNICExists,
		DeleteContext: resourceOpennebulaNICDelete,
		CustomizeDiff: resourceVMCustomizeDiff,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultVMTimeout),
			Update: schema.DefaultTimeout(defaultVMTimeout),
			Delete: schema.DefaultTimeout(defaultVMTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"vm_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the virtual machine",
			},
			"network_id": {
				Type:     schema.TypeInt,
				Required: true,
				ForceNew: true,
			},
			"ip": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"mac": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"model": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"virtio_queues": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"network": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"physical_device": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"security_groups": {
				Type:     schema.TypeList,
				Optional: true,
				Computed: true,
				ForceNew: true,
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
		},
	}
}

func resourceOpennebulaNICCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics
	var nicID int
	var err error

	vmID := d.Get("vm_id").(int)

	// avoid creation/update/deletion of multiple NIC/NICs and instances at the same time
	nicKey := &SubResourceKey{
		Type:    "virtual_machine",
		ID:      vmID,
		SubType: "nic",
	}
	config.mutex.Lock(nicKey)
	defer config.mutex.Unlock(nicKey)

	diskKey := &SubResourceKey{
		Type:    "virtual_machine",
		ID:      vmID,
		SubType: "disk",
	}
	config.mutex.Lock(diskKey)
	defer config.mutex.Unlock(diskKey)

	// detect if the NIC was already attached at
	// the virtual machine VM creation via the VM NIC description
	attached, nicID, diags := alreadyAttached(ctx, d, meta)
	if !attached {
		nicTpl := shared.NewNIC()
		networkID := d.Get("network_id").(int)
		nicTpl.Add(shared.NetworkID, strconv.Itoa(networkID))

		if v, ok := d.GetOk("ip"); ok {
			nicTpl.Add(shared.IP, v.(string))
		}
		if v, ok := d.GetOk("mac"); ok {
			nicTpl.Add(shared.MAC, v.(string))
		}
		if v, ok := d.GetOk("model"); ok {
			nicTpl.Add(shared.Model, v.(int))
		}
		if v, ok := d.GetOk("virtio_queues"); ok {
			nicTpl.Add("VIRTIO_QUEUES", v.(string))
		}
		if v, ok := d.GetOk("physical_device"); ok {
			nicTpl.Add("PHYSICAL_DEVICE", v.(string))
		}
		if v, ok := d.GetOk("security_groups"); ok {
			secGroups := ArrayToString(v.([]interface{}), ",")
			nicTpl.Add(shared.SecurityGroups, secGroups)
		}

		nicID, err = vmNICAttach(ctx, controller.VM(vmID), d.Timeout(schema.TimeoutCreate), nicTpl)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to attach NIC",
				Detail:   fmt.Sprintf("virtual machine NIC (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully attached virtual machine NIC\n")
	}

	d.SetId(fmt.Sprintf("%d", nicID))

	return resourceOpennebulaNICRead(ctx, d, meta)
}

func alreadyAttached(ctx context.Context, d *schema.ResourceData, meta interface{}) (bool, int, diag.Diagnostics) {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	vmID := d.Get("vm_id").(int)
	vm, err := controller.VM(vmID).Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("virtual machine NIC (ID: %s): %s", d.Id(), err),
		})
		return false, -1, diags
	}

	nics := vm.Template.GetNICs()
	for _, nc := range nics {

		networkID := d.Get("network_id").(int)
		if networkID > -1 {
			networkIDCfg, _ := nc.GetI(shared.NetworkID)
			if networkIDCfg != networkID {
				continue
			}
		}

		ip, ok := d.GetOk("ip")
		if ok {
			ipCfg, _ := nc.Get(shared.IP)
			if ipCfg != ip.(string) {
				continue
			}
		}

		mac, ok := d.GetOk("mac")
		if ok {
			macCfg, _ := nc.Get(shared.MAC)
			if macCfg != mac.(string) {
				continue
			}
		}

		model, ok := d.GetOk("model")
		if ok {
			modelCfg, _ := nc.Get(shared.Model)
			if modelCfg != model.(string) {
				continue
			}
		}
		virtioQueues, ok := d.GetOk("virtio_queues")
		if ok {
			virtioQueuesCfg, _ := nc.Get("VIRTIO_QUEUES")
			if virtioQueuesCfg != virtioQueues.(string) {
				continue
			}
		}
		physicalDevice, ok := d.GetOk("physical_device")
		if ok {
			physicalDeviceCfg, _ := nc.Get("PHYSICAL_DEVICE")
			if physicalDeviceCfg != physicalDevice.(string) {
				continue
			}
		}
		securityGroups, ok := d.GetOk("security_groups")
		if ok {
			securityGroupsCfg, _ := nc.Get(shared.SecurityGroups)
			if securityGroupsCfg != securityGroups.(string) {
				continue
			}
		}
		nicID, _ := nc.ID()
		return true, nicID, nil
	}
	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  "Failed to find the NIC",
		Detail:   fmt.Sprintf("virtual machine NIC (ID: %s): %s", d.Id(), err),
	})
	return false, -1, diags
}

func resourceOpennebulaNICRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	config := meta.(*Configuration)
	controller := config.Controller
	vmID := d.Get("vm_id").(int)

	vm, err := controller.VM(vmID).Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("virtual machine NIC (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	var nic *shared.NIC

	nics := vm.Template.GetNICs()
	for _, dk := range nics {
		nicID, _ := dk.Get(shared.NICID)
		if nicID == d.Id() {
			nic = &dk
			break
		}
	}

	if nic == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to find the NIC in the virtual machine NIC list",
			Detail:   fmt.Sprintf("virtual machine NIC (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	networkID, _ := nic.GetI(shared.NetworkID)
	ip, _ := nic.Get(shared.IP)
	mac, _ := nic.Get(shared.MAC)
	model, _ := nic.Get(shared.Model)
	virtioQueues, _ := nic.Get("VIRTIO_QUEUES")
	network, _ := nic.Get(shared.Network)
	physicalDevice, _ := nic.Get("PHYSICAL_DEVICE")
	securityGroupsArray, _ := nic.Get("SECURITY_GROUPS")

	sg := make([]int, 0)
	sgString := strings.Split(securityGroupsArray, ",")
	for _, s := range sgString {
		sgInt, _ := strconv.ParseInt(s, 10, 32)
		sg = append(sg, int(sgInt))
	}

	d.Set("network_id", networkID)
	d.Set("network", network)
	d.Set("ip", ip)
	d.Set("mac", mac)
	d.Set("model", model)
	d.Set("virtio_queues", virtioQueues)
	d.Set("physical_device", physicalDevice)
	d.Set("security_groups", sg)

	return nil
}

func resourceOpennebulaNICExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	config := meta.(*Configuration)
	controller := config.Controller
	vmID := d.Get("vm_id").(int)

	vmInfos, err := controller.VM(vmID).Info(false)
	if NoExists(err) {
		return false, err
	}

	for _, nic := range vmInfos.Template.GetNICs() {
		nicID, _ := nic.Get(shared.NICID)
		if nicID == d.Id() {
			return true, nil
		}
	}

	return false, err
}

func resourceOpennebulaNICDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	vmID := d.Get("vm_id").(int)
	vmc := controller.VM(vmID)

	nicID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse virtual machine NIC ID",
			Detail:   fmt.Sprintf("virtual machine NIC (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	// avoid creation/update/deletion of multiple NIC/NICs and instances at the same time
	nicKey := &SubResourceKey{
		Type:    "virtual_machine",
		ID:      vmID,
		SubType: "nic",
	}
	config.mutex.Lock(nicKey)
	defer config.mutex.Unlock(nicKey)

	diskKey := &SubResourceKey{
		Type:    "virtual_machine",
		ID:      vmID,
		SubType: "disk",
	}
	config.mutex.Lock(diskKey)
	defer config.mutex.Unlock(diskKey)

	err = vmNICDetach(ctx, vmc, d.Timeout(schema.TimeoutDelete), int(nicID))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to detach NIC",
			Detail:   fmt.Sprintf("virtual machine NIC (ID: %d): %s", nicID, err),
		})
		return diags
	}

	log.Printf("[INFO] Successfully detached virtual machine NIC\n")
	return nil
}
