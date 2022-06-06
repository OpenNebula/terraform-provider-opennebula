---
layout: "opennebula"
page_title: "OpenNebula: opennebula_image"
sidebar_current: "docs-opennebula-datasource-image"
description: |-
  Get the image information for a given name.
---

# opennebula_image

Use this data source to retrieve the image information for a given name.

## Example Usage

```hcl
data "opennebula_image" "example" {
  name = "My_Image"
}
```

## Argument Reference

* `name` - (Optional) The OpenNebula image to retrieve information for.
* `tags` - (Optional) Tags associated to the image.

## Attribute Reference

* `id` - ID of the image.
* `name` - Name of the image.
* `tags` - Tags of the image (Key = Value).
