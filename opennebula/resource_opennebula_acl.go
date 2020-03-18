package opennebula

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"

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

func aclCalculateIDs(idString string) (int64, error) {
	match, err := regexp.Match("^([\\#@\\%]\\d+|\\*)$", []byte(idString))
	if err != nil {
		return 0, err
	}
	if !match {
		return 0, fmt.Errorf("ID String %+v malformed", idString)
	}

	var value int64

	if strings.HasPrefix(idString, "#") {
		id, err := strconv.Atoi(strings.TrimLeft(idString, "#"))
		if err != nil {
			return 0, err
		}
		value = int64(acl.UID) + int64(id)
	}

	if strings.HasPrefix(idString, "@") {
		id, err := strconv.Atoi(strings.TrimLeft(idString, "@"))
		if err != nil {
			return 0, err
		}
		value = int64(acl.GID) + int64(id)
	}

	if strings.HasPrefix(idString, "*") {
		value = int64(acl.All)
	}

	if strings.HasPrefix(idString, "%") {
		id, err := strconv.Atoi(strings.TrimLeft(idString, "%"))
		if err != nil {
			return 0, err
		}
		value = int64(acl.ClusterUsr) + int64(id)
	}

	return value, nil
}

func aclParseUsers(users string) (string, error) {
	value, err := aclCalculateIDs(users)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%X", value), err
}

func aclParseResources(resources string) (string, error) {
	var ret int64
	resourceParts := strings.Split(resources, "/")
	if len(resourceParts) != 2 {
		return "", fmt.Errorf("Resource '%+v' malformed", resources)
	}

	res := strings.Split(resourceParts[0], "+")
	for _, resource := range res {
		val, ok := resourceMap[strings.ToUpper(resource)]
		if !ok {
			return "", fmt.Errorf("Resource '%+v' does not exist.", resource)
		}
		ret += int64(val)
	}
	ids, err := aclCalculateIDs(resourceParts[1])
	if err != nil {
		return "", err
	}
	ret += ids

	return fmt.Sprintf("%x", ret), nil
}

func aclParseRights(rights string) (string, error) {
	var ret int64

	rightsParts := strings.Split(rights, "+")
	for _, right := range rightsParts {
		val, ok := rightMap[strings.ToUpper(right)]
		if !ok {
			return "", fmt.Errorf("Right '%+v' does not exist.", right)
		}
		ret += int64(val)
	}

	return fmt.Sprintf("%x", ret), nil
}

func aclParseZone(zone string) (string, error) {
	ids, err := aclCalculateIDs(zone)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", ids), nil
}

func resourceOpennebulaACLCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	userHex, err := aclParseUsers(d.Get("user").(string))
	if err != nil {
		return err
	}

	resourceHex, err := aclParseResources(d.Get("resource").(string))
	if err != nil {
		return err
	}

	rightsHex, err := aclParseRights(d.Get("rights").(string))
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
