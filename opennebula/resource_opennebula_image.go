package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/image"
	img "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/image"
	imk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/image/keys"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
)

var imagetypes = []string{"OS", "CDROM", "DATABLOCK", "KERNEL", "RAMDISK", "CONTEXT"}
var defaultImageMinTimeout = 20
var defaultImageTimeout = time.Duration(defaultImageMinTimeout) * time.Minute

func resourceOpennebulaImage() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaImageCreate,
		ReadContext:   resourceOpennebulaImageRead,
		Exists:        resourceOpennebulaImageExists,
		UpdateContext: resourceOpennebulaImageUpdate,
		DeleteContext: resourceOpennebulaImageDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultImageTimeout),
			Delete: schema.DefaultTimeout(defaultImageTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		CustomizeDiff: SetTagsDiff,
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

					if !contains(value, imagetypes) {
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
				Default:     defaultImageMinTimeout,
				Description: "Timeout (in minutes) within resource should be available. Default: 10 minutes",
				Deprecated:  "Native terraform timeout facilities should be used instead",
			},
			"group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Group that onws the Image, If empty, it uses caller group",
			},
			"tags":             tagsSchema(),
			"default_tags":     defaultTagsSchemaComputed(),
			"tags_all":         tagsSchemaComputed(),
			"template_section": templateSectionSchema(),
		},
	}
}

func getImageController(d *schema.ResourceData, meta interface{}) (*goca.ImageController, error) {
	config := meta.(*Configuration)
	controller := config.Controller

	imgID, err := strconv.ParseUint(d.Id(), 10, 0)
	if err != nil {
		return nil, err
	}

	return controller.Image(int(imgID)), nil
}

// changeImageGroup: function to change Image Group ownership
func changeImageGroup(d *schema.ResourceData, meta interface{}) error {
	config := meta.(*Configuration)
	controller := config.Controller
	var gid int

	ic, err := getImageController(d, meta)
	if err != nil {
		return err
	}

	group := d.Get("group").(string)
	gid, err = controller.Groups().ByName(group)
	if err != nil {
		return fmt.Errorf("Can't find a group with name `%s`: %s", group, err)
	}

	err = ic.Chown(-1, gid)
	if err != nil {
		return fmt.Errorf("Can't find a group with ID `%d`: %s", gid, err)
	}

	return nil
}

func resourceOpennebulaImageCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller
	var imageID int
	var err error
	var diags diag.Diagnostics

	// Check if Image ID for cloning is set
	if len(d.Get("clone_from_image").(string)) > 0 {
		imageID, err = resourceOpennebulaImageClone(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to clone image",
				Detail:   err.Error(),
			})
			return diags
		}
	} else { //Otherwise allocate a new image
		var err error

		imgDef, err := generateImage(d)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to generate description",
				Detail:   err.Error(),
			})
			return diags
		}

		imageID, err = controller.Images().Create(imgDef, uint(d.Get("datastore_id").(int)))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to create the image",
				Detail:   err.Error(),
			})
			return diags
		}
	}

	ic := controller.Image(imageID)

	imgTpl, err := generateImageTemplate(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to generate image content",
			Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	timeout := time.Duration(d.Get("timeout").(int)) * time.Minute
	if timeout == defaultImageTimeout {
		timeout = d.Timeout(schema.TimeoutCreate)
	}

	_, err = waitForImageState(ctx, ic, timeout, "READY")
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to wait image to be in READY state",
			Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	// add template information into image
	err = ic.Update(imgTpl, 1)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve information",
			Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%v", imageID))

	ic, err = getImageController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the image controller",
			Detail:   err.Error(),
		})
		return diags
	}

	// update permisions
	if perms, ok := d.GetOk("permissions"); ok {
		err = ic.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change permissions",
				Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.Get("group") != "" {
		err = changeImageGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.Get("persistent").(bool) {
		err = ic.Persistent(d.Get("persistent").(bool))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to modify persistency",
				Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if lock, ok := d.GetOk("lock"); ok && lock.(string) != "UNLOCK" {

		var level shared.LockLevel
		err = StringToLockLevel(lock.(string), &level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to convert lock level",
				Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		err = ic.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to lock",
				Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return resourceOpennebulaImageRead(ctx, d, meta)
}

func resourceOpennebulaImageClone(d *schema.ResourceData, meta interface{}) (int, error) {
	config := meta.(*Configuration)
	controller := config.Controller
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

func waitForImageState(ctx context.Context, ic *goca.ImageController, timeout time.Duration, state ...string) (interface{}, error) {

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
		Timeout:    timeout,
		Delay:      10 * time.Second,
		MinTimeout: 3 * time.Second,
	}

	return stateConf.WaitForStateContext(ctx)
}

func resourceOpennebulaImageRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	// Get all images
	ic, err := getImageController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the image controller",
			Detail:   err.Error(),
		})
		return diags

	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	image, err := ic.Info(false)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing image %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	d.SetId(fmt.Sprintf("%v", image.ID))
	d.Set("name", image.Name)
	d.Set("uid", image.UID)
	d.Set("gid", image.GID)
	d.Set("uname", image.UName)
	d.Set("gname", image.GName)
	d.Set("permissions", permissionsUnixString(*image.Permissions))
	if image.Persistent != nil {
		d.Set("persistent", *image.Persistent == 1)
	}
	d.Set("path", image.Path)

	if contains(image.Type, imagetypes) {
		d.Set("type", image.Type)
	}

	flattenDiags := flattenImageTemplate(d, meta, &image.Template)
	for _, diag := range flattenDiags {
		diags = append(diags, diag)
	}

	if image.LockInfos != nil {
		d.Set("lock", LockLevelToString(image.LockInfos.Locked))
	}

	return diags
}

func flattenImageTemplate(d *schema.ResourceData, meta interface{}, imageTpl *image.Template) diag.Diagnostics {
	var diags diag.Diagnostics

	err := flattenTemplateSection(d, meta, &imageTpl.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to read template section",
			Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
		})
	}

	for i, _ := range imageTpl.Elements {
		pair, ok := imageTpl.Elements[i].(*dyn.Pair)
		if !ok {
			continue
		}

		switch pair.Key() {
		case "DEV_PREFIX":
			devpref, err := imageTpl.GetStr("DEV_PREFIX")
			if err == nil {
				d.Set("dev_prefix", devpref)
			}

		case "DRIVER":
			driver, err := imageTpl.GetStr("DRIVER")
			if err == nil {
				d.Set("driver", driver)
			}

		case "FORMAT":
			format, err := imageTpl.GetStr("FORMAT")
			if err == nil {
				d.Set("format", format)
			}

		case "DESCRIPTION":
			desc, err := imageTpl.GetStr("DESCRIPTION")
			if err == nil {
				d.Set("description", desc)
			}

		default:
		}
	}

	flattenDiags := flattenTemplateTags(d, meta, &imageTpl.Template)
	for _, diag := range flattenDiags {
		diag.Detail = fmt.Sprintf("image (ID: %s): %s", d.Id(), err)
		diags = append(diags, diag)
	}

	return diags
}

func resourceOpennebulaImageExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	imageID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		return false, err
	}

	_, err = controller.Image(int(imageID)).Info(false)
	if NoExists(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, nil
}

func resourceOpennebulaImageUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics
	//Get Image
	ic, err := getImageController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the image controller",
			Detail:   err.Error(),
		})
		return diags
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	image, err := ic.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	lock, lockOk := d.GetOk("lock")
	if d.HasChange("lock") && lockOk && lock.(string) == "UNLOCK" {

		err = ic.Unlock()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to unlock",
				Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.HasChange("name") {
		err := ic.Rename(d.Get("name").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to rename",
				Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated name for Image %s\n", image.Name)
	}

	if d.HasChange("permissions") {
		if perms, ok := d.GetOk("permissions"); ok {
			err = ic.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to change permissions",
					Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
		log.Printf("[INFO] Successfully updated Image %s\n", image.Name)
	}

	if d.HasChange("group") {
		err = changeImageGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated group for Image %s\n", image.Name)
	}

	if d.HasChange("persistent") {
		err = ic.Persistent(d.Get("persistent").(bool))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to modify persistency",
				Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated persistent flag for Image %s\n", image.Name)
	}

	if d.HasChange("type") {
		if imagetype, ok := d.GetOk("type"); ok {
			err = ic.Chtype(imagetype.(string))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to change image type",
					Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
		log.Printf("[INFO] Successfully updated Image Type %s\n", image.Name)
	}

	update := false
	tpl := image.Template

	if d.HasChange("description") {
		tpl.Del("DESCRIPTION")

		description := d.Get("description").(string)

		if len(description) > 0 {
			tpl.Add("DESCRIPTION", description)
		}

		update = true
	}

	if d.HasChange("template_section") {

		updateTemplateSection(d, &tpl.Template)

		update = true
	}

	if d.HasChange("tags") {

		oldTagsIf, newTagsIf := d.GetChange("tags")
		oldTags := oldTagsIf.(map[string]interface{})
		newTags := newTagsIf.(map[string]interface{})

		// delete tags
		for k, _ := range oldTags {
			_, ok := newTags[k]
			if ok {
				continue
			}
			tpl.Del(strings.ToUpper(k))
		}

		// add/update tags
		for k, v := range newTags {
			key := strings.ToUpper(k)
			tpl.Del(key)
			tpl.AddPair(key, v)
		}

		update = true
	}

	if d.HasChange("tags_all") {
		oldTagsAllIf, newTagsAllIf := d.GetChange("tags_all")
		oldTagsAll := oldTagsAllIf.(map[string]interface{})
		newTagsAll := newTagsAllIf.(map[string]interface{})

		tags := d.Get("tags").(map[string]interface{})

		// delete tags
		for k, _ := range oldTagsAll {
			_, ok := newTagsAll[k]
			if ok {
				continue
			}
			tpl.Del(strings.ToUpper(k))
		}

		// reapply all default tags that were neither applied nor overriden via tags section
		for k, v := range newTagsAll {
			_, ok := tags[k]
			if ok {
				continue
			}

			key := strings.ToUpper(k)
			tpl.Del(key)
			tpl.AddPair(key, v)
		}

		update = true
	}

	if update {
		err = ic.Update(tpl.String(), parameters.Replace)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update image content",
				Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.HasChange("lock") && lockOk && lock.(string) != "UNLOCK" {

		var level shared.LockLevel

		err = StringToLockLevel(lock.(string), &level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to convert lock level",
				Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		err = ic.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to lock",
				Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return resourceOpennebulaImageRead(ctx, d, meta)
}

func resourceOpennebulaImageDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	ic, err := getImageController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the image controller",
			Detail:   err.Error(),
		})
		return diags
	}

	err = ic.Delete()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
			Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
		})
		return diags
	}
	log.Printf("[INFO] Successfully deleted Image ID %s\n", d.Id())

	timeout := time.Duration(d.Get("timeout").(int)) * time.Minute
	if timeout == defaultImageTimeout {
		timeout = d.Timeout(schema.TimeoutDelete)
	}

	_, err = waitForImageState(ctx, ic, timeout, "notfound")
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to wait image to be in NOTFOUND state",
			Detail:   fmt.Sprintf("image (ID: %s): %s", d.Id(), err),
		})
		return diags
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

func generateImageTemplate(d *schema.ResourceData, meta interface{}) (string, error) {

	config := meta.(*Configuration)
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

	vectorsInterface := d.Get("template_section").(*schema.Set).List()
	if len(vectorsInterface) > 0 {
		addTemplateVectors(vectorsInterface, &tpl.Template)
	}

	tagsInterface := d.Get("tags").(map[string]interface{})
	for k, v := range tagsInterface {
		tpl.AddPair(strings.ToUpper(k), v)
	}

	// add default tags if they aren't overriden
	if len(config.defaultTags) > 0 {
		for k, v := range config.defaultTags {
			key := strings.ToUpper(k)
			p, _ := tpl.GetPair(key)
			if p != nil {
				continue
			}
			tpl.AddPair(key, v)
		}
	}

	str := tpl.String()
	log.Printf("[INFO] Image template: %s", str)

	return str, nil
}
