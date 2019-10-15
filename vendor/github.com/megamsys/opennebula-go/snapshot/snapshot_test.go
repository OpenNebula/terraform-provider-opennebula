package snapshot

// import (
// 	"github.com/megamsys/opennebula-go/api"
// 	"gopkg.in/check.v1"
// 	"testing"
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
// 	cm[api.ENDPOINT] = "http://192.168.0.100:2633/RPC2"
// 	cm[api.USERID] = "oneadmin"
// 	cm[api.PASSWORD] = "asdf"
// 	s.cm = cm
// }

// func (s *S) TestCreateSnapshot(c *check.C) {
// 	cl, _ := api.NewClient(s.cm)
// 	v := Snapshot{
// 		VMId:            87,
// 		DiskId:          0,
// 		DiskDiscription: "backy_test",
// 		T:               cl,
// 	}
// 	c.Assert(v, check.NotNil)
// 	_, err := v.CreateSnapshot()
// 	c.Assert(err, check.IsNil)
// }

// func (s *S) TestDeleteSnapshot(c *check.C) {
// 	cl, _ := api.NewClient(s.cm)
// 	v := Snapshot{
// 	 VMId: 333,
//   DiskId:         0,
// 		SnapId: 0,
//   T:     cl,
// }
// 	c.Assert(v, check.NotNil)
// 	res, err := v.DeleteSnapshot()
// 	fmt.Println(res)
// 	err = nil
// 	c.Assert(err, check.NotNil)
// }

// func (s *S) TestRevertSnapshot(c *check.C) {
// 	cl, _ := api.NewClient(s.cm)
// 	v := Snapshot{
// 	 VMId: 333,
//   DiskId:         0,
// 		SnapId: 0,
//   T:     cl,
// }
// 	c.Assert(v, check.NotNil)
// 	res, err := v.RevertSnapshot()
// 	fmt.Println(res)
// 	err = nil
// 	c.Assert(err, check.NotNil)
// }
