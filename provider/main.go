package main

import (
	"github.com/LeMikaelF/meetings-as-code/provider/provider"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// FIXME provider can't use stdout or stderr, so authentication needs to happen outside of provider binary.
func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			return provider.Provider()
		},
	})
}
