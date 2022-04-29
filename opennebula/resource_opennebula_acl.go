package opennebula

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/acl"
)

var resourceMap = map[string]acl.Resources{
	"VM":             acl.VM,
	"HOST":           acl.Host,
	"NET":            acl.Net,
	"IMAGE":          acl.Image,
	"USER":           acl.User,
	"TEMPLATE":       acl.Template,
	"GROUP":          acl.Group,
	"DATASTORE":      acl.Datastore,
	"CLUSTER":        acl.Cluster,
	"DOCUMENT":       acl.Document,
	"ZONE":           acl.Zone,
	"SECGROUP":       acl.SecGroup,
	"VDC":            acl.Vdc,
	"VROUTER":        acl.VRouter,
	"MARKETPLACE":    acl.MarketPlace,
	"MARKETPLACEAPP": acl.MarketPlaceApp,
	"VMGROUP":        acl.VMGroup,
	"VNTEMPLATE":     acl.VNTemplate,
}

var rightMap = map[string]acl.Rights{
	"USE":    acl.Use,
	"MANAGE": acl.Manage,
	"ADMIN":  acl.Admin,
	"CREATE": acl.Create,
}

func resourceOpennebulaACL() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpennebulaACLCreate,
		Read:   resourceOpennebulaACLRead,
		Delete: resourceOpennebulaACLDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
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
		},
	}
}

func resourceOpennebulaACLCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	userHex, err := acl.ParseUsers(d.Get("user").(string))
	if err != nil {
		return err
	}

	resourceHex, err := acl.ParseResources(d.Get("resource").(string))
	if err != nil {
		return err
	}

	rightsHex, err := acl.ParseRights(d.Get("rights").(string))
	if err != nil {
		return err
	}

	aclID, err := controller.ACLs().CreateRule(userHex, resourceHex, rightsHex)
	if err != nil {
		return err
	}
	d.SetId(fmt.Sprintf("%v", aclID))

	return resourceOpennebulaACLRead(d, meta)
}

func resourceOpennebulaACLRead(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)
	acls, err := controller.ACLs().Info()

	if err != nil {
		return err
	}

	numericID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Failed to convert %+v to integer: %+v", d.Id(), err)
	}

	for _, acl := range acls.ACLs {
		if acl.ID == numericID {
			// We don't call Set because that would overwrite our string values
			// With raw numbers.
			// We only check if an ACL with the given ID exists, and return an error if not.
			return nil
		}
	}

	return fmt.Errorf("ACL with ID '%+v' not found.", d.Id())
}

func resourceOpennebulaACLDelete(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	numericID, err := strconv.Atoi(d.Id())
	if err != nil {
		return fmt.Errorf("Failed to convert %+v to integer: %+v", d.Id(), err)
	}

	return controller.ACLs().DeleteRule(numericID)
}
