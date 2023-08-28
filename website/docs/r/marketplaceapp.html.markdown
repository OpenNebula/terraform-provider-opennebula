---
layout: "opennebula"
page_title: "OpenNebula: opennebula_marketplace_appliance"
sidebar_current: "docs-opennebula-resource-marketplace_appliance"
description: |-
  Provides an OpenNebula marketplace appliance resource.
---

# opennebula_marketplace_appliance

Provides an OpenNebula marketplace appliance resource.

This resource allows you to manage appliances on your OpenNebula marketplaces. When applied,
a new appliane is created. When destroyed, this appliance is removed.

## Example Usage

Create an appliance:

```hcl
resource "opennebula_marketplace_appliance" "example" {
  name = "test"
  market_id = "4"
  type = "VMTEMPLATE"
  description = "this is an app"
  version = "0.1.0"

  tags = {
    custom1 = "value1"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the host.
* `market_id` - (Required) The ID of the marketplace.
* `type` - (Required) Type of the new host: IMAGE, VMTEMPLATE, SERVICE_TEMPLATE.
* `origin_id` - (Optional) The ID of the source image. Default to `-1`.
* `description` - (Optional) Text description of the appliance.
* `publisher` - (Optional) Publisher of the appliance.
* `version` - (Optional) A string indicating the appliance version.
* `vmtemplate64` - (Optional) Creates this template pointing to the base image.
* `apptemplate64` - (Optional) Associated template that will be added to the registered object.
* `group` - (Optional) Name of the group owning the appliance.
* `disabled` - (Optional) Allow to enable or disable the appliance.
* `lock` - (Optional) Lock the image with a specific lock level. Supported values: `USE`, `MANAGE`, `ADMIN`, `ALL` or `UNLOCK`.
* `tags` - (Optional) Map of tags (`key=value`) assigned to the resource. Override matching tags present in the `default_tags` atribute when configured in the `provider` block. See [tags usage related documentation](https://registry.terraform.io/providers/OpenNebula/opennebula/latest/docs#using-tags) for more information.
* `template_section` - (Optional) Allow to add a custom vector. See [Template section parameters](#template-section-parameters)

### Template section parameters

`template_section` supports the following arguments:

* `name` - (Optional) The vector name.
* `elements` - (Optional) Collection of custom tags.

## Attribute Reference

The following attributes are exported:

* `id` - ID of the host.
* `tags_all` - Result of the applied `default_tags` and then resource `tags`.
* `default_tags` - Default tags defined in the provider configuration.

## Import

`opennebula_marketplace_appliance` can be imported using its ID:

```shell
terraform import opennebula_marketplace_appliance.example 123
```
