package main

import (
	"github.com/terraform-providers/terraform-provider-opennebula/opennebula"
	"github.com/hashicorp/terraform/plugin"
	"github.com/hashicorp/terraform/terraform"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return opennebula.Provider()
		},
	})
}
