package api

import (
	"bytes"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strings"
)

const (
	HOST                       = "url"
	PATH                       = "path"
	EMAIL                      = "email"
	PASSWORD                   = "password"
	MASTER_KEY                 = "master_key"
	API_KEY                    = "api_key"
	ORG_ID                     = "org_id"
	X_Megam_EMAIL              = "X-Megam-EMAIL"
	X_Megam_MASTERKEY          = "X-Megam-MASTERKEY"
	X_Megam_PUTTUSAVI          = "X-Megam-PUTTUSAVI"
	X_Megam_DATE               = "X-Megam-DATE"
	X_Megam_HMAC               = "X-Megam-HMAC"
	X_Megam_OTTAI              = "X-Megam-OTTAI"
	X_Megam_ORG                = "X-Megam-ORG"
	Content_Type               = "Content-Type"
	Accept                     = "Accept"
	application_vnd_megam_json = "application/vnd.megam+json"
)

func CalcHMAC(secret string, message string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	sumh := h.Sum(nil)

	sumi := make([]string, len(sumh))
	for i := 0; i < len(sumh); i++ {
		sumi[i] = ("00" + fmt.Sprintf("%x", (sumh[i]&0xff)))
		sumi[i] = sumi[i][len(sumi[i])-2 : len(sumi[i])]
	}
	outs := strings.Join(sumi, "")
	return outs
}

func GetMD5Hash(text []byte) string {
	hasher := md5.New()
	hasher.Write(text)
	return base64.URLEncoding.EncodeToString(hasher.Sum(nil))
}

// func GetMD5Hash(text []byte) string {
// 	hasher := md5.New()
// 	hasher.Write(text)
// 	return CalcBase64(string(hasher.Sum(nil)))
// }

func CalcBase64(data string) string {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)
	encoder.Write([]byte(data))
	encoder.Close()
	return buf.String()
}
