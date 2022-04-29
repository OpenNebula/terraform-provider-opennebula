package opennebula

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/image"
	img "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/image"
	imk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/image/keys"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

var imagetypes = []string{"OS", "CDROM", "DATABLOCK", "KERNEL", "RAMDISK", "CONTEXT"}

func resourceOpennebulaImage() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpennebulaImageCreate,
		Read:   resourceOpennebulaImageRead,
		Exists: resourceOpennebulaImageExists,
		Update: resourceOpennebulaImageUpdate,
		Delete: resourceOpennebulaImageDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Image",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the Image, in OpenNebula's XML or String format",
			},
			"permissions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Permissions for the Image (in Unix format, owner-group-other, use-manage-admin)",
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
				Description: "ID of the user that will own the Image",
			},
			"gid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group that will own the Image",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the Image",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the Image",
			},
			"clone_from_image": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				Description:   "ID or name of the Image to be cloned from",
				ConflictsWith: []string{"path", "size", "type"},
			},
			"datastore_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the datastore where Image will be stored",
			},
			"persistent": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Flag which indicates if the Image has to be persistent",
			},
			"lock": lockSchema(),
			"path": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				Description:   "Path to the new image (local path on the OpenNebula server or URL)",
				ConflictsWith: []string{"clone_from_image"},
			},
			"type": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"clone_from_image"},
				Description:   "Type of the new Image: OS, CDROM, DATABLOCK, KERNEL, RAMDISK, CONTEXT",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)

					if inArray(value, imagetypes) < 0 {
						errors = append(errors, fmt.Errorf("Type %q must be one of: %s", k, strings.Join(imagetypes, ",")))
					}

					return
				},
			},
			"size": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"clone_from_image"},
				Description:   "Size of the new image in MB",
			},
			"dev_prefix": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Device prefix, normally one of: hd, sd, vd",
			},
			"target": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Device target, example: vda",
			},
			"driver": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Driver to use, normally 'raw' or 'qcow2'",
			},
			"format": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Image format, normally 'raw' or 'qcow2'",
			},
			"timeout": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     10,
				Description: "Timeout (in minutes) within resource should be available. Default: 10 minutes",
			},
			"group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Group that onws the Image, If empty, it uses caller group",
			},
			"tags": tagsSchema(),
		},
	}
}

// getImagecontroller
// * d: ResourceData. Terraform ResrouceData information
// * meta: Interface. Interface use to interact with remote server via terraform core
// * args: Viable arguments to manage ImagePool variable arguments
//   see http://docs.opennebula.org/5.8/integration/system_interfaces/api.html#one-imagepool-info for details
func getImageController(d *schema.ResourceData, meta interface{}, args ...int) (*goca.ImageController, error) {
	controller := meta.(*goca.Controller)
	var ic *goca.ImageController

	// Try to find the Image by ID, if specified
	if d.Id() != "" {
		gid, err := strconv.ParseUint(d.Id(), 10, 64)
		if err != nil {
			return nil, err
		}
		ic = controller.Image(int(gid))
	}

	// Otherwise, try to find the Image by name as the de facto compound primary key
	if d.Id() == "" {
		gid, err := controller.Images().ByName(d.Get("name").(string), args...)
		if err != nil {
			return nil, err
		}
		ic = controller.Image(gid)
	}

	return ic, nil
}

// changeImageGroup: function to change Image Group ownership
func changeImageGroup(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)
	var gid int

	ic, err := getImageController(d, meta)
	if err != nil {
		return err
	}

	if d.Get("group") != "" {
		group := d.Get("group").(string)
		gid, err = controller.Groups().ByName(group)
		if err != nil {
			return fmt.Errorf("Can't find a group with name `%s`: %s", group, err)
		}
	}

	err = ic.Chown(-1, gid)
	if err != nil {
		return fmt.Errorf("Can't find a group with ID `%d`: %s", gid, err)
	}

	return nil
}

func resourceOpennebulaImageCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)
	var imageID int
	var err error

	// Check if Image ID for cloning is set
	if len(d.Get("clone_from_image").(string)) > 0 {
		imageID, err = resourceOpennebulaImageClone(d, meta)
		if err != nil {
			return err
		}
	} else { //Otherwise allocate a new image
		var err error

		imgDef, err := generateImage(d)
		if err != nil {
			return err
		}

		imageID, err = controller.Images().Create(imgDef, uint(d.Get("datastore_id").(int)))
		if err != nil {
			return err
		}
	}

	ic := controller.Image(imageID)

	imgTpl, err := generateImageTemplate(d)
	if err != nil {
		return err
	}

	timeout := d.Get("timeout").(int)
	_, err = waitForImageState(ic, timeout, "READY")
	if err != nil {
		return fmt.Errorf("Error waiting for Image (%s) to be in state READY: %s", d.Id(), err)
	}

	// add template information into image
	err = ic.Update(imgTpl, 1)

	d.SetId(fmt.Sprintf("%v", imageID))

	ic, err = getImageController(d, meta)
	if err != nil {
		return err
	}

	// update permisions
	if perms, ok := d.GetOk("permissions"); ok {
		err = ic.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			return err
		}
	}

	if d.Get("group") != "" {
		err = changeImageGroup(d, meta)
		if err != nil {
			return err
		}
	}

	if d.Get("persistent").(bool) {
		err = ic.Persistent(d.Get("persistent").(bool))
		if err != nil {
			return err
		}
	}

	if lock, ok := d.GetOk("lock"); ok && lock.(string) != "UNLOCK" {

		var level shared.LockLevel
		err = StringToLockLevel(lock.(string), &level)
		if err != nil {
			return err
		}

		err = ic.Lock(level)
		if err != nil {
			return err
		}
	}

	return resourceOpennebulaImageRead(d, meta)
}

