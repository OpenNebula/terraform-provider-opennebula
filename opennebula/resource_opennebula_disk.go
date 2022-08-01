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

func resourceOpennebulaDisk() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaDiskCreate,
		ReadContext:   resourceOpennebulaDiskRead,
		Exists:        resourceOpennebulaDiskExists,
		UpdateContext: resourceOpennebulaDiskUpdate,
		DeleteContext: resourceOpennebulaDiskDelete,
		CustomizeDiff: resourceVMCustomizeDiff,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultVMTimeout),
			Update: schema.DefaultTimeout(defaultVMTimeout),
			Delete: schema.DefaultTimeout(defaultVMTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpennebulaDiskImportState,
		},

		Schema: map[string]*schema.Schema{
			"vm_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the virtual machine",
			},
			"image_id": {
				Type:        schema.TypeInt,
				Default:     -1,
				Optional:    true,
				ForceNew:    true,
				Description: "Image Id  of the image to attach to the VM. Defaults to -1: no image attached.",
			},
			"size": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
			},
			"target": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"driver": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"volatile_type": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Computed:    true,
				Description: "Type of the volatile disk: swap or fs.",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					validtypes := []string{"swap", "fs"}
					value := v.(string)

					if inArray(value, validtypes) < 0 {
						errors = append(errors, fmt.Errorf("Type %q must be one of: %s", k, strings.Join(validtypes, ",")))
					}

					return
				},
			},
			"volatile_format": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				ForceNew:    true,
				Description: "Format of the volatile disk: raw or qcow2.",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					validtypes := []string{"raw", "qcow2"}
					value := v.(string)

					if inArray(value, validtypes) < 0 {
						errors = append(errors, fmt.Errorf("Format %q must be one of: %s", k, strings.Join(validtypes, ",")))
					}

					return
				},
			},
		},
	}
}

func resourceOpennebulaDiskCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics
	var diskID int
	var err error

	vmID := d.Get("vm_id").(int)

	// avoid creation/update/deletion of multiple disk/NICs and instances at the same time
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

	// build template disk then attach
	diskTpl := shared.NewDisk()
	imageID := d.Get("image_id").(int)
	if imageID >= 0 {
		diskTpl.Add(shared.ImageID, strconv.Itoa(imageID))
	}

	if v, ok := d.GetOk("target"); ok {
		diskTpl.Add(shared.TargetDisk, v.(string))
	}
	if v, ok := d.GetOk("driver"); ok {
		diskTpl.Add(shared.Driver, v.(string))
	}
	if v, ok := d.GetOk("size"); ok {
		diskTpl.Add(shared.Size, v.(int))
	}
	if v, ok := d.GetOk("volatile_type"); ok {
		diskTpl.Add("TYPE", v.(string))
	}
	if v, ok := d.GetOk("volatile_format"); ok {
		diskTpl.Add("FORMAT", v.(string))
	}

	diskID, err = vmDiskAttach(ctx, controller.VM(vmID), d.Timeout(schema.TimeoutCreate), diskTpl)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to attach disk",
			Detail:   fmt.Sprintf("virtual machine disk (ID: %s): %s", d.Id(), err),
		})
		return diags
	}
	d.SetId(fmt.Sprintf("%d", diskID))

	log.Printf("[INFO] Successfully attached virtual machine disk\n")

	return resourceOpennebulaDiskRead(ctx, d, meta)
}

func resourceOpennebulaDiskRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	config := meta.(*Configuration)
	controller := config.Controller
	vmID := d.Get("vm_id").(int)

	vm, err := controller.VM(vmID).Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("virtual machine disk (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	var disk *shared.Disk

	disks := vm.Template.GetDisks()
	for _, dk := range disks {
		diskID, _ := dk.Get(shared.DiskID)
		if diskID == d.Id() {
			disk = &dk
			break
		}
	}

	log.Printf("[INFO] Read data from disk %s/n", d.Id())

	if disk == nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to find the disk in the virtual machine disk list",
			Detail:   fmt.Sprintf("virtual machine disk (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	size, _ := disk.GetI(shared.Size)
	if size == -1 {
		size = 0
	}
	driver, _ := disk.Get(shared.Driver)
	target, _ := disk.Get(shared.TargetDisk)
	imageID, _ := disk.GetI(shared.ImageID)
	volatileType, _ := disk.Get("TYPE")
	volatileFormat, _ := disk.Get("FORMAT")

	d.Set("image_id", imageID)
	d.Set("size", size)
	d.Set("target", target)
	d.Set("driver", driver)
	d.Set("volatile_type", volatileType)
	d.Set("volatile_format", volatileFormat)

	return nil
}

func resourceOpennebulaDiskExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	var diskID int

	vmID := d.Get("vm_id").(int)
	vm, err := controller.VM(vmID).Info(false)
	if err != nil {
		return false, fmt.Errorf("Failed to retrieve virtual machine (ID: %s) information: %s", d.Id(), err)
	}

	disks := vm.Template.GetDisks()
	for _, dk := range disks {

		imageID := d.Get("image_id").(int)
		if imageID > -1 {
			imageIDCfg, _ := dk.GetI(shared.ImageID)
			if imageIDCfg != imageID {
				continue
			}
		}

		driver, ok := d.GetOk("driver")
		if ok {
			driverCfg, _ := dk.Get(shared.Driver)
			if driverCfg != driver.(string) {
				continue
			}
		}

		target, ok := d.GetOk("target")
		if ok {
			targetCfg, _ := dk.Get(shared.TargetDisk)
			if targetCfg != target.(string) {
				continue
			}
		}

		volatileType, ok := d.GetOk("volatile_type")
		if ok {
			volatileTypeCfg, _ := dk.Get("TYPE")
			if volatileTypeCfg != volatileType.(string) {
				continue
			}
		}

		volatileFormat, ok := d.GetOk("volatile_format")
		if ok {
			volatileFormatCfg, _ := dk.Get("FORMAT")
			if volatileFormatCfg != volatileFormat.(string) {
				continue
			}
		}

		diskID, _ = dk.ID()
		resourceDiskID, err := strconv.ParseInt(d.Id(), 10, 0)
		if err != nil {
			return false, fmt.Errorf("Failed to parse virtual machine (ID: %s) disk ID: %s", d.Id(), err)
		}

		if diskID == int(resourceDiskID) {
			return true, nil
		}

	}

	return false, nil
}

func resourceOpennebulaDiskUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	vmID := d.Get("vm_id").(int)
	vmc := controller.VM(vmID)

	// size is the only possible in-place update
	if d.HasChange("size") {

		size := d.Get("size").(int)

		// avoid update of multiple disks and instances at the same time
		diskKey := &SubResourceKey{
			Type:    "virtual_machine",
			ID:      vmID,
			SubType: "disk",
		}
		config.mutex.Lock(diskKey)
		defer config.mutex.Unlock(diskKey)

		diskID, err := strconv.ParseInt(d.Id(), 10, 0)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to parse virtual machine disk ID",
				Detail:   fmt.Sprintf("virtual machine disk (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		err = vmDiskResize(ctx, vmc, d.Timeout(schema.TimeoutUpdate), int(diskID), size)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to resize virtual machine disk",
				Detail:   fmt.Sprintf("virtual machine disk (ID: %d): %s", diskID, err),
			})
			return diags
		}
	}

	return resourceOpennebulaDiskRead(ctx, d, meta)
}

func resourceOpennebulaDiskDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	vmID := d.Get("vm_id").(int)
	vmc := controller.VM(vmID)

	diskID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse virtual machine disk ID",
			Detail:   fmt.Sprintf("virtual machine disk (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	// avoid creation/update/deletion of multiple disk/NICs and instances at the same time
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

	err = vmDiskDetach(ctx, vmc, d.Timeout(schema.TimeoutDelete), int(diskID))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to detach disk",
			Detail:   fmt.Sprintf("virtual machine disk (ID: %d): %s", diskID, err),
		})
		return diags
	}

	log.Printf("[INFO] Successfully detached virtual machine disk\n")
	return nil
}

func resourceOpennebulaDiskImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {

	fullID := d.Id()
	parts := strings.Split(fullID, ":")

	if len(parts) < 2 {
		return nil, fmt.Errorf("Invalid ID format. Expected: vm_id:disk_id")
	}

	vmID, err := strconv.ParseInt(parts[0], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse virtual machine ID: %s", err)
	}

	_, err = strconv.ParseInt(parts[1], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse vm disk ID: %s", err)
	}

	d.SetId(parts[1])
	d.Set("vm_id", vmID)

	return []*schema.ResourceData{d}, nil
}
