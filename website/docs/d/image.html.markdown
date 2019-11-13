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
data "opennebula_image" "ExistingImagr" {
  name = "My_Image"
}
```

## Argument Reference

 * `name` - (Required) The OpenNebula image to retrieve information for.

