package opennebula

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/template"
)

func resourceOpennebulaVirtualRouterInstanceTemplate() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaVirtualRouterInstanceTemplateCreate,
		ReadContext:   resourceOpennebulaVirtualRouterInstanceTemplateRead,
		Exists:        resourceOpennebulaVirtualRouterInstanceTemplateExists,
		UpdateContext: resourceOpennebulaVirtualRouterInstanceTemplateUpdate,
		DeleteContext: resourceOpennebulaVirtualRouterInstanceTemplateDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: commonTemplateSchemas(),
	}
}

func resourceOpennebulaVirtualRouterInstanceTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	err := resourceOpennebulaTemplateCreateCustom(d, meta, customVirtualRouterInstanceTemplate)

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router instance template creation failed",
			Detail:   fmt.Sprintf("VM (ID:%s) reading failed: %s", d.Id(), err),
		})
	}

	return resourceOpennebulaVirtualRouterInstanceTemplateRead(ctx, d, meta)
}

func customVirtualRouterInstanceTemplate(d *schema.ResourceData, tpl *dyn.Template) error {

	tpl.AddPair("VROUTER", "YES")

	return nil
}

func resourceOpennebulaVirtualRouterInstanceTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	err := resourceOpennebulaTemplateReadCustom(d, meta, customVirtualRouterInstanceTemplateRead)

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router instance template reading failed",
			Detail:   fmt.Sprintf("VM (ID:%s) reading failed: %s", d.Id(), err),
		})
	}

	return diags
}

func customVirtualRouterInstanceTemplateRead(d *schema.ResourceData, tpl *template.Template) error {

	vrouter, _ := tpl.Template.Get("VROUTER")
	if vrouter != "YES" {
		return fmt.Errorf("template with ID %d is not a template of a virtual router instance", tpl.ID)
	}

	return nil
}

func resourceOpennebulaVirtualRouterInstanceTemplateExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	config := meta.(*Configuration)
	controller := config.Controller

	tplID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		return false, err
	}

	_, err = controller.Template(int(tplID)).Info(false, false)
	if NoExists(err) {
		return false, err
	}

	return true, err
}

func resourceOpennebulaVirtualRouterInstanceTemplateUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	err := resourceOpennebulaTemplateUpdateCustom(d, meta)

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router instance template reading failed",
			Detail:   fmt.Sprintf("VM (ID:%s) reading failed: %s", d.Id(), err),
		})
	}

	return resourceOpennebulaVirtualRouterInstanceTemplateRead(ctx, d, meta)
}

func resourceOpennebulaVirtualRouterInstanceTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	err := resourceOpennebulaTemplateDelete(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router instance template delete failed",
			Detail:   fmt.Sprintf("template (ID:%s) delete failed: %s", d.Id(), err),
		})
	}

	return diags
}
