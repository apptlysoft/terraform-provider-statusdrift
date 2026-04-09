//go:generate tfplugindocs generate --provider-name statusdrift

package main

import (
	"context"
	"log"

	"github.com/apptlysoft/terraform-provider-statusdrift/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var version = "dev"

func main() {
	err := providerserver.Serve(context.Background(), provider.New(version), providerserver.ServeOpts{
		Address: "registry.terraform.io/apptlysoft/statusdrift",
	})
	if err != nil {
		log.Fatal(err)
	}
}
