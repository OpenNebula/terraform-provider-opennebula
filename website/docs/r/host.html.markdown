---
layout: "opennebula"
page_title: "OpenNebula: opennebula_host"
sidebar_current: "docs-opennebula-resource-host"
description: |-
  Provides an OpenNebula host resource.
---

# opennebula_host

Provides an OpenNebula host resource.

This resource allows you to manage hosts on your OpenNebula clusters. When applied,
a new host is created. When destroyed, this host is removed.

## Example Usage

Create a KVM host with overcommit:

```hcl
resource "opennebula_host" "example" {
  name       = "test-kvm"
  type       = "kvm"
  cluster_id = 0

  overcommit {
    cpu = 3200        # 32 cores
    memory = 1048576  # 1 Gb
  }

  tags = {
    environment = "example"
  }
}
```

Create a custom host:

```hcl
resource "opennebula_host" "example" {
  name       = "test-kvm"
  type       = "custom"

  custom = {
    virtualization = "custom"
    information    = "custom"
  }

  tags = {
    environment = "example"
  }
}
```


## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the host.
* `type` - (Required) Type of the new host: kvm, qemu, lxd, lxc, firecracker, custom. For now vcenter type is not managed by the provider.
* `cluster_id` - (Deprecated) ID of the cluster the host is part of.
* `custom` - (Optional) If `type="custom"` this section should be defined, see [Custom](#custom) section for details.
* `overcommit` - (Optional) This section allow to increase the allocatable capacity of the host. See [Overcommit](#overcommit)
* `tags` - (Optional) Host tags (Key = value)

### Custom

When the attribute `type` is set to `custom` this section allow to set the details with these arguments:

* `virtualization` - (Optional) name of the virtualization driver (named `VM_MAD` in OpenNebula)
* `information` - (Optional) name of the information driver (named `IM_MAD` in opennebula)

### Overcommit

* `cpu` - (Optional) Maximum allocatable CPU capacity  in number of cores multiplied by 100.
* `memory` - (Optional) Maximum allocatable memory in KB.

## Attribute Reference

The following attributes are exported:

* `id` - ID of the host.
* `tags_all` - Result of the applied `default_tags` and then resource `tags`.
* `default_tags` - Default tags defined in the provider configuration.
* `cluster` - ID of the cluster hosting the host.  Manager cluster membership from `hosts` fields of the `cluster` resource instead.

## Import

`opennebula_host` can be imported using its ID:

```shell
terraform import opennebula_host.example 123
```
