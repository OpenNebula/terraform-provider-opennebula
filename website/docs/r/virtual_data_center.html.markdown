---
layout: "opennebula"
page_title: "OpenNebula: opennebula_virtual_data_center"
sidebar_current: "docs-opennebula-resource-virtual-data-center"
description: |-
  Provides an OpenNebula virtual data center resource.
---

# opennebula_virtual_data_center

Provides an OpenNebula virtual data center resource.

This resource allows you to manage virtual data centers on your OpenNebula organization. When applied,
a new virtual data center will be created. When destroyed, that virtual data center will be removed.

## Example Usage

```hcl
resource "opennebula_virtual_data_center" "vdc" {
    name = "terravdc"
    group_ids = ["${opennebula_group.group.id}"]
    zones {
        id = 0
        host_ids = [0, 1]
        datastore_ids = [0, 1, 2]
        vnet_ids = ["${opennebula_virtual_network.vnet.id}"]
        cluster_ids = [0, 100]
    }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the virtual data center.
* `groups_ids` - (Optional) List of group IDs part of the virtual data center.
* `zones` - (Optional) List of zones. See [Zones parameters](#zones) below for details

### Zones parameters

`zones` supports the following arguments:

* `id` - (Optional) Zone ID from where resource to add in virtual data center are located. Defaults to 0.
* `host_ids` - (Optional) List of hosts from Zone ID to add in virtual data center.
* `datastore_ids` - (Optional) List of datastore from Zone ID to add in virtual data center.
* `vnet_ids` - (Optional) List of virtual networks from Zone ID to add in virtual data center.
* `cluster_ids` - (Optional) List of clusters from Zone ID to add in virtual data center.

## Attribute Reference

The following attribute is exported:
* `id` - ID of the virtual data center.

## Import

Not tested yet


