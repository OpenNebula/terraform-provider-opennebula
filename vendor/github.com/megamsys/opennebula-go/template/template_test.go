package template

// import (
// 	"testing"
//
// 	"github.com/megamsys/opennebula-go/api"
//
// 	"fmt"
// 	"gopkg.in/check.v1"
// )
//
// func Test(t *testing.T) {
// 	check.TestingT(t)
// }
//
// type S struct {
// 	cm map[string]string
// }
//
// var _ = check.Suite(&S{})
//
// func (s *S) SetUpSuite(c *check.C) {
// 	cm := make(map[string]string)
// 	cm[api.ENDPOINT] = "http://192.168.0.117:2633/RPC2"
// 	cm[api.USERID] = "oneadmin"
// 	cm[api.PASSWORD] = "dyovAupAuck9"
// 	s.cm = cm
// }

// func (s *S) TestGet(c *check.C) {
// 	client, _ := api.NewClient(s.cm)
// 	flav := TemplateReqs{TemplateName: "newone", T: client}
// 	res, error := flav.Get()
// 	c.Assert(error, check.IsNil)
// 	c.Assert(res, check.NotNil)
// }

// func (s *S) TestDatastoreAllocate(c *check.C) {
// 	cl, _ := api.NewClient(s.cm)
// 	user_template_essence := UserTemplates{}
// 	template_essence := Template{}
// 	tm := &NIC{
// 		Network:      "lvm",
// 		Network_uname: "lvm",
// 		}
// 	template_essence.Nic = append(template_essence.Nic,tm)
// 	utm := &UserTemplate{
// 			Id:    1,
// 			Uid:   1,
// 			Gid:   1,
// 			Uname: "h",
// 			Gname: "h",
// 			Name:  "h",
// 			Permissions: &Permissions{
// 				Owner_U: 1,
// 				Owner_M: 1,
// 				Owner_A: 1,
// 				Group_U: 1,
// 				Group_M: 1,
// 				Group_A: 1,
// 				Other_U: 1,
// 				Other_M: 1,
// 				Other_A: 1,
// 			},
// 			Template: &Template{
// 				Context: &Context{
// 					Network:        "lvm",
// 					Files:          "lvm",
// 					SSH_Public_key: "lvm",
// 					Set_Hostname:   "lvm",
// 					Node_Name:      "lvm",
// 					Accounts_id:    "lvm",
// 					Platform_id:    "lvm",
// 					Assembly_id:    "lvm",
// 					Assemblies_id:  "lvm",
// 				},
// 				Cpu:                      "lvm",
// 				Cpu_cost:                 "lvm",
// 				Description:              "lvm",
// 				Hypervisor:               "lvm",
// 				Logo:                     "lvm",
// 				Memory:                   "lvm",
// 				Memory_cost:              "lvm",
// 				Sunstone_capacity_select: "lvm",
// 				Sunstone_Network_select:  "lvm",
// 				VCpu: "lvm",
// 				Graphics: &Graphics{
// 					Listen:       "lvm",
// 					Port:         "lvm",
// 					Type:         "lvm",
// 					RandomPassWD: "lvm",
// 				},
// 				Disk: &Disk{
// 					Driver:      "lvm",
// 					Image:       "lvm",
// 					Image_Uname: "lvm",
// 					Size:        "lvm",
// 				},
// 				From_app:      "lvm",
// 				From_app_name: "lvm",
// 				Nic: template_essence.Nic,
// 					// Network:      "lvm",
// 					// Network_uname: "lvm",
//
// 				Os: &OS{
// 					Arch: "lvm",
// 				},
// 				Sched_requirments        :"lvm",
// 				Sched_ds_requirments     :"lvm",
// 			},
// 		}
// 		utma:= &Template{
// 			Context: &Context{
// 				Network:        "lvm",
// 				Files:          "lvm",
// 				SSH_Public_key: "lvm",
// 				Set_Hostname:   "lvm",
// 				Node_Name:      "lvm",
// 				Accounts_id:    "lvm",
// 				Platform_id:    "lvm",
// 				Assembly_id:    "lvm",
// 				Assemblies_id:  "lvm",
// 			},
// 			Cpu:                      "lvm",
// 			Cpu_cost:                 "lvm",
// 			Description:              "lvm",
// 			Hypervisor:               "lvm",
// 			Logo:                     "lvm",
// 			Memory:                   "lvm",
// 			Memory_cost:              "lvm",
// 			Sunstone_capacity_select: "lvm",
// 			Sunstone_Network_select:  "lvm",
// 			VCpu: "lvm",
// 			Graphics: &Graphics{
// 				Listen:       "lvm",
// 				Port:         "lvm",
// 				Type:         "lvm",
// 				RandomPassWD: "lvm",
// 			},
// 			Disk: &Disk{
// 				Driver:      "lvm",
// 				Image:       "lvm",
// 				Image_Uname: "lvm",
// 				Size:        "lvm",
// 			},
// 			From_app:      "lvm",
// 			From_app_name: "lvm",
// 			Nic: template_essence.Nic,
// 				// Network:      "lvm",
// 				// Network_uname: "lvm",
//
// 			Os: &OS{
// 				Arch: "lvm",
// 			},
// 			Sched_requirments        :"lvm",
// 			Sched_ds_requirments     :"lvm",
// 		}
// 		user_template_essence.UserTemplate.Template = append(user_template_essence.UserTemplate.Template,utma)
// 		// t := UserTemplates{
// 		// 	UserTemplate: user_template_essence.UserTemplate,
// 		// }
//
// 		// NAME: "lll",
// 		// Network :       "lvm",
// 		// Files   :       "lvm",
// 		// SSH_Public_key: "lvm",
// 		// Set_Hostname:   "lvm",
// 		// Node_Name   :   "lvm",
// 		// Accounts_id :   "lvm",
// 		// Platform_id :   "lvm",
// 		// Assembly_id :   "lvm",
// 		// Assemblies_id :"lvm",
//
//
// 	v := UserTemplate{T: cl, Template: user_template_essence.UserTemplate.Template}
//
// 	// c.Assert(v, check.NotNil)
// 	oja, err := v.AllocateTemplate()
// 	fmt.Println(oja)
// 	err = nil
// 	c.Assert(err, check.NotNil)
// }

