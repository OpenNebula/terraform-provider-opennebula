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
data "opennebula_template" "example" {
  name = "My_Template"
}
```

## Argument Reference

* `name` - (Optional) The OpenNebula template to retrieve information for.
* `has_cpu` - (Optional) Indicate if a CPU value has been defined.
* `cpu` - (Optional) Amount of CPU shares assigned to the VM.
* `has_cvpu` - (Optional) Indicate if a VCPU value has been defined.
* `vpcu` - (Optional) Number of CPU cores presented to the VM.
* `has_memory` - (Optional) Indicate if a memory value has been defined.
* `memory` - (Optional) Amount of RAM assigned to the VM in MB.
* `tags` - (Optional) Template tags (Key = Value).

## Attribute Reference

The following attributes are exported:

* `id` - ID of the template.
* `name` - Name of the template.
* `cpu` - Amount of CPU shares assigned to the VM.
* `vpcu` - Number of CPU cores presented to the VM.
* `memory` - Amount of RAM assigned to the VM in MB.
* `disk` - Disk parameters
* `nic` - NIC parameters
* `vmgroup` - VM group parameters
* `tags` - Tags of the template (Key = Value).
