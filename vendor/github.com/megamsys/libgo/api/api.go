package api

import (
	"bytes"
	"encoding/json"
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/libgo/utils"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
)

const (
	DELETE = "DELETE"
	POST   = "POST"
	GET    = "GET"
	UPDATE = "UPDATE"
)

type VerticeApi interface {
	ToMap() map[string]string
}

type ApiArgs struct {
	Email      string
	Api_Key    string
	Master_Key string
	Password   string
	Org_Id     string
	Url        string
	Path       string
}

func NewArgs(args map[string]string) ApiArgs {
	return ApiArgs{
		Email:      args[utils.USERMAIL],
		Api_Key:    args[utils.API_KEY],
		Master_Key: args[utils.MASTER_KEY],
		Password:   args[utils.PASSWORD],
		Org_Id:     args[utils.ORG_ID],
		Url:        args[utils.API_URL],
	}
}

func (c ApiArgs) ToMap() map[string]string {
	keys := make(map[string]string)
	s := reflect.ValueOf(&c).Elem()
	typ := s.Type()
	if s.Kind() == reflect.Struct {
		for i := 0; i < s.NumField(); i++ {
			key := s.Field(i)
			value := s.FieldByName(typ.Field(i).Name)
			switch key.Interface().(type) {
			case string:
				if value.String() != "" {
					keys[strings.ToLower(typ.Field(i).Name)] = value.String()
				}
			}
		}
	}
	return keys
}

func (c *Client) Get() ([]byte, error) {
	log.Debugf("Request [GET] ==> " + c.Url)
	return c.run(GET)
}

func (c *Client) Post(data interface{}) ([]byte, error) {
	jsonbody, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	c.Authly.JSONBody = jsonbody
	log.Debugf("Request [POST] ==> " + c.Url)
	log.Debugf("[Body]  (%s)", string(jsonbody))
	return c.run(POST)
}

func (c *Client) Delete() ([]byte, error) {
	log.Debugf("Request [DELETE] ==> " + c.Url)
	return c.run(DELETE)
}

func (c *Client) run(method string) ([]byte, error) {
	err := c.Authly.AuthHeader()
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(method, c.Url, bytes.NewReader(c.Authly.JSONBody))
	if err != nil {
		return nil, err
	}
	response, err := c.Do(request)
	if err != nil {
		return nil, err
	}

	jsonData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	return jsonData, nil
}
