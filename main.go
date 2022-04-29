package main

import (
	"github.com/OpenNebula/terraform-provider-opennebula/opennebula"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return opennebula.Provider()
		},
	})
}
