package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOpennebulaGroupQuotas() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaGroupQuotasCreate,
		ReadContext:   resourceOpennebulaGroupQuotasRead,
		UpdateContext: resourceOpennebulaGroupQuotasUpdate,
		DeleteContext: resourceOpennebulaGroupQuotasDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: mergeSchemas(map[string]*schema.Schema{
			"group_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the group to apply the quota",
			},
		},
			quotasMapSchema()),
	}
}

func resourceOpennebulaGroupQuotasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	groupID := d.Get("group_id").(int)
	gc := controller.Group(groupID)

	quotasStr, err := generateQuotas(d, false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to generate quotas description",
			Detail:   fmt.Sprintf("group (ID: %s) quotas: %s", d.Id(), err),
		})
		return diags
	}
	log.Printf("[DEBUG] create quotasStr: %s", quotasStr)
	err = gc.Quota(quotasStr)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to apply quotas",
			Detail:   fmt.Sprintf("group (ID: %s) quotas: %s", d.Id(), err),
		})
		return diags
	}

	d.SetId(fmt.Sprint(groupID))

	return resourceOpennebulaGroupQuotasRead(ctx, d, meta)
}

func resourceOpennebulaGroupQuotasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	groupID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse the group quotas ID",
			Detail:   err.Error(),
		})
		return diags
	}
	d.Set("group_id", groupID)

	groupInfos, err := controller.Group(int(groupID)).Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("group (ID: %s) quotas: %s", d.Id(), err),
		})
		return diags

	}

	err = flattenQuotasMapFromStructs(d, &groupInfos.QuotasList)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten quotas",
			Detail:   fmt.Sprintf("group (ID: %s) quotas: %s", d.Id(), err),
		})
		return diags
	}

	return diags
}

func resourceOpennebulaGroupQuotasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	groupID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse the group quotas ID",
			Detail:   err.Error(),
		})
		return diags
	}

	if d.HasChange("datastore") || d.HasChange("network") || d.HasChange("image") || d.HasChange("vm") {
		quotasStr, err := generateQuotas(d, false)
		if err != nil {
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to generate quotas description",
					Detail:   fmt.Sprintf("group (ID: %s) quotas: %s", d.Id(), err),
				})
				return diags
			}
		}
		err = controller.Group(int(groupID)).Quota(quotasStr)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed apply quotas",
				Detail:   fmt.Sprintf("group (ID: %s) quotas: %s", d.Id(), err),
			})
			return diags
		}

	}

	return resourceOpennebulaGroupQuotasRead(ctx, d, meta)
}

func resourceOpennebulaGroupQuotasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	groupID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse the group quotas ID",
			Detail:   err.Error(),
		})
		return diags
	}

	quotasStr, err := generateQuotas(d, true)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to generate quotas description",
			Detail:   fmt.Sprintf("group (ID: %s) quotas: %s", d.Id(), err),
		})
		return diags
	}

	err = controller.Group(int(groupID)).Quota(quotasStr)

	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed apply quotas",
			Detail:   fmt.Sprintf("group (ID: %s) quotas: %s", d.Id(), err),
		})
		return diags
	}

	return nil
}
