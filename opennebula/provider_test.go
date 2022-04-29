package opennebula

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"os"
	"testing"
)

func TestProvider(t *testing.T) {
	if err := Provider().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider()
	testAccProviders = map[string]terraform.ResourceProvider{
		"opennebula": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	testEnvIsSet("OPENNEBULA_ENDPOINT", t)
	testEnvIsSet("OPENNEBULA_USERNAME", t)
	testEnvIsSet("OPENNEBULA_PASSWORD", t)
	testEnvIsSet("OPENNEBULA_FLOW_ENDPOINT", t)
}

func testEnvIsSet(k string, t *testing.T) {
	if v := os.Getenv(k); v == "" {
		t.Fatalf("%s must be set for acceptance tests", k)
	}
}
