package api

import (
	"gopkg.in/yaml.v2"
)

const (
	REQUEST_CREATE = "/requests/content"
)

type Requests struct {
	Name      string `json:"name" cql:"name"`
	AccountId string `json:"account_id" cql:"account_id"`
	CatId     string `json:"cat_id" cql:"cat_id"`
	Action    string `json:"action" cql:"action"`
	Category  string `json:"category" cql:"category"`
	CatType   string `json:"cattype" cql:"cattype"`
	//	CreatedAt time.Time `json:"created_at" cql:"created_at"`
}

// type ApiRequests struct {
// 	JsonClaz string     `json:"json_claz" cql:"json_claz"`
// 	Results  []Requests `json:"results" cql:"results"`
// }

func NewRequest(email string) *Requests {
	return &Requests{
		AccountId: email,
	}
}
func (r *Requests) String() string {
	if d, err := yaml.Marshal(r); err != nil {
		return err.Error()
	} else {
		return string(d)
	}
}

func (r *Requests) PushRequest(mi map[string]string) error {
	args := NewArgs(mi)
	args.Email = r.AccountId
	cl := NewClient(args, REQUEST_CREATE)
	_, err := cl.Post(r)
	if err != nil {
		return err
	}
	return nil
}
