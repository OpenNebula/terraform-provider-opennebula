---
layout: "opennebula"
page_title: "OpenNebula: opennebula_template"
sidebar_current: "docs-opennebula-datasource-template"
description: |-
  Get the template information for a given name.
---

# opennebula_template

Use this data source to retrieve the template information for a given name.

## Example Usage

```hcl
data "opennebula_template" "ExistingTemplate" {
  name = "My_Template"
}
```

## Argument Reference

* `name` - (Required) The OpenNebula template to retrieve information for.
* `cpu` - (Optional) Amount of CPU shares assigned to the VM.
* `vpcu` - (Optional) Number of CPU cores presented to the VM.
* `memory` - (Optional) Amount of RAM assigned to the VM in MB.
* `context` - (Optional) Array of free form key=value pairs, rendered and added to the CONTEXT variables for the VM. Recommended to include: `NETWORK = "YES"` and `SET_HOSTNAME = "$NAME"`.
* `graphics` - (Optional) Graphics parameters.
* `os` - (Optional) OS parameters
* `disk` - (Optional) Disk parameters
* `nic` - (Optional) NIC parameters
* `vmgroup` - (Optional) VM group parameters
* `tags` - (Optional) Template tags (Key = Value).
* `template` - (Deprecated) Text describing the OpenNebula template object, in Opennebula's XML string format.
