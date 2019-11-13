package compute

import (
	"github.com/megamsys/opennebula-go/api"
	"gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) {
	check.TestingT(t)
}

type S struct {
	cm map[string]string
}

var _ = check.Suite(&S{})

func (s *S) SetUpSuite(c *check.C) {
	cm := make(map[string]string)
	cm[api.ENDPOINT] = "http://192.168.0.118:2633/RPC2"
	cm[api.USERID] = "oneadmin"
	cm[api.PASSWORD] = "oneadmin"
	s.cm = cm
}

/*
func (s *S) TestCreate(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	v := VirtualMachine {
		Name: "testmegam4",
		TemplateName: "megam",
		Cpu: "1",
		Memory: "1024",
		Image: "megam",
		ClusterId: "100" ,
		T: cl,
		ContextMap: map[string]string{"assembly_id": "ASM-007", "assemblies_id": "AMS-007", ACCOUNTS_ID: "info@megam.io"},
		Vnets:map[string]string{"0":"ipv4-pub"},
		} //memory in terms of MB! duh!

	c.Assert(v, check.NotNil)
  res, err := v.Create()
	fmt.Println("res  :",res)
	fmt.Println(err)
	err = nil
	c.Assert(err, check.NotNil)
}
*/

// func (s *S) TestCreateWithOldIP(c *check.C) {
// 	cl, _ := api.NewClient(s.cm)
// 	v := VirtualMachine {
// 		Name: "testmegam4",
// 		TemplateName: "megam",
// 		Cpu: "1",
// 		Memory: "1024",
// 		Image: "megam",
// 		ClusterId: "100" ,
// 		T: cl,
// 		ContextMap: map[string]string{"assembly_id": "ASM-007", "assemblies_id": "AMS-007", ACCOUNTS_ID: "info@megam.io"},
// 		Vnets:map[string]string{"0":"ipv4-pub"},
// 		} //memory in terms of MB! duh!
//
// 	c.Assert(v, check.NotNil)
// 	cm, err := v.Compute()
//   res, err := v.Create(cm)
// 	fmt.Println("res  :",res)
// 	fmt.Println(err)
// 	err = nil
// 	c.Assert(err, check.NotNil)
// }
/*

func (s *S) TestReboot(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	v := VirtualMachine{Name: "testrj", T: cl}
	c.Assert(v, check.NotNil)
	_, err := v.Reboot()
	c.Assert(err, check.NotNil)
}

func (s *S) TestResume(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	v := VirtualMachine{Name: "test", T: cl}
	c.Assert(v, check.NotNil)
	_, err := v.Resume()
	c.Assert(err, check.IsNil)
}

func (s *S) TestPoweroff(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	vmObj := VirtualMachine{Name: "test", T: cl}
	c.Assert(vmObj, check.NotNil)
	_, err := vmObj.Poweroff()
	c.Assert(err, check.IsNil)
}

func (s *S) TestPoweroffKVM(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	vmObj := VirtualMachine{Name: "kvm106", T: cl}
	c.Assert(vmObj, check.NotNil)
	_, err := vmObj.Resume()
	c.Assert(err, check.IsNil)
}

func (s *S) TestDelete(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	v := VirtualMachine{VMId: 77, T: cl}
	c.Assert(v, check.NotNil)
	_, err := v.Delete()
	c.Assert(nil, check.NotNil)
}
/*
func (s *S) TestDiskSnap(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	v := VirtualMachine{Name: "rj",T: cl}
	c.Assert(v, check.NotNil)
	_, err := v.DiskSnap()
	c.Assert(err, check.IsNil)
}


*/
