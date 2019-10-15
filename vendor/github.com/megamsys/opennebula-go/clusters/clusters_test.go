package clusters

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
	cm[api.ENDPOINT] = "http://192.168.0.117:2633/RPC2"
	cm[api.USERID] = "oneadmin"
	cm[api.PASSWORD] = "asdf"
	s.cm = cm
}

/*
func (s *S) TestClustersInfo(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	v := Clusters{T: cl} //memory in terms of MB! duh!

	c.Assert(v, check.NotNil)
	res, err := v.ClusterPoolinfo()
  xmlVM := &Clusters{}
  fmt.Println(res)
  assert := res[1].(string)
  if err = xml.Unmarshal([]byte(assert), xmlVM); err != nil {
     fmt.Println(err)
  }
  fmt.Printf("%#v",xmlVM.Clusters)

	c.Assert(err, check.NotNil)
}

//*/
/*
func (s *S) TestClusterInfo(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	v := Clusters{T: cl}

	c.Assert(v, check.NotNil)
	res, err := v.ClusterInfo(0)
  xmlCL := &Cluster{}
  fmt.Println(res)
  assert := res[1].(string)
  if err = xml.Unmarshal([]byte(assert), xmlCL); err != nil {
     fmt.Println(err)
  }
  fmt.Printf("%s",*xmlCL.Datastores.ID[0])

	c.Assert(err, check.NotNil)
}


func (s *S) TestClustersCreate(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	v := Clusters{T: cl}

	c.Assert(v, check.NotNil)
	res, err := v.CreateCluster("testcluster")
  fmt.Println(res)

	c.Assert(err, check.IsNil)
}
*/

/*
func (s *S) TestClusterAddResources(c *check.C) {
	cl, _ := api.NewClient(s.cm)
	v := Clusters{T: cl}

	c.Assert(v, check.NotNil)
  resource_id := 0
  //cluster_name := "testcluster"
	cluster_id := 102
	res, err := v.ClusterAddResources(CLUSTER_DELVNET,cluster_id,resource_id)
  xmlVM := &Clusters{}
  fmt.Println(res)

  fmt.Printf("%#v",xmlVM)

	c.Assert(err, check.IsNil)
}
//*/
