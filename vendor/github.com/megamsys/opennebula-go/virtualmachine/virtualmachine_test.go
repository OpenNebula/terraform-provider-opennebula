package virtualmachine

// import (
// 	"github.com/megamsys/opennebula-go/api"
// 	"gopkg.in/check.v1"
//   "fmt"
// )

// func (s *S) TestGetByName(c *check.C) {
// 	client, _ := api.NewClient(s.cm)
// 	vm := Query{VMName: "kvm109", T: client}
// 	_, err := vm.GetByName()
// 	c.Assert(err, check.IsNil)
// }

// func (s *S) TestGet(c *check.C) {
// 	client, _ := api.NewClient(s.cm)
// 	vm := Vnc{T: client, VmId: "8"}
// 	_, err := vm.GetVm()
// 	c.Assert(err, check.IsNil)
// }

//
// func (s *S) TestGetVMTemplate(c *check.C) {
// 	client, _ := api.NewClient(s.cm)
// 	vm := Vnc{VmId: "177", T: client}
// 	_, err := vm.GetVm()
// 	c.Assert(err, check.IsNil)
// }
//
// func (s *S) TestAttachNic(c *check.C) {
// 	client, _ := api.NewClient(s.cm)
//   vm := Vnc{VmId: "108", T: client}
//   err := vm.AttachNic("public-1", "192.168.0.102")
//   fmt.Println(err)
// 	c.Assert(nil, check.NotNil)
// }

// func (s *S) TestDetachNic(c *check.C) {
// 	client, _ := api.NewClient(s.cm)
//   vm := Vnc{VmId: "425", T: client}
// 	nic := 3
//   err := vm.DetachNic(nic)
// 	c.Assert(err, check.IsNil)
// }
