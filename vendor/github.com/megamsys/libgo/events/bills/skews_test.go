package bills

import (
	"gopkg.in/check.v1"
)

// func (s *S) TestNewSkews(c *check.C) {
//   email := "hello@virtengine.com"
//   cat_id := "ASM5552431390436283932"
//   _, err := NewEventsSkews(email, cat_id, s.m)
// 	c.Assert(err, check.IsNil)
// }
//
// func (s *S) TestCreateEvent(c *check.C) {
//   	o := &BillOpts{
//       AccountId: "hello@virtengine.com",
//   		AssemblyId: "ASM7862622526115144252",
//       AssemblyName: "long-flower-7084.atominstance.com",
//       AssembliesId: "AMS4821400633983299225",
//       SoftGracePeriod: "60h0m0s",
//       HardGracePeriod: "120h0m0s",
//       SoftLimit: "-1",
//       HardLimit: "-10",
//       SkewsType: "vm.ondemand.bills",
//     }
//     sk := &EventsSkews{
//       AccountId: o.AccountId,
//       CatId:     o.AssemblyId,
//       EventType: o.SkewsType,
//     }
//     err := sk.ActionEvents(o, "-2", s.m)
//     c.Assert(nil, check.NotNil)
// }

func (s *S) TestUpdateEvent(c *check.C) {
	// email := "vino.v@megam.io"
	// cat_id := "ASM7805386297957110462"
	// res, err := NewEventsSkews(email, cat_id, s.m)
	// err = res[0].update(s.m)
	// c.Assert(err, check.IsNil)
}