func resourceOpennebulaImageClone(d *schema.ResourceData, meta interface{}) (int, error) {
	controller := meta.(*goca.Controller)
	var originalic *goca.ImageController

	//Test if clone_from_image is an integer or not
	if val, err := strconv.Atoi(d.Get("clone_from_image").(string)); err == nil {
		originalic = controller.Image(int(val))
	} else {
		imageID, err := controller.Images().ByName(d.Get("clone_from_image").(string))
		if err != nil {
			return 0, fmt.Errorf("Unable to find Image by name %s", d.Get("clone_from_image"))
		}
		originalic = controller.Image(imageID)
	}

	// Clone Image from given ID
	return originalic.Clone(d.Get("name").(string), d.Get("datastore_id").(int))
}

func waitForImageState(ic *goca.ImageController, timeout int, state ...string) (interface{}, error) {

	stateConf := &resource.StateChangeConf{
		Pending: []string{"anythingelse"},
		Target:  state,
		Refresh: func() (interface{}, string, error) {

			log.Println("Refreshing Image state...")

			// TODO: fix it after 5.10 release
			// Force the "decrypt" bool to false to keep ONE 5.8 behavior
			imgInfos, err := ic.Info(false)
			if err != nil {
				if NoExists(err) {
					return imgInfos, "notfound", nil
				}
				return imgInfos, "", err
			}
			state, err := imgInfos.State()
			if err != nil {
				return imgInfos, "", err
			}

			log.Printf("Image (ID:%d, name:%s) is currently in state %v", imgInfos.ID, imgInfos.Name, state.String())

			switch state {
			case img.Ready:
				return imgInfos, state.String(), nil
			case img.Error:
				return imgInfos, state.String(), fmt.Errorf("Image (ID:%d) entered error state.", imgInfos.ID)
			default:
				return imgInfos, "anythingelse", nil
			}
		},
		Timeout:    time.Duration(timeout) * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForState()
}

func resourceOpennebulaImageRead(d *schema.ResourceData, meta interface{}) error {
	// Get all images
	ic, err := getImageController(d, meta, -2, -1, -1)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing image %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		return err
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	image, err := ic.Info(false)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", image.ID))
	d.Set("name", image.Name)
	d.Set("uid", image.UID)
	d.Set("gid", image.GID)
	d.Set("uname", image.UName)
	d.Set("gname", image.GName)
	d.Set("permissions", permissionsUnixString(*image.Permissions))
	if image.Persistent != nil {
		d.Set("persistent", *image.Persistent)
	}
	d.Set("path", image.Path)

	if inArray(image.Type, imagetypes) >= 0 {
		d.Set("type", image.Type)
	}

	tags := make(map[string]interface{})
	for i, _ := range image.Template.Elements {
		pair, ok := image.Template.Elements[i].(*dyn.Pair)
		if !ok {
			continue
		}

		switch pair.Key() {
		case "DEV_PREFIX":
			devpref, err := image.Template.GetStr("DEV_PREFIX")
			if err == nil {
				d.Set("dev_prefix", devpref)
			}

		case "DRIVER":
			driver, err := image.Template.GetStr("DRIVER")
			if err == nil {
				d.Set("driver", driver)
			}

		case "FORMAT":
			format, err := image.Template.GetStr("FORMAT")
			if err == nil {
				d.Set("format", format)
			}

		case "DESCRIPTION":
			desc, err := image.Template.GetStr("DESCRIPTION")
			if err == nil {
				d.Set("description", desc)
			}

		default:
			if tagsInterface, ok := d.GetOk("tags"); ok {
				for k, _ := range tagsInterface.(map[string]interface{}) {
					if strings.ToUpper(k) == pair.Key() {
						tags[k] = pair.Value
					}
				}
			}
		}
	}

	if len(tags) > 0 {
		err := d.Set("tags", tags)
		if err != nil {
			return err
		}
	}

	if image.LockInfos != nil {
		d.Set("lock", LockLevelToString(image.LockInfos.Locked))
	}

	return nil
}

func resourceOpennebulaImageExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceOpennebulaImageRead(d, meta)
	if err != nil || d.Id() == "" {
		return false, err
	}

	return true, nil
}

