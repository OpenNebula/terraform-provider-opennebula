package opennebula

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/hashcode"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	errs "github.com/OpenNebula/one/src/oca/go/src/goca/errors"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
)

type vmTemplate struct {
	CPU                float64                `xml:"CPU,omitempty"`
	VCPU               int                    `xml:"VCPU,omitempty"`
	Memory             int                    `xml:"MEMORY,omitempty"`
	NICs               []vmNIC                `xml:"NIC,omitempty"`
	NICAliases         []vm.NicAlias          `xml:"NIC_ALIAS"`
	Context            stringMap              `xml:"CONTEXT,omitempty"`
	Disks              []vmDisk               `xml:"DISK,omitempty"`
	Graphics           []vmGraphics           `xml:"GRAPHICS,omitempty"`
	OS                 []vm.OS                `xml:"OS,omitempty"`
	Snapshots          []vm.Snapshot          `xml:"SNAPSHOT"`
	SecurityGroupRules []vm.SecurityGroupRule `xml:"SECURITY_GROUP_RULE"`
}

type vmNIC struct {
	ID              int    `xml:"NIC_ID,omitempty"`
	IP              string `xml:"IP,omitempty"`
	Model           string `xml:"MODEL,omitempty"`
	MAC             string `xml:"MAC,omitempty"`
	Network_ID      int    `xml:"NETWORK_ID"`
	PhyDev          string `xml:"PHYDEV"`
	Network         string `xml:"NETWORK"`
	Security_Groups string `xml:"SECURITY_GROUPS,omitempty"`
}

type vmDisk struct {
	ID       string `xml:"DISK_ID,omitempty"`
	Image_ID int    `xml:"IMAGE_ID"`
	Image    string `xml:"IMAGE"`
	Size     int    `xml:"SIZE,omitempty"`
	Target   string `xml:"TARGET,omitempty"`
	Driver   string `xml:"DRIVER,omitempty"`
}

type vmGraphics struct {
	Keymap string `xml:"KEYMAP,omitempty"`
	Listen string `xml:"LISTEN,omitempty"`
	Port   string `xml:"PORT"`
	Type   string `xml:"TYPE,omitempty"`
}

//This type and the MarshalXML functions are needed to handle converting the CONTEXT map to xml and back
//From: https://stackoverflow.com/questions/30928770/marshall-map-to-xml-in-go/33110881
type stringMap map[string]string
type xmlMapEntry struct {
	XMLName xml.Name
	Value   string `xml:",chardata"`
}

// MarshalXML marshals the map to XML, with each key in the map being a
// tag and it's corresponding value being it's contents.
func (m stringMap) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	if len(m) == 0 {
		return nil
	}

	err := e.EncodeToken(start)
	if err != nil {
		return err
	}

	for k, v := range m {
		e.Encode(xmlMapEntry{XMLName: xml.Name{Local: k}, Value: v})
	}

	return e.EncodeToken(start.End())
}

// UnmarshalXML unmarshals the XML into a map of string to strings,
// creating a key in the map for each tag and setting it's value to the
// tags contents.
//
// The fact this function is on the pointer of Map is important, so that
// if m is nil it can be initialized, which is often the case if m is
// nested in another xml structurel. This is also why the first thing done
// on the first line is initialize it.
func (m *stringMap) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	*m = stringMap{}
	for {
		var e xmlMapEntry

		err := d.Decode(&e)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		(*m)[e.XMLName.Local] = e.Value
	}
	return nil
}

