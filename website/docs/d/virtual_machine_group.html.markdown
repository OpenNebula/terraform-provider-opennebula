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

 * `name` - (Required) The OpenNebula virtual machine group to retrieve information for.

