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
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: commonTemplateSchemas(),
	}
}

func resourceOpennebulaVirtualRouterInstanceTemplateCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	diags := resourceOpennebulaTemplateCreateCustom(ctx, d, meta, customVirtualRouterInstanceTemplate)

	if len(diags) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create virtual router instance template",
		})
	}

	return resourceOpennebulaVirtualRouterInstanceTemplateRead(ctx, d, meta)
}

func customVirtualRouterInstanceTemplate(d *schema.ResourceData, tpl *dyn.Template) diag.Diagnostics {

	tpl.AddPair("VROUTER", "YES")

	return nil
}

func resourceOpennebulaVirtualRouterInstanceTemplateRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	diags := resourceOpennebulaTemplateReadCustom(ctx, d, meta, customVirtualRouterInstanceTemplateRead)

	if len(diags) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "virtual router instance template reading failed",
		})
	}

	return diags
}

func customVirtualRouterInstanceTemplateRead(ctx context.Context, d *schema.ResourceData, tpl *template.Template) diag.Diagnostics {

	var diags diag.Diagnostics

	vrouter, _ := tpl.Template.Get("VROUTER")
	if vrouter != "YES" {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Wrong virtual router instance template configuration",
			Detail:   fmt.Sprintf("template with ID %d is not a template of a virtual router instance", tpl.ID),
		})

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

	diags := resourceOpennebulaTemplateUpdateCustom(ctx, d, meta)
	if len(diags) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to update",
		})
	}

	return resourceOpennebulaVirtualRouterInstanceTemplateRead(ctx, d, meta)
}

func resourceOpennebulaVirtualRouterInstanceTemplateDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	diags := resourceOpennebulaTemplateDelete(ctx, d, meta)
	if len(diags) > 0 {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
		})
	}

	return diags
}
