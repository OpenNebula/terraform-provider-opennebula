package users

import (
	"github.com/megamsys/opennebula-go/api"
)

type UserTemplate struct {
	Users User `xml:"USER"`
	T     *api.Rpc
}

type User struct {
	UserId     int    `json:"id" xml:"ID"`
	UserName   string `json:"username" xml:"USERNAME"`
	Password   string `json:"password" xml:"PASSWORD"`
	AuthDriver string `json:"auth_driver" xml:"AUTH_DRIVER"`
	GroupIds   []int  `json:"group_ids" xml:"GROUPIDS"`
}

type ID struct {
	Id string `json:"id" xml:"ID"`
}

func (u *UserTemplate) CreateUsers() (interface{}, error) {
	args := []interface{}{u.T.Key, u.Users.UserName, u.Users.Password, u.Users.AuthDriver, u.Users.GroupIds}
	res, err := u.T.Call(api.ONE_USER_CREATE, args)
	if err != nil {
		return nil, err
	}
	return res, nil
}