func resourceOpennebulaVirtualMachine() *schema.Resource {
	return &schema.Resource{
		Create:        resourceOpennebulaVirtualMachineCreate,
		Read:          resourceOpennebulaVirtualMachineRead,
		Exists:        resourceOpennebulaVirtualMachineExists,
		Update:        resourceOpennebulaVirtualMachineUpdate,
		Delete:        resourceOpennebulaVirtualMachineDelete,
		CustomizeDiff: resourceVMCustomizeDiff,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the VM. If empty, defaults to 'templatename-<vmid>'",
			},
			"instance": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Final name of the VM instance",
			},
			"template_id": {
				Type:        schema.TypeInt,
				Optional:    true,
				ForceNew:    true,
				Description: "Id of the VM template to use",
			},
			"pending": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Pending state of the VM during its creation, by default it is set to false",
			},
			"permissions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Permissions for the template (in Unix format, owner-group-other, use-manage-admin)",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)

					if len(value) != 3 {
						errors = append(errors, fmt.Errorf("%q has specify 3 permission sets: owner-group-other", k))
					}

					all := true
					for _, c := range strings.Split(value, "") {
						if c < "0" || c > "7" {
							all = false
						}
					}
					if !all {
						errors = append(errors, fmt.Errorf("Each character in %q should specify a Unix-like permission set with a number from 0 to 7", k))
					}

					return
				},
			},

			"uid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the user that will own the VM",
			},
			"gid": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "ID of the group that will own the VM",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the VM",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the VM",
			},
			"state": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Current state of the VM",
			},
			"lcmstate": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Current LCM state of the VM",
			},
			"cpu": {
				Type:        schema.TypeFloat,
				Optional:    true,
				Description: "Amount of CPU quota assigned to the virtual machine",
			},
			"vcpu": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Number of virtual CPUs assigned to the virtual machine",
			},
			"memory": {
				Type:        schema.TypeInt,
				Optional:    true,
				Description: "Amount of memory (RAM) in MB assigned to the virtual machine",
			},
			"context": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Context variables",
			},
			"disk": {
				Type:     schema.TypeSet,
				Optional: true,
				//Computed:    true,
				MinItems:    0,
				MaxItems:    8,
				Description: "Definition of disks assigned to the Virtual Machine",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"image_id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"size": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"target": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"driver": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"graphics": {
				Type:     schema.TypeSet,
				Optional: true,
				MinItems: 0,
				MaxItems: 1,
				//Computed:    true,
				Description: "Definition of graphics adapter assigned to the Virtual Machine",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"listen": {
							Type:     schema.TypeString,
							Required: true,
						},
						"port": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"keymap": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "en-us",
						},
					},
				},
			},
			"nic": {
				Type:     schema.TypeSet,
				Optional: true,
				//Computed:    true,
				MinItems:    0,
				MaxItems:    8,
				Description: "Definition of network adapter(s) assigned to the Virtual Machine",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mac": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"model": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"network_id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"network": {
							Type:     schema.TypeString,
							Computed: true,
						},
						"physical_device": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"security_groups": {
							Type:     schema.TypeList,
							Optional: true,
							Elem: &schema.Schema{
								Type: schema.TypeInt,
							},
						},
						"nic_id": {
							Type:     schema.TypeInt,
							Computed: true,
						},
					},
				},
				Set: resourceVMNicHash,
			},
			"os": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				//Computed:    true,
				Description: "Definition of OS boot and type for the Virtual Machine",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"arch": {
							Type:     schema.TypeString,
							Required: true,
						},
						"boot": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"ip": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Primary IP address assigned by OpenNebula",
			},
			"group": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"gid"},
				Description:   "Name of the Group that onws the VM, If empty, it uses caller group",
			},
		},
	}
}

func getVirtualMachineController(d *schema.ResourceData, meta interface{}, args ...int) (*goca.VMController, error) {
	controller := meta.(*goca.Controller)
	var vmc *goca.VMController

	// Try to find the VM by ID, if specified
	if d.Id() != "" {
		id, err := strconv.ParseUint(d.Id(), 10, 64)
		if err != nil {
			return nil, err
		}
		vmc = controller.VM(int(id))
	}

	// Otherwise, try to find the VM by name as the de facto compound primary key
	if d.Id() == "" {
		gid, err := controller.VMs().ByName(d.Get("name").(string), args...)
		if err != nil {
			return nil, err
		}
		vmc = controller.VM(gid)
	}

	return vmc, nil
}

func changeVmGroup(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)
	var gid int

	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		return err
	}

	if d.Get("group") != "" {
		gid, err = controller.Groups().ByName(d.Get("group").(string))
		if err != nil {
			return err
		}
	} else {
		gid = d.Get("gid").(int)
	}

	err = vmc.Chown(-1, gid)
	if err != nil {
		return err
	}

	return nil
}

func resourceOpennebulaVirtualMachineCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	//Call one.template.instantiate only if template_id is defined
	//otherwise use one.vm.allocate
	var err error
	var vmID int

	if v, ok := d.GetOk("template_id"); ok {
		// if template id is set, instantiate a VM from this template
		tc := controller.Template(v.(int))

		// customize template except for memory and cpu.
		vmxml, xmlerr := generateVmXML(d)
		if xmlerr != nil {
			return xmlerr
		}

		// Instantiate template without creating a persistent copy of the template
		// Note that the new VM is not pending
		vmID, err = tc.Instantiate(d.Get("name").(string), d.Get("pending").(bool), vmxml, false)
	} else {
		if _, ok := d.GetOk("cpu"); !ok {
			return fmt.Errorf("cpu is mandatory as template_id is not used")
		}
		if _, ok := d.GetOk("memory"); !ok {
			return fmt.Errorf("memory is mandatory as template_id is not used")
		}

		vmxml, xmlerr := generateVmXML(d)
		if xmlerr != nil {
			return xmlerr
		}

		// Create VM not in pending state
		vmID, err = controller.VMs().Create(vmxml, d.Get("pending").(bool))
	}

	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", vmID))
	vmc := controller.VM(vmID)

	_, err = waitForVmState(d, meta, "running")
	if err != nil {
		return fmt.Errorf(
			"Error waiting for virtual machine (%s) to be in state RUNNING: %s", d.Id(), err)
	}

	// Rename the VM with its real name
	if d.Get("name") != nil {
		err := vmc.Rename(d.Get("name").(string))
		if err != nil {
			return err
		}
	}

	//Set the permissions on the VM if it was defined, otherwise use the UMASK in OpenNebula
	if perms, ok := d.GetOk("permissions"); ok {
		err = vmc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			log.Printf("[ERROR] template permissions change failed, error: %s", err)
			return err
		}
	}

	if d.Get("group") != "" || d.Get("gid") != "" {
		err = changeVmGroup(d, meta)
		if err != nil {
			return err
		}
	}

	return resourceOpennebulaVirtualMachineRead(d, meta)
}

func resourceOpennebulaVirtualMachineRead(d *schema.ResourceData, meta interface{}) error {
	vmc, err := getVirtualMachineController(d, meta, -2, -1, -1)
	if err != nil {
		switch err.(type) {
		case *errs.ClientError:
			clientErr, _ := err.(*errs.ClientError)
			if clientErr.Code == errs.ClientRespHTTP {
				response := clientErr.GetHTTPResponse()
				if response.StatusCode == http.StatusNotFound {
					log.Printf("[WARN] Removing virtual machine %s from state because it no longer exists in", d.Get("name"))
					d.SetId("")
					return nil
				}
			}
			return err
		default:
			return err
		}
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vm, err := vmc.Info(false)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", vm.ID))
	d.Set("instance", vm.Name)
	d.Set("name", vm.Name)
	d.Set("uid", vm.UID)
	d.Set("gid", vm.GID)
	d.Set("uname", vm.UName)
	d.Set("gname", vm.GName)
	d.Set("state", vm.StateRaw)
	d.Set("lcmstate", vm.LCMStateRaw)
	//TODO fix this:
	err = d.Set("permissions", permissionsUnixString(vm.Permissions))
	if err != nil {
		return err
	}

	//Pull in NIC config from OpenNebula into schema
	if vm.Template.NICs != nil {
		err = flattenNics(d, vm.Template.NICs)
		if err != nil {
			return err
		}
		err = d.Set("ip", &vm.Template.NICs[0].IP)
		if err != nil {
			return err
		}
	}

	if vm.Template.Disks != nil {
		err = flattenDisks(d, vm.Template.Disks)
		if err != nil {
			return err
		}
	}

	if vm.Template.OS != nil {
		err = flattenOs(d, *vm.Template.OS)
		if err != nil {
			return err
		}
	}

	if vm.Template.Graphics != nil {
		err = flattenGraphics(d, *vm.Template.Graphics)
		if err != nil {
			return err
		}
	}
	return nil
}

func flattenGraphics(d *schema.ResourceData, graph vm.Graphics) error {

	graphmap := make([]map[string]interface{}, 0)
	listen := graph.Listen
	port := graph.Port
	t := graph.Type

	graphmap = append(graphmap, map[string]interface{}{
		"listen": listen,
		"port":   port,
		"type":   t,
		"keymap": d.Get("keymap"),
	})

	return d.Set("graphics", graphmap)
}

func flattenOs(d *schema.ResourceData, os vm.OS) error {

	osmap := make([]map[string]interface{}, 0)

	osmap = append(osmap, map[string]interface{}{
		"arch": os.Arch,
		"boot": os.Boot,
	})

	return d.Set("os", osmap)
}

func flattenDisks(d *schema.ResourceData, slice []vm.Disk) error {

	diskmap := make([]map[string]interface{}, 0)

	for _, disk := range slice {
		size := disk.Size
		driver := disk.Driver

		target, _ := disk.Dynamic.GetContentByName("TARGET")
		imageIdStr, _ := disk.Dynamic.GetContentByName("IMAGE_ID")
		imageId, err := strconv.ParseInt(imageIdStr, 10, 64)
		if err != nil {
			return err
		}

		diskmap = append(diskmap, map[string]interface{}{
			"image_id": imageId,
			"size":     size,
			"target":   target,
			"driver":   driver,
		})
	}

	return d.Set("disk", diskmap)
}

func flattenNics(d *schema.ResourceData, slice []vm.Nic) error {

	nicmap := make([]map[string]interface{}, 0)

	for _, nic := range slice {
		sg := make([]int, 0)
		ip := nic.IP
		mac := nic.MAC
		physicalDevice := nic.PhyDev
		network := nic.Network
		nicId := nic.ID

		model, _ := nic.Dynamic.GetContentByName("MODEL")
		networkIdStr, _ := nic.Dynamic.GetContentByName("NETWORK_ID")
		networkId, err := strconv.ParseInt(networkIdStr, 10, 64)
		if err != nil {
			return err
		}
		securityGroupsArray, _ := nic.Dynamic.GetContentByName("SECURITY_GROUPS")

		sgString := strings.Split(securityGroupsArray, ",")
		for _, s := range sgString {
			sgInt, _ := strconv.ParseInt(s, 10, 32)
			sg = append(sg, int(sgInt))
		}

		nicmap = append(nicmap, map[string]interface{}{
			"ip":              ip,
			"mac":             mac,
			"network_id":      int(networkId),
			"physical_device": physicalDevice,
			"network":         network,
			"nic_id":          nicId,
			"model":           model,
			"security_groups": sg,
		})
	}

	return d.Set("nic", nicmap)
}

func resourceOpennebulaVirtualMachineExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceOpennebulaVirtualMachineRead(d, meta)
	// a terminated VM is in state 6 (DONE)
	if err != nil || d.Id() == "" || d.Get("state").(int) == 6 {
		return false, err
	}

	return true, nil
}

