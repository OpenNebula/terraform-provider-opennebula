package bills

import (
	"bytes"
	"crypto/md5"
	"encoding/base64"
	b64 "encoding/base64"
	"encoding/hex"
	log "github.com/Sirupsen/logrus"
	"github.com/megamsys/libgo/events/addons"
	"github.com/megamsys/libgo/pairs"
	constants "github.com/megamsys/libgo/utils"
	whmcs "github.com/megamsys/whmcs_go/whmcs"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	billerName = "whmcs"
)

func init() {
	Register(billerName, createBiller())
}

type whmcsBiller struct {
	enabled bool
	apiKey  string
	domain  string
}

func createBiller() BillProvider {
	vBiller := whmcsBiller{
		enabled: false,
		apiKey:  "",
		domain:  "",
	}
	log.Debugf("%s ready", billerName)
	return vBiller
}

func (w *whmcsBiller) String() string {
	return "WHMCS:(" + w.apiKey + "," + w.domain + ")"
}

func (w whmcsBiller) IsEnabled() bool {
	return w.enabled
}

func (w whmcsBiller) Onboard(o *BillOpts, m map[string]string) error {
	log.Debugf("User Onboarding...")

	bacc, err := NewAccounts(m)
	if err != nil {
		return err
	}

	org, err := AccountsOrg(o.AccountId, m)
	if err != nil {
		return err
	}

	sDec, _ := b64.StdEncoding.DecodeString(bacc.Password.Password)

	client := whmcs.NewClient(nil, m[constants.DOMAIN])
	a := map[string]string{
		"username":     m[constants.USERNAME],
		"password":     GetMD5Hash(m[constants.PASSWORD]),
		"firstname":    bacc.Name.FirstName,
		"lastname":     bacc.Name.LastName,
		"email":        bacc.Email,
		"address1":     "Dummy address",
		"city":         "Dummy city",
		"state":        "Dummy state",
		"postcode":     "00001",
		"country":      "IN",
		"phonenumber":  bacc.Phone.Phone,
		"password2":    string(sDec),
		"customfields": GetBase64(map[string]string{m[constants.VERTICE_EMAIL]: bacc.Email, m[constants.VERTICE_ORGID]: org.Id, m[constants.VERTICE_APIKEY]: bacc.ApiKey}),
	}

	_, res, err := client.Accounts.Create(a)
	if err != nil {
		return err
	}

	err = onboardNotify(bacc.Email, res.Body, m)
	return err
}

func (w whmcsBiller) Deduct(o *BillOpts, m map[string]string) (err error) {

	add := &addons.Addons{
		ProviderName: constants.WHMCS,
		AccountId:    o.AccountId,
	}

	addon, err := add.Get(m)
	if err != nil {
		return err
	}

	log.Debugf("Request WHMCS [POST] ==> " + m[constants.DOMAIN])
	client := whmcs.NewClient(nil, m[constants.DOMAIN])
	a := map[string]string{
		"username":      m[constants.USERNAME],
		"password":      GetMD5Hash(m[constants.WHMCS_PASSWORD]),
		"clientid":      addon.ProviderId,
		"description":   o.AssemblyName,
		"hours":         "0.17",
		"amount":        o.Consumed,
		"invoiceaction": "nextcron",
	}
	_, _, err = client.Billables.Create(a)
	return err
}

func (w whmcsBiller) Transaction(o *BillOpts, m map[string]string) error {
	return nil
}

func (w whmcsBiller) AuditUnpaid(o *BillOpts, m map[string]string) error {
	return nil
}

func (w whmcsBiller) Invoice(o *BillOpts) error {
	return nil
}

func (w whmcsBiller) Nuke(o *BillOpts) error {
	return nil
}

func (w whmcsBiller) Suspend(o *BillOpts) error {
	return nil
}

func (w whmcsBiller) Notify(o *BillOpts) error {
	return nil
}

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetBase64(dummymap map[string]string) string {
	var dummyVar string
	for key, value := range dummymap {
		klen := strconv.Itoa(len(key))
		vlen := strconv.Itoa(len(value))
		dummyVar += "s:" + klen + ":" + strconv.Quote(key) + ";" + "s:" + vlen + ":" + strconv.Quote(value) + ";"
	}
	kvlen := strconv.Itoa(len(dummymap))
	dummyVar1 := "a:" + kvlen + ":" + "{" + dummyVar + "}"
	input := []byte(dummyVar1)
	var outBuffer bytes.Buffer
	writer := io.MultiWriter(&outBuffer, os.Stdout)
	encoder := base64.NewEncoder(base64.StdEncoding, writer)
	encoder.Write(input)
	encoder.Close()
	res := outBuffer.String()
	return res
}

func onboardNotify(email string, r string, m map[string]string) error {
	cid := getClientId(r)
	if cid != "0" {
		return recordStatus(email, cid, "onboarded", m)
	}
	return nil
}

func getClientId(body string) string {
	id := "0"
	result := strings.Split(body, ";")
	for i := range result {
		if len(result[i]) > 0 {
			k := strings.Split(result[i], "=")
			if k[0] == "clientid" {
				id = k[1]
			}
		}
	}
	return id
}

func recordStatus(email, cid, status string, mi map[string]string) error {
	js := make(pairs.JsonPairs, 0)
	m := make(map[string][]string, 2)
	m["status"] = []string{status}
	js.NukeAndSet(m) //just nuke the matching output key:

	addon := &addons.Addons{
		Id:           "",
		ProviderName: constants.WHMCS,
		ProviderId:   cid,
		AccountId:    email,
		Options:      js.ToString(),
	}
	return addon.Onboard(mi)
}
