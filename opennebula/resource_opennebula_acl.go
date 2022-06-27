package opennebula

import (
	"context"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/acl"
)

func resourceOpennebulaACL() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaACLCreate,
		ReadContext:   resourceOpennebulaACLRead,
		DeleteContext: resourceOpennebulaACLDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: map[string]*schema.Schema{
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "User component of the new rule. ACL String Syntax is expected.",
			},
			"resource": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Resource component of the new rule. ACL String Syntax is expected.",
			},
			"rights": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Rights component of the new rule. ACL String Syntax is expected.",
			},
			"zone": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Zone component of the new rule. ACL String Syntax is expected.",
			},
		},
	}
}

func resourceOpennebulaACLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics
	var err error

	userHex, err := acl.ParseUsers(d.Get("user").(string))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse ACL users",
			Detail:   err.Error(),
		})
		return diags
	}

	resourceHex, err := acl.ParseResources(d.Get("resource").(string))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse ACL resources",
			Detail:   err.Error(),
		})
		return diags
	}

	rightsHex, err := acl.ParseRights(d.Get("rights").(string))
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse ACL rights",
			Detail:   err.Error(),
		})
		return diags
	}

	var aclID int
	var zoneHex string
	zone := d.Get("zone").(string)
	if len(zone) > 0 {
		zoneHex, err = acl.ParseZone(zone)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to parse zone",
				Detail:   err.Error(),
			})
			return diags
		}

		aclID, err = controller.ACLs().CreateRule(userHex, resourceHex, rightsHex, zoneHex)
	} else {
		aclID, err = controller.ACLs().CreateRule(userHex, resourceHex, rightsHex)
	}
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to create the ACL rule",
			Detail:   err.Error(),
		})
		return diags
	}
	d.SetId(fmt.Sprintf("%v", aclID))

	return resourceOpennebulaACLRead(ctx, d, meta)
}

func resourceOpennebulaACLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	acls, err := controller.ACLs().Info()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve ACL informations",
			Detail:   fmt.Sprintf("ACL (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	numericID, err := strconv.Atoi(d.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse the ACL rule ID",
			Detail:   fmt.Sprintf("ACL (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	for _, acl := range acls.ACLs {
		if acl.ID == numericID {
			// We don't call Set because that would overwrite our string values
			// With raw numbers.
			// We only check if an ACL with the given ID exists, and return an error if not.
			return nil
		}
	}

	diags = append(diags, diag.Diagnostic{
		Severity: diag.Error,
		Summary:  fmt.Sprintf("Failed to find ACL rule %s", d.Id()),
		Detail:   fmt.Sprintf("ACL (ID: %s): %s", d.Id(), err),
	})

	return diags
}

func resourceOpennebulaACLDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller

	var diags diag.Diagnostics

	numericID, err := strconv.Atoi(d.Id())
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to parse ACL rule ID",
			Detail:   fmt.Sprintf("ACL (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	err = controller.ACLs().DeleteRule(numericID)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete ACL rule",
			Detail:   fmt.Sprintf("ACL (ID: %s): %s", d.Id(), err),
		})
	}

	return diags

}
