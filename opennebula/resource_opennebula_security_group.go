package opennebula

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/securitygroup"
	sgk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/securitygroup/keys"
)

func resourceOpennebulaSecurityGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpennebulaSecurityGroupCreate,
		Read:   resourceOpennebulaSecurityGroupRead,
		Exists: resourceOpennebulaSecurityGroupExists,
		Update: resourceOpennebulaSecurityGroupUpdate,
		Delete: resourceOpennebulaSecurityGroupDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the Security Group",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Description of the Security Group Rule Set",
			},
			"permissions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Permissions for the Security Group (in Unix format, owner-group-other, use-manage-admin)",
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
				Description: "ID of the user that will own the Security Group",
			},
			"gid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group that will own the Security Group",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the Security Group",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the Security Group",
			},
			"rule": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				Description: "List of rules to be in the Security Group",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"protocol": {
							Type:        schema.TypeString,
							Description: "Protocol for the rule, must be one of: ALL, TCP, UDP, ICMP or IPSEC",
							Required:    true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								validprotos := []string{"ALL", "TCP", "UDP", "ICMP", "IPSEC"}
								value := v.(string)

								if inArray(value, validprotos) < 0 {
									errors = append(errors, fmt.Errorf("Protocol %q must be one of: %s", k, strings.Join(validprotos, ",")))
								}

								return
							},
						},
						"rule_type": {
							Type:        schema.TypeString,
							Description: "Direction of the traffic flow to allow, must be INBOUND or OUTBOUND",
							Required:    true,
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								validtypes := []string{"INBOUND", "OUTBOUND"}
								value := v.(string)

								if inArray(value, validtypes) < 0 {
									errors = append(errors, fmt.Errorf("Rule type %q must be one of: %s", k, strings.Join(validtypes, ",")))
								}

								return
							},
						},
						"ip": {
							Type:        schema.TypeString,
							Description: "IP (or starting IP if used with 'size') to apply the rule to",
							Optional:    true,
							Computed:    true,
						},
						"size": {
							Type:        schema.TypeString,
							Description: "Number of IPs to apply the rule from, starting with 'ip'",
							Optional:    true,
							Computed:    true,
						},
						"range": {
							Type:        schema.TypeString,
							Description: "Comma separated list of ports and port ranges",
							Optional:    true,
							Computed:    true,
						},
						"icmp_type": {
							Type:        schema.TypeString,
							Description: "Type of ICMP traffic to apply to when 'protocol' is ICMP",
							Optional:    true,
							Computed:    true,
						},
						"network_id": {
							Type:        schema.TypeString,
							Description: "VNET ID to be used as the source/destination IP addresses",
							Optional:    true,
							Computed:    true,
						},
					},
				},
			},
			"commit": {
				Type:        schema.TypeBool,
				Description: "Should changes to the Security Group rules be commited to running Virtual Machines?",
				Optional:    true,
				Default:     true,
			},
			"group": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"gid"},
				Description:   "Name of the Group that onws the Security Group, If empty, it uses caller group",
			},
			"tags": tagsSchema(),
		},
	}
}

func getSecurityGroupController(d *schema.ResourceData, meta interface{}, args ...int) (*goca.SecurityGroupController, error) {
	controller := meta.(*goca.Controller)
	var sgc *goca.SecurityGroupController

	// Try to find the Security Group by ID, if specified
	if d.Id() != "" {
		gid, err := strconv.ParseUint(d.Id(), 10, 64)
		if err != nil {
			return nil, err
		}
		sgc = controller.SecurityGroup(int(gid))
	}

	// Otherwise, try to find the security Group by name as the de facto compound primary key
	if d.Id() == "" {
		gid, err := controller.SecurityGroups().ByName(d.Get("name").(string), args...)
		if err != nil {
			return nil, err
		}
		sgc = controller.SecurityGroup(gid)
	}

	return sgc, nil
}

