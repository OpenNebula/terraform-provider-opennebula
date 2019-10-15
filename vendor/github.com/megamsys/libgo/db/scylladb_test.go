package db

import (
	//	"fmt"
	//  "github.com/megamsys/gocassa"
	"gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct {
	sy *ScyllaDB
}

type Customer struct {
	Id   string `json:"id" cql:"id"`
	Name string `json:"name" cql:"name"`
}

type Customers struct {
	Customers []*Customers
}

var _ = check.Suite(&S{})

/*
func (s *S) TestGetRecord(c *check.C) {
	cus := &Customer{}

	ops := Options{
		TableName:   "customers",
		Pks:         []string{"id"},
		Ccms:        []string{},
    Hosts:       []string{"192.168.0.108"},
    Keyspace:    "megdc",
		PksClauses:  map[string]interface{}{"id": "1"},
		CcmsClauses: map[string]interface{}{},
	}
		err := Fetchdb(ops, cus)
	c.Assert(err, check.NotNil)
}

func (s *S) TestGetListOfRecords(c *check.C) {
	cus := &[]Customer{}
  cu := &Customer{}
	ops := Options{
		TableName:   "customers",
		Pks:         []string{},
		Ccms:        []string{},
    Hosts:       []string{"192.168.0.108"},
    Keyspace:    "megdc",
		PksClauses:  map[string]interface{}{},
		CcmsClauses: map[string]interface{}{},
	}
		err := FetchListdb(ops,20, cu,cus)
	c.Assert(err, check.NotNil)
}*/

/*
func (s *S) TestInsterRecords(c *check.C) {
cus := &Customer{
  Id: "4",
  Name: "raj",
}

ops := Options{
	TableName:   "customers",
	Pks:         []string{"id"},
	Ccms:        []string{"},
	Hosts:       []string{"192.168.0.108"},
	Keyspace:    "megdc",
	PksClauses:  map[string]interface{}{"id": cus.Id},
	CcmsClauses: map[string]interface{}{},
}
	err := Storedb(ops, cus)
c.Assert(err, check.NotNil )
}
*/