func resourceOpennebulaVirtualMachineUpdate(d *schema.ResourceData, meta interface{}) error {

	// Enable partial state mode
	d.Partial(true)

	//Get VM
	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		return err
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vm, err := vmc.Info(false)
	if err != nil {
		return err
	}

	if d.HasChange("name") {
		err := vmc.Rename(d.Get("name").(string))
		if err != nil {
			return err
		}
		// TODO: fix it after 5.10 release
		// Force the "decrypt" bool to false to keep ONE 5.8 behavior
		vm, err := vmc.Info(false)
		d.SetPartial("name")
		log.Printf("[INFO] Successfully updated name (%s) for VM ID %x\n", vm.Name, vm.ID)
	}

	if d.HasChange("permissions") && d.Get("permissions") != "" {
		if perms, ok := d.GetOk("permissions"); ok {
			err = vmc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				return err
			}
		}
		d.SetPartial("permissions")
		log.Printf("[INFO] Successfully updated Permissions VM %s\n", vm.Name)
	}

	if d.HasChange("group") {
		err := changeVmGroup(d, meta)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated group for VM %s\n", vm.Name)
	}

	// We succeeded, disable partial mode. This causes Terraform to save
	// save all fields again.
	d.Partial(false)

	return nil
}

func resourceOpennebulaVirtualMachineDelete(d *schema.ResourceData, meta interface{}) error {
	err := resourceOpennebulaVirtualMachineRead(d, meta)
	if err != nil || d.Id() == "" {
		return err
	}

	//Get VM
	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		return err
	}

	if err = vmc.TerminateHard(); err != nil {
		return err
	}

	_, err = waitForVmState(d, meta, "done")
	if err != nil {
		vm, _ := vmc.Info(false)

		vmState, vmLcmState, _ := vm.State()
		if vmLcmState.String() == "EPILOG_FAILURE" {
			if err = vmc.TerminateHard(); err != nil {
				return err
			}
		} else {
			return fmt.Errorf(
				"Error waiting for virtual machine (%s) to be in state DONE: %s (state: %v, lcmState: %v)", d.Id(), err, vmState, vmLcmState)
		}
	}

	log.Printf("[INFO] Successfully terminated VM\n")
	return nil
}

