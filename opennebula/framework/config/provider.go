package config

import (
	"github.com/OpenNebula/one/src/oca/go/src/goca"
	ver "github.com/hashicorp/go-version"

	"github.com/OpenNebula/terraform-provider-opennebula/opennebula/framework/utils"
)

type Provider struct {
	OneVersion  *ver.Version
	Controller  *goca.Controller
	Mutex       utils.MutexKV
	DefaultTags map[string]string
}