func changeSecurityGroupGroup(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)
	var gid int

	sgc, err := getSecurityGroupController(d, meta)
	if err != nil {
		return err
	}

	if d.Get("group") != "" {
		group := d.Get("group").(string)
		gid, err = controller.Groups().ByName(group)
		if err != nil {
			return fmt.Errorf("Can't find a group with name `%s`: %s", group, err)
		}
	} else {
		gid = d.Get("gid").(int)
	}

	err = sgc.Chown(-1, gid)
	if err != nil {
		return fmt.Errorf("Can't find a group with ID `%d`: %s", gid, err)
	}

	return nil
}

func resourceOpennebulaSecurityGroupRead(d *schema.ResourceData, meta interface{}) error {
	// Get all Security Group
	sgc, err := getSecurityGroupController(d, meta, -2, -1, -1)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing security group %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}
		return err
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	securitygroup, err := sgc.Info(false)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", securitygroup.ID))
	d.Set("uid", securitygroup.UID)
	d.Set("gid", securitygroup.GID)
	d.Set("uname", securitygroup.UName)
	d.Set("gname", securitygroup.GName)
	d.Set("permissions", permissionsUnixString(*securitygroup.Permissions))

	description, _ := securitygroup.Template.Get(sgk.Description)
	d.Set("description", description)

	if err := d.Set("rule", generateSecurityGroupMapFromStructs(securitygroup.Template.GetRules())); err != nil {
		log.Printf("[WARN] Error setting rule for Security Group %x, error: %s", securitygroup.ID, err)
	}

	err = flattenSecurityGroupTags(d, &securitygroup.Template)
	if err != nil {
		return err
	}

	return nil
}

func flattenSecurityGroupTags(d *schema.ResourceData, sgTpl *securitygroup.Template) error {

	tags := make(map[string]interface{})
	var err error
	// Get only tags from userTemplate
	if tagsInterface, ok := d.GetOk("tags"); ok {
		for k, _ := range tagsInterface.(map[string]interface{}) {
			tags[k], err = sgTpl.GetStr(strings.ToUpper(k))
			if err != nil {
				return err
			}
		}
	}

	if len(tags) > 0 {
		err := d.Set("tags", tags)
		if err != nil {
			return err
		}
	}

	return nil
}

func generateSecurityGroupMapFromStructs(rulesVectors []securitygroup.Rule) []map[string]interface{} {

	rules := make([]map[string]interface{}, 0, len(rulesVectors))
	for _, ruleVec := range rulesVectors {
		rule := make(map[string]interface{}, len(ruleVec.Pairs))
		for _, pair := range ruleVec.Pairs {
			rule[strings.ToLower(pair.Key())] = pair.Value
		}
		rules = append(rules, rule)
	}

	return rules
}

func resourceOpennebulaSecurityGroupExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceOpennebulaSecurityGroupRead(d, meta)
	if err != nil || d.Id() == "" {
		return false, err
	}

	return true, nil
}

func resourceOpennebulaSecurityGroupCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)

	secGroupDef, err := generateSecurityGroup(d)
	if err != nil {
		return err
	}

	secGroupID, err := controller.SecurityGroups().Create(secGroupDef)
	if err != nil {
		log.Printf("[ERROR] Security group creation failed, error: %s", err)
		return err
	}

	sgc := controller.SecurityGroup(secGroupID)

	secGroupTpl := generateSecurityGroupTemplate(d)

	// add template information into Security group
	err = sgc.Update(secGroupTpl, 1)
	if err != nil {
		return err
	}

	d.SetId(fmt.Sprintf("%v", secGroupID))

	// Change Permissions
	if perms, ok := d.GetOk("permissions"); ok {
		err = sgc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			log.Printf("[ERROR] Security group permissions change failed, error: %s", err)
			return err
		}
	}

	if d.Get("group") != "" || d.Get("gid") != "" {
		err = changeSecurityGroupGroup(d, meta)
		if err != nil {
			log.Printf("[ERROR] Security group owner group change failed, error: %s", err)
			return err
		}
	}

	return resourceOpennebulaSecurityGroupRead(d, meta)
}

