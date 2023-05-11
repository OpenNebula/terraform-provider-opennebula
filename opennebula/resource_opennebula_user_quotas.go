package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceOpennebulaUserQuotas() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaUserQuotasCreate,
		ReadContext:   resourceOpennebulaUserQuotasRead,
		UpdateContext: resourceOpennebulaUserQuotasUpdate,
		DeleteContext: resourceOpennebulaUserQuotasDelete,
		Importer: &schema.ResourceImporter{
			StateContext: resourceOpennebulaUserQuotasImportState,
		},
		Schema: mergeSchemas(map[string]*schema.Schema{
			"user_id": {
				Type:        schema.TypeInt,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the user to apply the quota",
			},
		},
			quotasMapSchema()),
	}
}

func resourceOpennebulaUserQuotasCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	userID := d.Get("user_id").(int)
	uc := controller.User(userID)

	quotasStr, err := generateQuotas(d, false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to generate quotas description",
			Detail:   fmt.Sprintf("user (ID: %s) quotas: %s", d.Id(), err),
		})
		return diags
	}
	log.Printf("[DEBUG] create quotasStr: %s", quotasStr)
	err = uc.Quota(quotasStr)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to apply quotas",
			Detail:   fmt.Sprintf("user (ID: %s) quotas: %s", d.Id(), err),
		})
		return diags
	}

	d.SetId(fmt.Sprint(userID))

	return resourceOpennebulaUserQuotasRead(ctx, d, meta)
}

func resourceOpennebulaUserQuotasRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	userID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse the user quotas ID",
			Detail:   err.Error(),
		})
		return diags
	}
	d.Set("user_id", userID)

	userInfos, err := controller.User(int(userID)).Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("user (ID: %s) quotas: %s", d.Id(), err),
		})
		return diags

	}

	err = flattenQuotasMapFromStructs(d, &userInfos.QuotasList)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten quotas",
			Detail:   fmt.Sprintf("user (ID: %s) quotas: %s", d.Id(), err),
		})
		return diags
	}

	return diags
}

func resourceOpennebulaUserQuotasUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	userID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse the user quotas ID",
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
					Detail:   fmt.Sprintf("user (ID: %s) quotas: %s", d.Id(), err),
				})
				return diags
			}
		}
		err = controller.User(int(userID)).Quota(quotasStr)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed apply quotas",
				Detail:   fmt.Sprintf("user (ID: %s) quotas: %s", d.Id(), err),
			})
			return diags
		}

	}

	return resourceOpennebulaUserQuotasRead(ctx, d, meta)
}

func resourceOpennebulaUserQuotasDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	userID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse the user quotas ID",
			Detail:   err.Error(),
		})
		return diags
	}

	quotasStr, err := generateQuotas(d, true)
	if err != nil {
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to generate quotas description",
				Detail:   fmt.Sprintf("user (ID: %s) quotas: %s", d.Id(), err),
			})
			return diags
		}
	}

	err = controller.User(int(userID)).Quota(quotasStr)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed apply quotas",
			Detail:   fmt.Sprintf("user (ID: %s) quotas: %s", d.Id(), err),
		})
		return diags
	}

	return nil
}

func resourceOpennebulaUserQuotasImportState(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	fullID := d.Id()
	parts := strings.Split(fullID, ":")

	if len(parts) < 2 {
		return nil, fmt.Errorf("Invalid ID format. Expected: user_id:quotas_section")
	}

	userID, err := strconv.ParseInt(parts[0], 10, 0)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse user ID: %s", err)
	}

	d.SetId(fmt.Sprint(userID))
	d.Set("user_id", userID)

	quotaType := parts[1]
	d.Set("type", quotaType)

	if inArray(quotaType, validQuotaTypes) < 0 {
		return nil, fmt.Errorf("Invalid quota type %q must be one of: %s", quotaType, strings.Join(validQuotaTypes, ","))
	}

	return []*schema.ResourceData{d}, nil
}
