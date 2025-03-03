package main

import (
	"flag"

	"github.com/OpenNebula/terraform-provider-opennebula/opennebula"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	plugin.Serve(&plugin.ServeOpts{
		Debug: debug,
		ProviderFunc: func() *schema.Provider {
			return opennebula.Provider()
		},
		ProviderAddr: "OpenNebula/opennebula",
	})
}