// func (s *S) TestDatastoreAllocate(c *check.C) {
// 	cl, _ := api.NewClient(s.cm)
// 		template_essence := Template{}
// 		tm := &NIC{
// 			Network:      "lvm",
// 			Network_uname: "lvm",
// 			}
// 		template_essence.Nic = append(template_essence.Nic,tm)
// 	t := &Template{
// 		Name: "Rathish",
// 		Context: &Context{
// 						Network:        "lvm",
// 						Files:          "lvm",
// 						SSH_Public_key: "lvm",
// 						Set_Hostname:   "lvm",
// 						Node_Name:      "lvm",
// 						Accounts_id:    "lvm",
// 						Platform_id:    "lvm",
// 						Assembly_id:    "lvm",
// 						Assemblies_id:  "lvm",
// 					},
// 					Cpu:                      "lvm",
// 					Cpu_cost:                 "lvm",
// 					Description:              "lvm",
// 					Hypervisor:               "lvm",
// 					Logo:                     "lvm",
// 					Memory:                   "lvm",
// 					Memory_cost:              "lvm",
// 					Sunstone_capacity_select: "lvm",
// 					Sunstone_Network_select:  "lvm",
// 					VCpu: "lvm",
// 					Graphics: &Graphics{
// 						Listen:       "lvm",
// 						Port:         "lvm",
// 						Type:         "lvm",
// 						RandomPassWD: "lvm",
// 					},
// 					Disk: &Disk{
// 						Driver:      "lvm",
// 						Image:       "lvm",
// 						Image_Uname: "lvm",
// 						Size:        "lvm",
// 					},
// 					From_app:      "lvm",
// 					From_app_name: "lvm",
// 					Nic: template_essence.Nic,
//
// 					Os: &OS{
// 						Arch: "lvm",
// 					},
// 					Sched_requirments        :"lvm",
// 					Sched_ds_requirments     :"lvm",
//
// 	}
// 	v := UserTemplate{T: cl, Template: t}
// 	oja, err := v.AllocateTemplate()
// 		fmt.Println(oja)
// 		err = nil
// 		c.Assert(err, check.NotNil)
// }
