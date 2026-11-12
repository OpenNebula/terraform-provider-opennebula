package opennebula

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func securityGroupProtocolValidator(t *testing.T) schema.SchemaValidateFunc {
	t.Helper()
	rule := resourceOpennebulaSecurityGroup().Schema["rule"].Elem.(*schema.Resource)
	return rule.Schema["protocol"].ValidateFunc
}

func TestSecurityGroupProtocolValidation(t *testing.T) {
	validate := securityGroupProtocolValidator(t)

	for _, proto := range []string{"ALL", "TCP", "UDP", "ICMP", "VRRP", "IPSEC"} {
		t.Run("should accept "+proto, func(t *testing.T) {
			if _, errs := validate(proto, "protocol"); len(errs) != 0 {
				t.Fatalf("expected %s to be valid, got %v", proto, errs)
			}
		})
	}

	t.Run("should reject an unknown protocol", func(t *testing.T) {
		if _, errs := validate("BOGUS", "protocol"); len(errs) == 0 {
			t.Fatal("expected BOGUS to be rejected")
		}
	})
}
