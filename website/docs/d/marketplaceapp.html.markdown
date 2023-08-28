---
layout: "opennebula"
page_title: "OpenNebula: opennebula_marketplace_appliance"
sidebar_current: "docs-opennebula-datasource-marketplace_appliance"
description: |-
  Get the marketplace appliance information for a given name.
---

# opennebula_marketplace_appliance

Use this data source to retrieve the marketplace appliance information for a given name.

## Example Usage

```hcl
data "opennebula_marketplace_appliance" "example" {
  name = "My_Appliance"
}
```

## Argument Reference

* `id` - (Optional) ID of the marketplace appliance.
* `name` - (Optional) The OpenNebula marketplace appliance to retrieve information for.
* `tags` - (Optional) Tags associated to the marketplace appliance.

## Attribute Reference

* `id` - ID of the marketplace appliance.
* `name` - Name of the marketplace appliance.
* `tags` - Tags of the marketplace appliance (Key = Value).
