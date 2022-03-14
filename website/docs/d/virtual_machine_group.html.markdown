---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_machine_group"
sidebar_current: "docs-opennebula-datasource-virtual-machine-group"
description: |-
  Get the virtual machine group information for a given name.
---

# opennebula_virtual_machine_group

Use this data source to retrieve the virtual machine group information for a given name.

## Example Usage

```hcl
data "opennebula_virtual_machine_group" "ExistingVMGroup" {
  name = "My_VMGroup"
}
```

## Argument Reference

* `name` - (Optional) The OpenNebula virtual machine group to retrieve information for.
* `tags` - (Optional) Virtual Machine group tags (Key = Value).

## Attribute Reference

* `id` - ID of the virtual machine group.
* `name` - Name of the virtual machine group.
* `tags` - Tags of the virtual machine group (Key = Value).
