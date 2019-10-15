/*
** Copyright [2013-2017] [Megam Systems]
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
** http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
 */
package bills

import (
	"encoding/json"
	"github.com/megamsys/libgo/api"
)

const (
	ORGANIZATIONGET = "/organizations"
	ACCOUNTS_GET    = "/accounts/"
)

type ApiAccounts struct {
	JsonClaz string   `json:"json_claz" cql:"json_claz"`
	Results  Accounts `json:"results" cql:"results"`
}

type ApiOrganizations struct {
	JsonClaz string       `json:"json_claz" cql:"json_claz"`
	Results  Organization `json:"results" cql:"results"`
}

type Accounts struct {
	Id           string   `json:"id" cql:"id"`
	Name         Name     `json:"name" cql:"name"`
	Phone        Phone    `json:"phone" cql:"phone"`
	Email        string   `json:"email" cql:"email"`
	Dates        Dates    `json:"dates" cql:"dates"`
	ApiKey       string   `json:"api_key" cql:"api_key"`
	Password     Password `json:"password" cql:"password"`
	Approval     Approval `json:"approval" cql:"approval"`
	Suspend      Suspend  `json:"suspend" cql:"suspend"`
	RegIpAddress string   `json:"registration_ip_address" cql:"registration_ip_address"`
	States       States   `json:"states" cql:"states"`
}

type Organization struct {
	Id         string `json:"id" cql:"id"`
	AccountsId string `json:"accounts_id" cql:"accounts_id"`
	CreatedAt  string `json:"created_at" cql:"created_at"`
	JsonClaz   string `json:"json_claz" cql:"json_claz"`
	Name       string `json:"name" cql:"name"`
}

type Name struct {
	FirstName string `json:"first_name" cql:"first_name"`
	LastName  string `json:"last_name" cql:"last_name"`
}

type Password struct {
	Password            string `json:"password" cql:"password"`
	PasswordResetKey    string `json:"password_reset_key" cql:"password_reset_key"`
	PasswordResetSentAt string `json:"password_reset_sent_at" cql:"password_reset_sent_at"`
}

type Phone struct {
	Phone         string `json:"phone" cql:"phone"`
	PhoneVerified string `json:"phone_verified" cql:"phone_verified"`
}

type Approval struct {
	Approved     string `json:"approved" cql:"approved"`
	ApprovedById string `json:"approved_by_id" cql:"approved_by_id"`
	ApprovedAt   string `json:"approved_at" cql:"approved_at"`
}

type Suspend struct {
	Suspended     string `json:"suspended" cql:"suspended"`
	SuspendedAt   string `json:"suspended_at" cql:"suspended_at"`
	SuspendedTill string `json:"suspended_till" cql:"suspended_till"`
}

type Dates struct {
	CreatedAt       string `json:"created_at" cql:"created_at"`
	LastPostedAt    string `json:"last_posted_at" cql:"last_posted_at"`
	LastEmailedAt   string `json:"last_emailed_at" cql:"last_emailed_at"`
	PreviousVisitAt string `json:"previous_visit_at" cql:"previous_visit_at"`
	FirstSeenAt     string `json:"first_seen_at" cql:"first_seen_at"`
}

type States struct {
	Authority string `json:"authority" cql:"authority"`
	Active    string `json:"active" cql:"active"`
	Blocked   string `json:"blocked" cql:"blocked"`
	Staged    string `json:"staged" cql:"staged"`
}

func NewAccounts(m map[string]string) (*Accounts, error) {
	args := api.NewArgs(m)
	cl := api.NewClient(args, ACCOUNTS_GET+args.Email)
	response, err := cl.Get()
	if err != nil {
		return nil, err
	}

	ac := &ApiAccounts{}
	err = json.Unmarshal(response, ac)
	if err != nil {
		return nil, err
	}
	a := &ac.Results
	return a, nil
}

func AccountsOrg(email string, m map[string]string) (*Organization, error) {
	args := api.NewArgs(m)
	cl := api.NewClient(args, ORGANIZATIONGET)
	response, err := cl.Get()
	if err != nil {
		return nil, err
	}

	o := &ApiOrganizations{}
	err = json.Unmarshal(response, o)
	if err != nil {
		return nil, err
	}
	org := &o.Results
	return org, nil
}

// func (a *Accounts) convertBillAccount() (*BillAccounts, error) {
// 	b := &BillAccounts{}
// 	a.parseStringToStruct([]byte(a.Name), &b.Name)
// 	a.parseStringToStruct([]byte(a.Phone), &b.Phone)
// 	a.parseStringToStruct([]byte(a.Password), &b.Password)
// 	a.parseStringToStruct([]byte(a.Suspend), &b.Suspend)
// 	a.parseStringToStruct([]byte(a.Approval), &b.Approval)
// 	a.parseStringToStruct([]byte(a.States), &b.States)
// 	a.parseStringToStruct([]byte(a.Dates), &b.Dates)
// 	b.Id = a.Id
// 	b.Email = a.Email
// 	b.ApiKey = a.ApiKey
// 	b.RegIpAddress = a.RegIpAddress
// 	return b, nil
// }
