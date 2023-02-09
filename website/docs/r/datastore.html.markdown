---
layout: "opennebula"
page_title: "OpenNebula: opennebula_datastore"
sidebar_current: "docs-opennebula-resource-datastore"
description: |-
  Provides an OpenNebula datastore resource.
---

# opennebula_datastore

Provides an OpenNebula datastore resource.

This resource allows you to manage datastores.

## Example Usage

Create a custom datastore:

```hcl
resource "opennebula_datastore" "example" {
 name = "example"
 type = "image"

 custom {
  datastore = "dummy"
  transfer = "dummy"
 }

 tags = {
  environment = "example"
 }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the datastore.
* `type` - (Required) Type of the new datastore: image, system, file.
* `cluster_id` - (Deprecated) ID of the cluster the datastore is part of.
* `cluster_ids` - (Optional) IDs of the clusters the datastore is part of. Minimum 1 item.
* `restricted_directories` - (Optional) Paths that cannot be used to register images. A space separated list of paths.
* `safe_directories` - (Optional) If you need to allow a directory listed under RESTRICTED_DIRS. A space separated list of paths.
* `no_decompress` - (Optional) Boolean, do not try to untar or decompress the file to be registered.
* `storage_usage_limit` - (Optional) The maximum capacity allowed for the Datastore in MB.
* `transfer_bandwith_limit` - (Optional) Specify the maximum transfer rate in bytes/second when downloading images from a http/https URL. Suffixes K, M or G can be used.
* `check_available_capacity` - (Optional) If yes, the available capacity of the Datastore is checked before creating a new image.
* `bridge_list` - (Optional) List of hosts that have access to the storage to add new images to the datastore.
* `staging_dir` - (Optional) Path in the storage bridge host to copy an Image before moving it to its final destination.
* `driver` - (Optional) Specific image mapping driver enforcement. If present it overrides image DRIVER set in the image attributes and VM template.
* `compatible_system_datastore` - (Optional) Specify the compatible system datastores.
* `ceph` - (Optional) See [Ceph](#ceph) section for details.
* `custom` - (Optional) See [Custom](#custom) section for details.
* `tags` - (Optional) Datastore tags (Key = value).

### Ceph

The following arguments are supported:

* `pool_name` - (Optional) Ceph pool name.
* `user` - (Optional) Ceph user name.
* `key` - (Optional) Key file for user. if not set default locations are used.
* `config` - (Optional) Non-default Ceph configuration file if needed.
* `rbd_format` - (Optional) By default RBD Format 2 will be used.
* `secret` - (Optional) The UUID of the libvirt secret.
* `host` - (Optional) List of Ceph monitors.
* `local_storage` - (Optional) Use local host storage, SSH mode.
* `trash` - (Optional) Enables trash feature on given datastore.

### Custom

The following arguments are supported:

* `datastore` - (Optional) name of the datastore driver (named `DS_MAD` in OpenNebula).
* `transfer` - (Optional) name of the transfer driver (named `TM_MAD` in opennebula).

### Overcommit

* `cpu` - (Optional) Maximum allocatable CPU capacity  in number of cores multiplied by 100.
* `memory` - (Optional) Maximum allocatable memory in KB.

## Attribute Reference

The following attributes are exported:

* `id` - ID of the datastore.
* `tags_all` - Result of the applied `default_tags` and then resource `tags`.
* `default_tags` - Default tags defined in the provider configuration.

## Import

`opennebula_datastore` can be imported using its ID:

```shell
terraform import opennebula_datastore.example 123
```
