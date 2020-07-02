---
layout: "opennebula"
page_title: "OpenNebula: opennebula_image"
sidebar_current: "docs-opennebula-resource-image"
description: |-
  Provides an OpenNebula image resource.
---

# opennebula_image

Provides an OpenNebula image resource.

This resource allows you to manage images on your OpenNebula clusters. When applied,
a new image is created. When destroyed, this image is removed.

## Example Usage

Clone an existing image and make it persistent:

```hcl
resource "opennebula_image" "osimageclone" {
    clone_from_image = 12937
    name             = "terraclone-image"
    datastore_id     = 113
    persistent       = true
    permissions      = "660"
    group            = "terraform"
}
```

Allocate a new OS image using a URL:

```hcl
resource "opennebula_image" "osimage" {
    name         = "Ubuntu 18.04"
    description  = "Terraform image"
    datastore_id = 103
    persistent   = false
    lock         = "MANAGE"
    path         = "http://marketplace.opennebula.org/appliance/ca5c3632-359a-429c-ac5b-b86178ee2390/download/0"
    dev_prefix   = "vd"
    driver       = "qcow2"
    permissions  = "660"
    group        = "terraform"
    timeout      = 15
    tags = {
      environment = "dev"
    }
}
```

Allocate a new persistent 1GB datablock image:

```hcl
resource "opennebula_image" "datablockimage" {
    name         = "terra-datablock"
    description  = "Terraform datablock"
    datastore_id = 103
    persistent   = true
    type         = "DATABLOCK"
    size         = "1024"
    dev_prefix   = "vd"
    driver       = "qcow2"
    group        = "terraform"
    tags = {
      environment = "dev"
    }
}
```

Allocate a new context file:

```hcl
resource "opennebula_image" "contextfile" {
    name         = "terra-contextfile"
    description  = "Terraform context"
    datastore_id = 2
    type         = "CONTEXT"
    path         = "http://server/myscript.sh"
    tags = {
      environment = "dev"
    }
}
```

Allocate a new CDROM image:

```hcl
resource "opennebula_image" "cdimage" {
    name         = "terra-cdimage"
    description  = "Terraform cdrom"
    datastore_id = 103
    type         = "CDROM"
    path         = "http://server/mini.iso"
    tags = {
      environment = "dev"
    }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the image.
* `description` - (Optional) Description of the image.
* `permissions` - (Optional) Permissions applied to the image. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin).
* `clone_from_image` - (Optional) ID or name of the image to clone from. Conflicts with `path`, `size` and `type`.
* `datastore_id` - (Required) ID of the datastore used to store the image. The `datastore_id` must be an active `IMAGE` datastore.
* `persistent` - (Optional) Flag which indicates if the Image has to be persistent. Defaults to `false`.
* `lock` - (Optional) Lock the image with a specific lock level. Supported values: `USE`, `MANAGE`, `ADMIN`, `ALL` or `UNLOCK`.
* `path` - (Optional) Path or URL of the original image to use. Conflicts with `clone_from_image`.
* `type` - (Optional) Type of the image. Supported values: `OS`, `CDROM`, `DATABLOCK`, `KERNEL`, `RAMDISK` or `CONTEXT`. Conflicts with `clone_from_image`.
* `size` - (Optional) Size of the image in MB. Conflicts with `clone_from_image`.
* `dev_prefix` - (Optional) Device prefix on Virtual Machine. Usually one of these: `hd`, `sd` or `vd`.
* `target` - (Optional) Device target on Virtual Machine.
* `driver` - (Optional) OpenNebula Driver to use.
* `format` - (Optional) Image format. Example: `raw`, `qcow2`.
* `group` - (Optional) Name of the group which owns the image. Defaults to the caller primary group.
* `tags` - (Optional) Image tags (Key = value)
* `timeout` - (Optional) Timeout (in Minutes) for Image availability. Defaults to 10 minutes.

## Attribute Reference

The following attributes are exported:
* `id` - ID of the image.
* `uid` - User ID whom owns the image.
* `gid` - Group ID which owns the image.
* `uname` - User Name whom owns the image.
* `gname` - Group Name which owns the image.

## Import

To import an existing image #14 into Terraform, add this declaration to your .tf file:

```hcl
resource "opennebula_image" "importimage" {
    name = "importedimage"
}
```

And then run:

```
terraform import opennebula_image.importimage 14
```

Verify that Terraform does not perform any change:

```
terraform plan
```
