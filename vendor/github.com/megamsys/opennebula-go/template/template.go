package template

import (
	"encoding/xml"
	"errors"
	"fmt"
	"github.com/megamsys/opennebula-go/api"
)

var (
	ErrNoTemplate = errors.New("no templates found, Did you have templates ?")
)

type UserTemplates struct {
	UserTemplate []*UserTemplate `xml:"VMTEMPLATE"`
}

type UserTemplate struct {
	T           *api.Rpc
	Id          int          `xml:"ID"`
	Uid         int          `xml:"UID"`
	Gid         int          `xml:"GID"`
	Uname       string       `xml:"UNAME"`
	Gname       string       `xml:"GNAME"`
	Name        string       `xml:"NAME"`
	Permissions *Permissions `xml:"PERMISSIONS"`
	Template    *Template    `xml:"TEMPLATE"`
	RegTime     int          `xml:"REGTIME"`
}

type Template struct {
	Name                     string      `xml:"NAME"`
	Context                  *Context    `xml:"CONTEXT"`
	Cpu                      string      `xml:"CPU"`
	Cpu_cost                 string      `xml:"CPU_COST"`
	Description              string      `xml:"DESCRIPTION"`
	Hypervisor               string      `xml:"HYPERVISOR"`
	Logo                     string      `xml:"LOGO"`
	Memory                   string      `xml:"MEMORY"`
	Memory_cost              string      `xml:"MEMORY_COST"`
	Sunstone_capacity_select string      `xml:"SUNSTONE_CAPACITY_SELECT"`
	Sunstone_Network_select  string      `xml:"SUNSTONE_NETWORK_SELECT"`
	VCpu                     string      `xml:"VCPU"`
	Graphics                 *Graphics   `xml:"GRAPHICS"`
	Disks                    []*Disk     `xml:"DISK"`
	Disk_cost                string      `xml:"DISK_COST"`
	From_app                 string      `xml:"FROM_APP"`
	From_app_name            string      `xml:"FROM_APP_NAME"`
	Nic                      []*NIC      `xml:"NIC"`
	Os                       *OS         `xml:"OS"`
	Sched_requirments        string      `xml:"SCHED_REQUIREMENTS"`
	Sched_ds_requirments     string      `xml:"SCHED_DS_REQUIREMENTS"`
	Vcenter_datastore        string      `xml:"VCENTER_DATASTORE,omitempty"`
	Public_cloud             PublicCloud `xml:"PUBLIC_CLOUD,omitempty"`
	KeepDiskOnDone           string      `xml:"KEEP_DISKS_ON_DONE,omitempty"`
}

type PublicCloud struct {
	Type       string `xml:"TYPE,omitempty"`
	VmTemplate string `xml:"VM_TEMPLATE,omitempty"`
}

type Graphics struct {
	Listen       string `xml:"LISTEN"`
	Port         string `xml:"PORT"`
	Type         string `xml:"TYPE"`
	RandomPassWD string `xml:"RANDOM_PASSWD"`
}

type OS struct {
	Arch string `xml:"ARCH"`
}

type NIC struct {
	Id            int    `xml:"-"`
	Network       string `xml:"NETWORK"`
	Network_uname string `xml:"NETWORK_UNAME"`
	IP            string `xml:"IP"`
}

type Context struct {
	Network        string `xml:"NETWORK"`
	Files          string `xml:"FILES"`
	SSH_Public_key string `xml:"SSH_PUBLIC_KEY"`
	Set_Hostname   string `xml:"SET_HOSTNAME"`
	Node_Name      string `xml:"NODE_NAME"`
	Accounts_id    string `xml:"ACCOUNTS_ID"`
	Platform_id    string `xml:"PLATFORM_ID"`
	Assembly_id    string `xml:"ASSEMBLY_ID"`
	Assemblies_id  string `xml:"ASSEMBLIES_ID"`
	Marketplace_id string `xml:"MARKETPLACE_ID"`
	ApiKey         string `xml:"API_KEY"`
	Org_id         string `xml:"ORG_ID"`
	Quota_id       string `xml:"QUOTA_ID"`
}

type Disk struct {
	Driver      string `xml:"DRIVER"`
	Image       string `xml:"IMAGE"`
	Image_Uname string `xml:"IMAGE_UNAME"`
	Size        string `xml:"SIZE"`
}

type Permissions struct {
	Owner_U int `xml:"OWNER_U"`
	Owner_M int `xml:"OWNER_M"`
	Owner_A int `xml:"OWNER_A"`
	Group_U int `xml:"GROUP_U"`
	Group_M int `xml:"GROUP_M"`
	Group_A int `xml:"GROUP_A"`
	Other_U int `xml:"OTHER_U"`
	Other_M int `xml:"OTHER_M"`
	Other_A int `xml:"OTHER_A"`
}

type TemplateReqs struct {
	TemplateName string
	TemplateId   int
	TemplateData string
	T            *api.Rpc
}

/**
 *
 * Given a templateId it would return the XML of that particular template
 *
 **/
func (t *TemplateReqs) GetTemplate() (interface{}, error) {
	args := []interface{}{t.T.Key, -2, t.TemplateId, t.TemplateId}
	res, err := t.T.Call(api.TEMPLATEPOOL_INFO, args)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (v *UserTemplate) AllocateTemplate() (interface{}, error) {
	finalXML := UserTemplate{}
	finalXML.Template = v.Template
	finalData, _ := xml.Marshal(finalXML.Template)
	data := string(finalData)
	args := []interface{}{v.T.Key, data}
	res, err := v.T.Call(api.ONE_TEMPLATE_ALLOCATE, args)
	if err != nil {
		return nil, err
	}
	return res, nil
}

/**
 *
 * Gets a particular template with a template name given
 *
 **/
func (t *TemplateReqs) Get() ([]*UserTemplate, error) {
	args := []interface{}{t.T.Key, -2, -1, -1}
	templatePool, err := t.T.Call(api.TEMPLATEPOOL_INFO, args)
	if err != nil {
		return nil, err
	}

	xmlStrt := UserTemplates{}
	if err = xml.Unmarshal([]byte(templatePool), &xmlStrt); err != nil {
		return nil, err
	}

	if len(xmlStrt.UserTemplate) < 1 {
		return nil, ErrNoTemplate
	}
	var matchedTemplate = make([]*UserTemplate, len(xmlStrt.UserTemplate))
	for _, v := range xmlStrt.UserTemplate {
		if v.Name == t.TemplateName {
			matchedTemplate[0] = v
		}
	}

	if matchedTemplate[0] == nil {
		return nil, fmt.Errorf("Unavailable templatename [" + t.TemplateName + "]")
	}
	fmt.Println(matchedTemplate)
	return matchedTemplate, nil
}

/**
 * template instentiate in OpenNebula
 **/

func (t *TemplateReqs) Instantiate(name string) (interface{}, error) {
	args := []interface{}{t.T.Key, t.TemplateId, name, false, t.TemplateData}
	return t.T.Call(api.TEMPLATE_INSTANTIATE, args)
}

/**
 *
 * Update a template in OpenNebula
 *
 **/
func (t *TemplateReqs) Update() error {
	args := []interface{}{t.T.Key, t.TemplateId, t.TemplateData, 0}
	if _, err := t.T.Call(api.TEMPLATE_UPDATE, args); err != nil {
		return err
	}
	return nil
}