func resourceOpennebulaImageUpdate(d *schema.ResourceData, meta interface{}) error {
	//Get Image
	ic, err := getImageController(d, meta)
	if err != nil {
		return err
	}
	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	image, err := ic.Info(false)
	if err != nil {
		return err
	}

	lock, lockOk := d.GetOk("lock")
	if d.HasChange("lock") && lockOk && lock.(string) == "UNLOCK" {

		err = ic.Unlock()
		if err != nil {
			return err
		}
	}

	if d.HasChange("name") {
		err := ic.Rename(d.Get("name").(string))
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated name for Image %s\n", image.Name)
	}

	if d.HasChange("permissions") {
		if perms, ok := d.GetOk("permissions"); ok {
			err = ic.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				return err
			}
		}
		log.Printf("[INFO] Successfully updated Image %s\n", image.Name)
	}

	if d.HasChange("group") {
		err = changeImageGroup(d, meta)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated group for Image %s\n", image.Name)
	}

	if d.HasChange("persistent") {
		err = ic.Persistent(d.Get("persistent").(bool))
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated persistent flag for Image %s\n", image.Name)
	}

	if d.HasChange("type") {
		if imagetype, ok := d.GetOk("permissions"); ok {
			err = ic.Chtype(imagetype.(string))
			if err != nil {
				return err
			}
		}
		log.Printf("[INFO] Successfully updated Image Type %s\n", image.Name)
	}

	update := false
	tpl := img.NewTemplate()

	if d.HasChange("description") {
		update = true
		tpl.Add("DESCRIPTION", d.Get("description").(string))
	}

	if update {
		err = ic.Update(tpl.String(), 1)
		if err != nil {
			return err
		}
	}

	if d.HasChange("tags") {
		tagsInterface := d.Get("tags").(map[string]interface{})
		for k, v := range tagsInterface {
			image.Template.Del(strings.ToUpper(k))
			image.Template.AddPair(strings.ToUpper(k), v)
		}

		err = ic.Update(image.Template.String(), 1)
		if err != nil {
			return err
		}
	}

	if d.HasChange("lock") && lockOk && lock.(string) != "UNLOCK" {

		var level shared.LockLevel

		err = StringToLockLevel(lock.(string), &level)
		if err != nil {
			return err
		}

		err = ic.Lock(level)
		if err != nil {
			return err
		}
	}

	return resourceOpennebulaImageRead(d, meta)
}

func resourceOpennebulaImageDelete(d *schema.ResourceData, meta interface{}) error {
	ic, err := getImageController(d, meta)
	if err != nil {
		return err
	}

	err = ic.Delete()
	if err != nil {
		return err
	}
	log.Printf("[INFO] Successfully deleted Image ID %s\n", d.Id())

	timeout := d.Get("timeout").(int)
	_, err = waitForImageState(ic, timeout, "notfound")
	if err != nil {
		return fmt.Errorf("Error waiting for Image (%s) to be in state NOTFOUND: %s", d.Id(), err)
	}

	return nil
}

func generateImage(d *schema.ResourceData) (string, error) {

	tpl := image.NewTemplate()

	imgName := d.Get("name").(string)
	tpl.Add(imk.Name, imgName)

	if val, ok := d.GetOk("type"); ok {
		tpl.Add(imk.Type, val.(string))
	}

	if d.Get("persistent").(bool) {
		tpl.Add(imk.Persistent, "1")
	}

	if val, ok := d.GetOk("size"); ok {
		imgSize := fmt.Sprint(val.(int))
		tpl.Add(imk.Size, imgSize)
	}

	if val, ok := d.GetOk("path"); ok {
		tpl.Add(imk.Path, val.(string))
	}

	tplStr := tpl.String()
	log.Printf("[INFO] Image definition: %s", tplStr)

	return tplStr, nil
}

func generateImageTemplate(d *schema.ResourceData) (string, error) {

	tpl := image.NewTemplate()

	if val, ok := d.GetOk("description"); ok {
		tpl.Add("DESCRIPTION", val.(string))
	}

	if val, ok := d.GetOk("dev_prefix"); ok {
		tpl.Add(imk.DevPrefix, val.(string))
	}

	if val, ok := d.GetOk("driver"); ok {
		tpl.Add(imk.Driver, val.(string))
	}

	if val, ok := d.GetOk("format"); ok {
		tpl.Add("FORMAT", val.(string))
	}

	if val, ok := d.GetOk("target"); ok {
		tpl.Add(imk.Target, val.(string))
	}

	tagsInterface := d.Get("tags").(map[string]interface{})
	for k, v := range tagsInterface {
		tpl.AddPair(strings.ToUpper(k), v)
	}

	str := tpl.String()
	log.Printf("[INFO] Image template: %s", str)

	return str, nil
}
