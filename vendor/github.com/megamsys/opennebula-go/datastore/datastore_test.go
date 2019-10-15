package datastore

//
// import (
// 	"github.com/megamsys/opennebula-go/api"
// 	"gopkg.in/check.v1"
// 	"testing"
// 	"fmt"
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
//
// func (s *S) TestGetDATASTOREs(c *check.C) {
// 	client, _ := api.NewClient(s.cm)
// 	vm := DatastoreTemplate{T: client}
//   fmt.Printf("%#v",vm)
// 	oja, err := vm.GetDATAs(2)
// 	fmt.Println(oja)
// 	err = nil
// 	c.Assert(err, check.NotNil)
// }

// func (s *S) TestGetDATASTOREALLs(c *check.C) {
// 	client, _ := api.NewClient(s.cm)
// 	vm := DatastoreTemplate{T: client}
//   fmt.Printf("%#v",vm)
// 	oja, err := vm.GetALL()
// 	fmt.Println(oja)
// 	err = nil
// 	c.Assert(err, check.NotNil)
// }

// func (s *S) TestDatastoreAllocate(c *check.C) {
// 	cl, _ := api.NewClient(s.cm)
// 	t := Datastore{
// 		Name:        "lvm",
// 		Ds_mad:      "fs",
// 		Tm_mad:      "fs_lvm",
// 		Disk_type:   "block",
// 		Bridge_list: "192.168.1.103",
// 		Type:        "image_ds",
// 		Safe_dirs:   "/var/tmp /var/lib/megam/images",
// 		Host:        "192.168.1.103",
// 		Vg_name:     "vg-one-0",
// 	}
// 	v := DatastoreTemplate{T: cl, Template: t}
//
// 	c.Assert(v, check.NotNil)
// 	_, err := v.AllocateDatastore(-1)
//
// 	err = nil
// 	c.Assert(err, check.NotNil)
// }
