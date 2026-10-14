package opennebula

import (
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm"
	vmk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/vm/keys"
)

func featuresTestData(t *testing.T) *schema.ResourceData {
	t.Helper()
	return schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"features": commonTemplateSchemas()["features"],
	}, map[string]interface{}{})
}

func TestFlattenFeaturesReadsLocalTime(t *testing.T) {
	tpl := vm.NewTemplate()
	if err := tpl.AddFeature(featureLocalTime, "YES"); err != nil {
		t.Fatal(err)
	}
	if err := tpl.AddFeature(vmk.ACPI, "YES"); err != nil {
		t.Fatal(err)
	}

	// goca names this LOCAL_TIME, which is not what OpenNebula stores.
	if got, _ := tpl.GetFeature(vmk.LocalTime); got != "" {
		t.Fatalf("vmk.LocalTime = %q, want empty", got)
	}

	d := featuresTestData(t)
	if err := flattenFeatures(d, tpl); err != nil {
		t.Fatal(err)
	}

	features := d.Get("features").([]interface{})
	if len(features) != 1 {
		t.Fatalf("features length = %d, want 1", len(features))
	}
	got := features[0].(map[string]interface{})
	if got["localtime"] != "YES" {
		t.Fatalf("localtime = %#v, want YES", got["localtime"])
	}
	if got["acpi"] != "YES" {
		t.Fatalf("acpi = %#v, want YES", got["acpi"])
	}
}

func TestFlattenFeaturesReadsLegacyLocalTime(t *testing.T) {
	tpl := vm.NewTemplate()
	if err := tpl.AddFeature(vmk.LocalTime, "YES"); err != nil {
		t.Fatal(err)
	}

	d := featuresTestData(t)
	if err := flattenFeatures(d, tpl); err != nil {
		t.Fatal(err)
	}

	features := d.Get("features").([]interface{})
	got := features[0].(map[string]interface{})
	if got["localtime"] != "YES" {
		t.Fatalf("localtime = %#v, want YES", got["localtime"])
	}
}

func TestAddFeaturesUpdatesAttributeInPlace(t *testing.T) {
	tpl := vm.NewTemplate()
	if err := tpl.AddFeature(vmk.ACPI, "YES"); err != nil {
		t.Fatal(err)
	}
	if err := tpl.AddFeature(vmk.Feature("MIGRATE_AUTO_CONVERGE"), "YES"); err != nil {
		t.Fatal(err)
	}
	if err := tpl.AddFeature(vmk.LocalTime, "NO"); err != nil {
		t.Fatal(err)
	}

	addFeatures(tpl, []map[string]interface{}{
		{
			"acpi":      "YES",
			"localtime": "YES",
		},
	})

	if got := featureValue(tpl, vmk.ACPI); got != "YES" {
		t.Fatalf("ACPI = %q", got)
	}
	if got := featureValue(tpl, featureLocalTime); got != "YES" {
		t.Fatalf("LOCALTIME = %q", got)
	}
	if got := featureValue(tpl, vmk.LocalTime); got != "" {
		t.Fatalf("LOCAL_TIME = %q, want removed", got)
	}
	if got := featureValue(tpl, vmk.Feature("MIGRATE_AUTO_CONVERGE")); got != "YES" {
		t.Fatalf("unmanaged feature = %q, want preserved", got)
	}
	if strings.Count(tpl.String(), "FEATURES=") != 1 {
		t.Fatalf("template = %s", tpl.String())
	}
}

func TestVirtioBlkQueuesRoundTrip(t *testing.T) {
	tpl := vm.NewTemplate()
	addFeatures(tpl, []map[string]interface{}{
		{"virtio_blk_queues": "4"},
	})

	if got := featureValue(tpl, featureVirtioBlkQueues); got != "4" {
		t.Fatalf("VIRTIO_BLK_QUEUES = %q", got)
	}

	d := featuresTestData(t)
	if err := flattenFeatures(d, tpl); err != nil {
		t.Fatal(err)
	}

	features := d.Get("features").([]interface{})
	if len(features) != 1 {
		t.Fatalf("features length = %d, want 1", len(features))
	}
	if got := features[0].(map[string]interface{})["virtio_blk_queues"]; got != "4" {
		t.Fatalf("virtio_blk_queues = %#v", got)
	}
}

func TestNormalizeRawDataIgnoresEdgeWhitespace(t *testing.T) {
	config := "  <features>\n    <hyperv/>\n  </features>\n"
	remote := "<features>\n    <hyperv/>\n  </features>"

	if !diffSuppressRawData("raw.0.data", remote, config, nil) {
		t.Fatalf("normalize(%q) = %q, normalize(%q) = %q",
			remote, normalizeRawData(remote), config, normalizeRawData(config))
	}
	if diffSuppressRawData("raw.0.data", remote, "<features>\n    <hyperv state='on'/>\n  </features>", nil) {
		t.Fatal("real XML changes must not be suppressed")
	}
}
