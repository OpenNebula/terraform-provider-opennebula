package images

import (
	"github.com/megamsys/opennebula-go/api"
	"gopkg.in/check.v1"
	"testing"
	//	"fmt"
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

// func (s *S) TestImageShow(c *check.C) {
// 	cl, _ := api.NewClient(s.cm)
//
// 	v := &Image{T: cl, Id: 19}
// res, err := v.Show()
//  fmt.Println(err,"Image State: ",res)
// 	c.Assert(nil, check.NotNil)
// }

/*
 func (s *S) TestImageList(c *check.C) {
 	cl, _ := api.NewClient(s.cm)

 	v := &Image{T: cl}

 	c.Assert(v, check.NotNil)
 res, err := v.ImageList()
  fmt.Println(res)
 	c.Assert(err, check.IsNil)
 }

 func (s *S) TestCreateISO(c *check.C)  {
  cl, _ := api.NewClient(s.cm)

  v := &Image{
 	 T: cl,
 	 Name: "NEW_TEST",
 	 Path: "http://archive.ubuntu.com/ubuntu/dists/xenial/main/installer-amd64/current/images/netboot/mini.iso",
 	 Type: CD_ROM,
 	 DatastoreID: 100,
  }
  c.Assert(v, check.NotNil)
   _, err := v.Create()
  fmt.Println("Error image create :",err)
  c.Assert(err, check.IsNil)
 }

 func (s *S) TestCreateDataBlock(c *check.C)  {
	cl, _ := api.NewClient(s.cm)

	v := &Image{
		T: cl,
		Name: "NEW_TEST 10GB",
		Size: 1024,
		Type: DATABLOCK,
		DatastoreID: 100,
	}
	c.Assert(v, check.NotNil)
	 _, err := v.Create()
	fmt.Println("Error image create :",err)
	c.Assert(err, check.IsNil)
 }

 func (s *S) TestCreateOS(c *check.C) {
	  cl, _ := api.NewClient(s.cm)

	  v := &Image{
	 	 T: cl,
	 	 Name: "NEW_TEST OS",
	 	 Path: "/var/lib/megam/images/ajenti.img",
	 	 Type: CD_ROM,
	 	 DatastoreID: 100,
	  }
	  c.Assert(v, check.NotNil)
	   _, err := v.Create()
	  fmt.Println("Error image create :",err)
	  c.Assert(err, check.IsNil)
 }


 func (s *S) TestDelete(c *check.C) {
	  cl, _ := api.NewClient(s.cm)

	  v := &Image{T: cl, Name: "NEW_TEST OS" }

	  c.Assert(v, check.NotNil)
	   _, err := v.Create()
	  fmt.Println("Error image create :",err)
	  c.Assert(err, check.IsNil)
 }
*/
