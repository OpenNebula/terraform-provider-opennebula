package api

import (
	"net/url"
	"strings"
	"time"
)

type Authly struct {
	UrlSuffix string
	Date      string
	JSONBody  []byte
	Keys      map[string]string
	AuthMap   map[string]string
}

func NewAuthly(c VerticeApi) *Authly {
	m := c.ToMap()
	auth := &Authly{
		Date:      time.Now().Format(time.RFC850),
		UrlSuffix: m[PATH],
		Keys:      m,
		AuthMap:   map[string]string{},
	}
	return auth
}

func GetPort() string {
	return "port"
}

func (auth *Authly) GetURL() string {
	return strings.TrimRight(auth.Keys[HOST], "/") + strings.TrimRight(auth.UrlSuffix, "/")
}

func (authly *Authly) AuthHeader() error {
	headMap := make(map[string]string)
	key := ""
	v, err := url.Parse(authly.Keys[HOST])
	if err != nil {
		return err
	}
	timeStampedPath := authly.Date + "\n" + v.Path + authly.UrlSuffix
	md5Body := GetMD5Hash(authly.JSONBody)
	switch true {
	case (authly.Keys[API_KEY] != ""):
		key = authly.Keys[API_KEY]
	case (authly.Keys[PASSWORD] != ""):
		key = authly.Keys[EMAIL]
		headMap[X_Megam_PUTTUSAVI] = "true"
	case (authly.Keys[MASTER_KEY] != ""):
		key = authly.Keys[MASTER_KEY]
		headMap[X_Megam_MASTERKEY] = "true"
	}

	headMap[X_Megam_ORG] = authly.Keys[ORG_ID]
	headMap[X_Megam_DATE] = authly.Date
	headMap[X_Megam_EMAIL] = authly.Keys[EMAIL]
	headMap[Accept] = application_vnd_megam_json
	headMap[X_Megam_HMAC] = authly.Keys[EMAIL] + ":" + CalcHMAC(key, (timeStampedPath+"\n"+md5Body))
	headMap["Content-Type"] = "application/json"
	authly.AuthMap = headMap
	return nil
}