func waitForVmState(d *schema.ResourceData, meta interface{}, state string) (interface{}, error) {
	var vm *vm.VM
	var err error
	//Get VM controller
	vmc, err := getVirtualMachineController(d, meta)
	if err != nil {
		return vm, err
	}

	log.Printf("Waiting for VM (%s) to be in state Done", d.Id())

	stateConf := &resource.StateChangeConf{
		Pending: []string{"anythingelse"}, Target: []string{state},
		Refresh: func() (interface{}, string, error) {
			log.Println("Refreshing VM state...")
			if d.Id() != "" {
				//Get VM controller
				vmc, err = getVirtualMachineController(d, meta)
				if err != nil {
					return vm, "", fmt.Errorf("Could not find VM by ID %s", d.Id())
				}
			}
			// TODO: fix it after 5.10 release
			// Force the "decrypt" bool to false to keep ONE 5.8 behavior
			vm, err = vmc.Info(false)
			if err != nil {
				if strings.Contains(err.Error(), "Error getting") {
					return vm, "notfound", nil
				}
				return vm, "", err
			}
			vmState, vmLcmState, err := vm.State()
			if err != nil {
				if strings.Contains(err.Error(), "Error getting") {
					return vm, "notfound", nil
				}
				return vm, "", err
			}
			log.Printf("VM %v is currently in state %v and in LCM state %v", vm.ID, vmState, vmLcmState)
			if vmState == 3 && vmLcmState == 3 {
				return vm, "running", nil
			} else if vmState == 6 {
				return vm, "done", nil
			} else if vmState == 3 && vmLcmState == 36 {
				return vm, "boot_failure", fmt.Errorf("VM ID %s entered fail state, error message: %s", d.Id(), vm.UserTemplate.Error)
			} else if vmState == 3 && vmLcmState == 39 {
				return vm, "prolog_failure", fmt.Errorf("VM ID %s entered fail state, error message: %s", d.Id(), vm.UserTemplate.Error)
			} else if vmState == 3 && vmLcmState == 40 {
				return vm, "epilog_failure", fmt.Errorf("VM ID %s entered fail state, error message: %s", d.Id(), vm.UserTemplate.Error)
			} else {
				return vm, "anythingelse", nil
			}
		},
		Timeout:    3 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForState()
}

func generateVmXML(d *schema.ResourceData) (string, error) {

	//Generate CONTEXT definition
	//context := d.Get("context").(*schema.Set).List()
	context := d.Get("context").(map[string]interface{})
	log.Printf("Number of CONTEXT vars: %d", len(context))
	log.Printf("CONTEXT Map: %s", context)

	vmcontext := make(stringMap)
	for key, value := range context {
		//contextvar = v.(map[string]interface{})
		vmcontext[key] = fmt.Sprint(value)
	}

	//Generate NIC definition
	nics := d.Get("nic").(*schema.Set).List()
	log.Printf("Number of NICs: %d", len(nics))
	vmnics := make([]vmNIC, len(nics))
	for i := 0; i < len(nics); i++ {
		nicconfig := nics[i].(map[string]interface{})
		nicip := nicconfig["ip"].(string)
		nicmac := nicconfig["mac"].(string)
		nicmodel := nicconfig["model"].(string)
		nicphydev := nicconfig["physical_device"].(string)
		nicnetworkid := nicconfig["network_id"].(int)
		nicsecgroups := ArrayToString(nicconfig["security_groups"].([]interface{}), ",")

		vmnic := vmNIC{
			IP:              nicip,
			MAC:             nicmac,
			Model:           nicmodel,
			PhyDev:          nicphydev,
			Network_ID:      nicnetworkid,
			Security_Groups: nicsecgroups,
		}
		vmnics[i] = vmnic
	}

	//Generate DISK definition
	disks := d.Get("disk").(*schema.Set).List()
	log.Printf("Number of disks: %d", len(disks))
	vmdisks := make([]vmDisk, len(disks))
	for i := 0; i < len(disks); i++ {
		diskconfig := disks[i].(map[string]interface{})
		diskimageid := diskconfig["image_id"].(int)
		disksize := diskconfig["size"].(int)
		disktarget := diskconfig["target"].(string)
		diskdriver := diskconfig["driver"].(string)

		vmdisk := vmDisk{
			Image_ID: diskimageid,
			Size:     disksize,
			Target:   disktarget,
			Driver:   diskdriver,
		}
		vmdisks[i] = vmdisk
	}

	//Generate GRAPHICS definition
	graphics := d.Get("graphics").(*schema.Set).List()
	vmgraphics := make([]vmGraphics, len(graphics))
	for i := 0; i < len(graphics); i++ {
		graphicsconfig := graphics[i].(map[string]interface{})
		gfxlisten := graphicsconfig["listen"].(string)
		gfxtype := graphicsconfig["type"].(string)
		gfxport := graphicsconfig["port"].(string)
		gfxkeymap := graphicsconfig["keymap"].(string)
		vmgraphic := vmGraphics{
			Listen: gfxlisten,
			Port:   gfxport,
			Type:   gfxtype,
			Keymap: gfxkeymap,
		}
		vmgraphics[i] = vmgraphic
	}

	//Generate OS definition
	os := d.Get("os").(*schema.Set).List()
	vmos := make([]vm.OS, len(os))
	for i := 0; i < len(os); i++ {
		osconfig := os[i].(map[string]interface{})
		osarch := osconfig["arch"].(string)
		osboot := osconfig["boot"].(string)
		vmo := vm.OS{
			Arch: osarch,
			Boot: osboot,
		}
		vmos[i] = vmo
	}

	//Pull all the bits together into the main VM template
	var vmvcpu interface{}
	var vmcpu interface{}
	var vmmemory interface{}
	var vmtpl *vmTemplate
	var ok bool
	if vmcpu, ok = d.GetOk("cpu"); ok {
		if vmmemory, ok = d.GetOk("memory"); ok {
			if vmvcpu, ok = d.GetOk("vcpu"); ok {
				vmtpl = &vmTemplate{
					VCPU:     vmvcpu.(int),
					CPU:      vmcpu.(float64),
					Memory:   vmmemory.(int),
					Context:  vmcontext,
					NICs:     vmnics,
					Disks:    vmdisks,
					Graphics: vmgraphics,
					OS:       vmos,
				}
			}
			vmtpl = &vmTemplate{
				CPU:      vmcpu.(float64),
				Memory:   vmmemory.(int),
				Context:  vmcontext,
				NICs:     vmnics,
				Disks:    vmdisks,
				Graphics: vmgraphics,
				OS:       vmos,
			}
		} else {
			vmtpl = &vmTemplate{
				CPU:      vmcpu.(float64),
				Context:  vmcontext,
				NICs:     vmnics,
				Disks:    vmdisks,
				Graphics: vmgraphics,
				OS:       vmos,
			}
		}
	} else {
		vmtpl = &vmTemplate{
			Context:  vmcontext,
			NICs:     vmnics,
			Disks:    vmdisks,
			Graphics: vmgraphics,
			OS:       vmos,
		}
	}

	w := &bytes.Buffer{}

	//Encode the VM template schema to XML
	enc := xml.NewEncoder(w)
	//enc.Indent("", "  ")
	if err := enc.Encode(vmtpl); err != nil {
		return "", err
	}

	log.Printf("VM XML: %s", w.String())
	return w.String(), nil
}

func resourceVMNicHash(v interface{}) int {
	var buf bytes.Buffer
	m := v.(map[string]interface{})
	buf.WriteString(fmt.Sprintf("%s-", m["model"].(string)))
	buf.WriteString(fmt.Sprintf("%d-", m["network_id"].(int)))
	return hashcode.String(buf.String())
}

func resourceVMCustomizeDiff(diff *schema.ResourceDiff, v interface{}) error {
	// If the VM is in error state, force the VM to be recreated
	if diff.Get("lcmstate") == 36 {
		log.Printf("[INFO] VM is in error state, forcing recreate.")
		diff.SetNew("lcmstate", 3)
		if err := diff.ForceNew("lcmstate"); err != nil {
			return err
		}
	}

	return nil
}
