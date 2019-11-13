package vnet

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
	cm[api.ENDPOINT] = "http://192.168.0.100:2666/RPC2"
	cm[api.USERID] = "oneadmin"
	cm[api.PASSWORD] = "asdf"
	s.cm = cm
}

/*
func (s *S) TestGetVnetInfos(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	vm := VNETemplate{T: cl}
	_, err := vm.VnetInfos([]int{0})
	// for _, addr := range res[0].AddrPool.Addrs {
	// 	for _, leases := range addr.Leases {
	//     for i, lease := range leases.Leases {
	//       fmt.Printf("\n\n %v  %#v     ",i,lease)
	//      }
	// 		}
	// 	}
	c.Assert(err, check.NotNil)
}

func (s *S) TestVnethold(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	vm := VNETemplate{T: cl}
	res, err := vm.VnetHold(0,"192.168.0.100")
	fmt.Printf("\n\n %v  %#v     ",res,err)
	c.Assert(nil, check.NotNil)
}

func (s *S) TestVnetRelease(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	vm := VNETemplate{T: cl}
	res, err := vm.VnetRelease(0,"192.168.0.100")
	fmt.Printf("\n\nrelease %v  %#v     ",res,err)
	c.Assert(nil, check.NotNil)
}

/*
func (s *S) TestVnetCreate(c *check.C) {
	cl, _ := api.NewClient(s.cm)
  temp := Vnet{}
  ar := &Address{
      Type: "IP4",
      Size: "1",
      StartIP: "192.168.1.128",
    }
  temp.Addrs = append(temp.Addrs,ar)
  t := Vnet{
    Name: "vnet2",
    Type: "fixed",
    Description: "vnet for iPV4 ",
    Bridge: "one",
    Network_addr: "10.0.0.0",
    Network_mask: "255.255.255.0",
    Dns: "10.0.0.1",
    Gateway: "10.0.0.1",
    Vn_mad: "dummy",
    Addrs: temp.Addrs,
  }
	v := VNETemplate{T: cl, Template: t}

	c.Assert(v, check.NotNil)
	res, err := v.CreateVnet(-1)
	fmt.Println(res)
	err = nil
	c.Assert(err, check.NotNil)
}
*/
// func (s *S) TestGetVNets(c *check.C) {
// 	client, _ := api.NewClient(s.cm)
// 	vm := VNETemplate{T: client}
// 	_, err := vm.VnetInfos(2)
//   err = nil
// 	c.Assert(err, check.NotNil)
// }

// func (s *S) TestListVNets(c *check.C) {
// 	client, _ := api.NewClient(s.cm)
// 	vm := VNetPool{T: client}
//    err := vm.VnetPoolInfos(-1)
// 	 c.Assert(err, check.IsNil)
// 	 for _, i := range vm.Vnets {
// 		fmt.Println(i.Name, "  =    " , i.TotalIps)
// 	 }
//   err = fmt.Errorf("test")
// 	c.Assert(err, check.IsNil)
// }

// func (s *S) TestVnetAddIp(c *check.C) {
// 	cl, _ := api.NewClient(s.cm)
//   temp := Vnet{}
//   ar := &Address{
//       Type: "IP4",
//       Size: "1",
//       StartIP: "192.168.1.104",
//     }
//   var i int = 0
//   temp.Addrs = append(temp.Addrs,ar)
//   t := Vnet{
//     Id:  i,
//     Addrs: temp.Addrs,
//   }
//   v := VNETemplate{T: cl, Template: t}
//
//   c.Assert(v, check.NotNil)
//   res, err := v.VnetAddIps()
//   c.Assert(err, check.IsNil)
// }
// */