func resourceOpennebulaSecurityGroupUpdate(d *schema.ResourceData, meta interface{}) error {

	//Get Security Group
	sgc, err := getSecurityGroupController(d, meta)
	if err != nil {
		return err
	}
	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	securitygroup, err := sgc.Info(false)
	if err != nil {
		return err
	}
	tpl := &securitygroup.Template
	changes := false

	if d.HasChange("rule") && d.Get("rule") != "" {

		tpl.Del((string(sgk.RuleVec)))
		generateSecurityGroupRules(d, tpl)
		changes = true
	}

	if d.HasChange("tags") {
		tagsInterface := d.Get("tags").(map[string]interface{})
		for k, v := range tagsInterface {
			tpl.Del(strings.ToUpper(k))
			tpl.AddPair(strings.ToUpper(k), v)
		}
		changes = true
	}

	if changes {
		err = sgc.Update(tpl.String(), 0)
		if err != nil {
			return err
		}

		log.Printf("[INFO] Successfully updated Security Group template %s\n", securitygroup.Name)

		//Commit changes to running VMs if desired
		if d.Get("commit") == true {
			// Only update outdated VMs not all
			err = sgc.Commit(true)
			if err != nil {
				return err
			}

			log.Printf("[INFO] Successfully commited Security Group %s changes to outdated Virtual Machines\n", securitygroup.Name)
		}
	}

	if d.HasChange("name") {
		err := sgc.Rename(d.Get("name").(string))
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated name for SecurityGroup %s\n", securitygroup.Name)
	}

	if d.HasChange("permissions") {
		if perms, ok := d.GetOk("permissions"); ok {
			err = sgc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				return err
			}
		}
		log.Printf("[INFO] Successfully updated Permissions Security Group %s\n", securitygroup.Name)
	}

	if d.HasChange("group") || d.HasChange("gid") {
		err = changeSecurityGroupGroup(d, meta)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated group for Security Group %s\n", securitygroup.Name)
	}

	return resourceOpennebulaSecurityGroupRead(d, meta)
}

func resourceOpennebulaSecurityGroupDelete(d *schema.ResourceData, meta interface{}) error {
	sgc, err := getSecurityGroupController(d, meta)
	if err != nil {
		return err
	}

	err = sgc.Delete()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Successfully deleted Security Group ID %s\n", d.Id())
	return nil
}

func generateSecurityGroup(d *schema.ResourceData) (string, error) {
	secgroupname := d.Get("name").(string)

	tpl := securitygroup.NewTemplate()
	tpl.Add(sgk.Name, secgroupname)

	tplStr := tpl.String()
	log.Printf("[INFO] Security Group definition: %s", tplStr)

	return tplStr, nil
}

func generateSecurityGroupRules(d *schema.ResourceData, tpl *securitygroup.Template) {
	//Generate rules definition
	rules := d.Get("rule").([]interface{})
	log.Printf("Number of Security Group rules: %d", len(rules))

	for i := 0; i < len(rules); i++ {
		ruleconfig := rules[i].(map[string]interface{})
		rule := tpl.AddRule()

		for k, v := range ruleconfig {

			if isEmptyValue(reflect.ValueOf(v)) {
				continue
			}

			switch k {
			case "protocol":
				rule.Add(sgk.Protocol, v.(string))
			case "rule_type":
				rule.Add(sgk.RuleType, v.(string))
			case "ip":
				rule.Add(sgk.IP, v.(string))
			case "size":
				rule.Add(sgk.Size, v.(string))
			case "range":
				rule.Add(sgk.Range, v.(string))
			case "icmp_type":
				rule.Add(sgk.IcmpType, v.(string))
			case "network_id":
				rule.Add(sgk.NetworkID, v.(string))
			}
		}
	}

}

func generateSecurityGroupTemplate(d *schema.ResourceData) string {
	tpl := securitygroup.NewTemplate()

	generateSecurityGroupRules(d, tpl)

	description := d.Get("description").(string)
	if len(description) > 0 {
		tpl.Add(sgk.Description, description)
	}

	tagsInterface := d.Get("tags").(map[string]interface{})
	for k, v := range tagsInterface {
		tpl.AddPair(strings.ToUpper(k), v)
	}

	tplStr := tpl.String()
	log.Printf("[INFO] Security Group template: %s", tplStr)

	return tplStr

}
