package api

/*
import (
  // "net/http"
   //"io/ioutil"
  	"gopkg.in/check.v1"
)

func (s *S) testGet(path string) ([]byte, error) {
  cl := NewClient(s.ApiArgs, path)
  return cl.Get()
}

func (s *S) testPost(path string, item interface{}) ([]byte, error) {
  cl := NewClient(s.ApiArgs, path)
  return cl.Post(item)
}

/*
func (s *S) TestGetUser(c *check.C) {
  response, err := s.testGet("/accounts/" + s.ApiArgs.Email)
  c.Assert(err, check.IsNil)
}

type Assembly struct {
	AccountId string `json:"accounts_id"  cql:"accounts_id"`
	OrgId   string `json:"org_id" cql:"org_id"`
	Id      string `json:"id"  cql:"id"`
  Status  string `json:"status" cql:"status"`
}

type Components struct {
	Id      string `json:"id"  cql:"id"`
  Status  string `json:"status" cql:"status"`
}
//
func (s *S) TestGetAssembly(c *check.C) {
  htmlData, err := s.testGet("/assembly/ASM5195889316410516789")
  c.Assert(err, check.IsNil)
}
/*
func (s *S) TestAssemblyPost(c *check.C) {
  response, err := s.testGet("/assembly/ASM5285833184590940525")
  c.Assert(err, check.IsNil)
  response, err = s.testPost("/assembly/update", Assembly{AccountId: 	s.ApiArgs.Email, OrgId: s.ApiArgs.Org_Id , Id: "ASM5285833184590940525",Status:"testing"})
  c.Assert(err, check.IsNil)
}

func (s *S) TestComponentPost(c *check.C) {
  response, err := s.testGet("/components/CMP5285833184590940525")
  c.Assert(err, check.IsNil)
  response, err = s.testPost("/components/update", Components{Id: "CMP5285833184590940525",Status:"testing"})
  c.Assert(err, check.IsNil)
}
//


// type Sensor struct {
// 	Id                   string  `json:"id" cql:"id"`
// 	AccountId            string  `json:"account_id" cql:"account_id"`
// 	SensorType           string  `json:"sensor_type" cql:"sensor_type"`
// 	AssemblyId           string  `json:"assembly_id" cql:"assembly_id"`
// }
//
//
// func (s *S) TestSensorPost(c *check.C) {
//   response, err := s.testPost("/sensors/content", Sensor{Id: "SNS5285833184590940525",AccountId:"info@megam.io",SensorType: "VM",AssemblyId: "ASM000001" })
//   c.Assert(err, check.IsNil)
// }

// func (s *S) TestGetBalances(c *check.C) {
//   s.ApiArgs.Email = "vijaykanthm28@gmail.com"
//   response, err := s.testGet("/balances/vijaykanthm28@gmail.com")
//   c.Assert(err, check.IsNil)
// }

func (s *S) TestGetAddons(c *check.C) {
  s.ApiArgs.Email = "karthik@gmail.com"
  htmlData, err := s.testGet("/addons/WHMCS")
  c.Assert(err, check.IsNil)
}
*/
