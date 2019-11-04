package opennebula

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/image"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

// ImageTemplate is the definition of an ONE image template
type ImageTemplate struct {
	DevPrefix   string `xml:"DEV_PREFIX,omitempty"`
	Driver      string `xml:"DRIVER,omitempty"`
	Format      string `xml:"FORMAT,omitempty"`
	Target      string `xml:"TARGET,omitempty"`
	Description string `xml:"DESCRIPTION,omitempty"`
}

var imagetypes = []string{"OS", "CDROM", "DATABLOCK", "KERNEL", "RAMDISK", "CONTEXT"}
var locktypes = []string{"USE", "MANAGE", "ADMIN", "ALL", "UNLOCK"}

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
			"lock": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Lock level of the new Image: USE, MANAGE, ADMIN, ALL, UNLOCK",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)

					if inArray(value, locktypes) < 0 {
						errors = append(errors, fmt.Errorf("Type %q must be one of: %s", k, strings.Join(locktypes, ",")))
					}

					return
				},
			},
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
			"group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Group that onws the Image, If empty, it uses caller group",
			},
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
			return nil, fmt.Errorf("Image Id (%s) is not an integer", d.Id())
		}
		ic = controller.Image(int(gid))
	}

	// Otherwise, try to find the Image by name as the de facto compound primary key
	if d.Id() == "" {
		gid, err := controller.Images().ByName(d.Get("name").(string), args...)
		if err != nil {
			d.SetId("")
			return nil, fmt.Errorf("Could not find Image with name %s, got id: %d", d.Get("name").(string), gid)
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
		gid, err = controller.Groups().ByName(d.Get("group").(string))
		if err != nil {
			return err
		}
	}

	err = ic.Chown(-1, gid)
	if err != nil {
		return err
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

		imagexml, xmlerr := generateImageXML(d)
		if xmlerr != nil {
			return xmlerr
		}

		imageID, err = controller.Images().Create(imagexml, uint(d.Get("datastore_id").(int)))
		if err != nil {
			return err
		}
	}

	ic := controller.Image(imageID)

	template, xmlerr := generateImageTemplate(d)
	if xmlerr != nil {
		return xmlerr
	}

	// add template information into image
	err = ic.Update(template, 1)

	d.SetId(fmt.Sprintf("%v", imageID))

	_, err = waitForImageState(d, meta, "ready")
	if err != nil {
		return fmt.Errorf("Error waiting for Image (%s) to be in state READY: %s", d.Id(), err)
	}

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

	if lock, ok := d.GetOk("lock"); ok {
		if lock.(string) == "UNLOCK" {
			err = ic.Unlock()
		} else {
			var level shared.LockLevel
			err = StringToLockLevel(lock.(string), &level)
			if err != nil {
				return err
			}
			err = ic.Lock(level)
		}
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

func waitForImageState(d *schema.ResourceData, meta interface{}, state string) (interface{}, error) {
	var ic *goca.ImageController
	var image *image.Image
	var err error
	//Get Image
	ic, err = getImageController(d, meta)
	if err != nil {
		return image, err
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{"anythingelse"},
		Target:  []string{state},
		Refresh: func() (interface{}, string, error) {
			log.Println("Refreshing Image state...")
			if d.Id() != "" {
				// Get Image Info
				ic, err = getImageController(d, meta)
				if err != nil {
					log.Printf("Image %v was not found", d.Id())
					return image, "notfound", nil
				}
			}
			// TODO: fix it after 5.10 release
			// Force the "decrypt" bool to false to keep ONE 5.8 behavior
			image, err = ic.Info(false)
			if err != nil {
				if strings.Contains(err.Error(), "Error getting image") {
					return image, "notfound", nil
				}
				return image, "", err
			}
			state, err := image.StateString()
			if err != nil {
				if strings.Contains(err.Error(), "Error getting image") {
					return image, "notfound", nil
				}
				return image, "notfound", err
			}
			log.Printf("Image %v is currently in state %v", image.ID, state)
			if state == "READY" {
				return image, "ready", nil
			} else if state == "ERROR" {
				return image, "error", fmt.Errorf("image ID %v entered error state", d.Id())
			} else {
				return image, "anythingelse", nil
			}
		},
		Timeout:    10 * time.Minute,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForState()
}

func resourceOpennebulaImageRead(d *schema.ResourceData, meta interface{}) error {
	// Get all images
	ic, err := getImageController(d, meta, -2, -1, -1)
	if err != nil {
		return err
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	image, err := ic.Info(false)
	if err != nil {
		return err
	}

	imageTypeIDName := map[int]string{
		0: "OS",
		1: "CDROM",
		2: "DATABLOCK",
		3: "KERNEL",
		4: "RAMDISK",
		5: "CONTEXT",
	}

	d.SetId(fmt.Sprintf("%v", image.ID))
	d.Set("name", image.Name)
	d.Set("uid", image.UID)
	d.Set("gid", image.GID)
	d.Set("uname", image.UName)
	d.Set("gname", image.GName)
	d.Set("permissions", permissionsUnixString(image.Permissions))
	d.Set("persistent", image.PersistentValue)
	d.Set("path", image.Path)

	imageidx, err := strconv.Atoi(image.Type)
	if err != nil {
		return err
	}
	if val, ok := imageTypeIDName[imageidx]; ok {
		d.Set("type", val)
	}

	d.Set("size", image.Size)
	devpref, err := image.Template.Dynamic.GetContentByName("DEV_PREFIX")
	if err == nil {
		d.Set("dev_prefix", devpref)
	}
	driver, err := image.Template.Dynamic.GetContentByName("DRIVER")
	if err == nil {
		d.Set("driver", driver)
	}
	format, err := image.Template.Dynamic.GetContentByName("FORMAT")
	if err == nil {
		d.Set("format", format)
	}
	desc, err := image.Template.Dynamic.GetContentByName("DESCRIPTION")
	if err == nil {
		d.Set("description", desc)
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

	if d.HasChange("lock") {
		lock := d.Get("lock").(string)
		if lock == "UNLOCK" {
			err = ic.Unlock()
		} else {
			var level shared.LockLevel
			err = StringToLockLevel(lock, &level)
			if err != nil {
				return err
			}
			err = ic.Lock(level)
		}
		if err != nil {
			return err
		}
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

	return nil
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

	_, err = waitForImageState(d, meta, "notfound")
	if err != nil {
		return fmt.Errorf("Error waiting for Image (%s) to be in state NOTFOUND: %s", d.Id(), err)
	}

	return nil
}

func generateImageXML(d *schema.ResourceData) (string, error) {

	var imagetype string
	var imagesize int
	var imagepersistent int
	var imagepath string

	imagename := d.Get("name").(string)

	if val, ok := d.GetOk("type"); ok {
		imagetype = val.(string)
	}

	if d.Get("persistent").(bool) {
		imagepersistent = 1
	}

	if val, ok := d.GetOk("size"); ok {
		imagesize = val.(int)
	}

	if val, ok := d.GetOk("path"); ok {
		imagepath = val.(string)
	}

	imagetplfull := &image.Image{
		Name:            imagename,
		Size:            imagesize,
		Type:            imagetype,
		PersistentValue: imagepersistent,
		Path:            imagepath,
	}

	w := &bytes.Buffer{}

	//Encode the Image template schema to XML
	enc := xml.NewEncoder(w)
	if err := enc.Encode(imagetplfull); err != nil {
		return "", err
	}

	log.Printf("[INFO] Image Definition XML: %s", w.String())
	return w.String(), nil
}

func generateImageTemplate(d *schema.ResourceData) (string, error) {

	var imagedescription string
	var imagedevprefix string
	var imagedriver string
	//var imagedisktype string
	var imageformat string
	var imagetarget string

	if val, ok := d.GetOk("description"); ok {
		imagedescription = val.(string)
	}

	if val, ok := d.GetOk("dev_prefix"); ok {
		imagedevprefix = val.(string)
	}

	if val, ok := d.GetOk("driver"); ok {
		imagedriver = val.(string)
	}

	if val, ok := d.GetOk("format"); ok {
		imageformat = val.(string)
	}

	if val, ok := d.GetOk("target"); ok {
		imagetarget = val.(string)
	}

	imagetpl := &ImageTemplate{
		DevPrefix:   imagedevprefix,
		Driver:      imagedriver,
		Format:      imageformat,
		Target:      imagetarget,
		Description: imagedescription,
	}

	w := &bytes.Buffer{}

	//Encode the Image template schema to XML
	enc := xml.NewEncoder(w)
	if err := enc.Encode(imagetpl); err != nil {
		return "", err
	}

	log.Printf("[INFO] Image Template XML: %s", w.String())
	return w.String(), nil
}
