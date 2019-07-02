package opennebula

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
	"os"
	"testing"
)

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"opennebula": testAccProvider,
	}
}

func testAccPreCheck(t *testing.T) {
	testEnvIsSet("OPENNEBULA_ENDPOINT", t)
	testEnvIsSet("OPENNEBULA_USERNAME", t)
	testEnvIsSet("OPENNEBULA_PASSWORD", t)
}

func testEnvIsSet(k string, t *testing.T) {
	if v := os.Getenv(k); v == "" {
		t.Fatalf("%s must be set for acceptance tests", k)
	}
}
