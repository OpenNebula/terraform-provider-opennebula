---
layout: "opennebula"
page_title: "OpenNebula: opennebula_marketplace"
sidebar_current: "docs-opennebula-resource-marketplace"
description: |-
  Provides an OpenNebula marketplace resource.
---

# opennebula_marketplace

Provides an OpenNebula marketplace resource.

This resource allows you to manage marketplaces.

## Example Usage

Create a custom marketplace:

```hcl
resource "opennebula_marketplace" "example" {
  name = "example"

  s3 {
    type = "aws"
    access_key_id = "XXX"
    secret_access_key = "XXX"
    region = "us"
    bucket = "123"
  }

 tags = {
  environment = "example"
 }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the marketplace.
* `description` - (Optional) The description of the marketplace.
* `permissions` - (Optional) Permissions applied to the appliance. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin).
* `one` - (Optional) See [One](#One) section for details.
* `http` - (Optional) See [Http](#Http) section for details.
* `s3` - (Optional) See [S3](#S3) section for details.
* `lxc` - (Optional) See [LXC](#LXC) section for details.
* `turnkey` - (Optional) See [Turnkey](#Turnkey) section for details.
* `dockerhub` - (Optional) Flag as a dockerhub marketplace, this provide access to DockerHub Official Images.
* `tags` - (Optional) Marketplace tags (Key = value).

### One

The OpenNebula Marketplace is a catalog of virtual appliances ready to run in OpenNebula environments.

The following arguments are supported:

* `endpoint_url` - (Optional) The endpoint URL of the marketplace.

### Http

Http Marketplace uses a conventional HTTP server to expose the images (Marketplace Appliances) uploaded to the Marketplace.

The following arguments are supported:

* `endpoint_url` - (Required) Base URL of the Marketplace HTTP endpoint.
* `path` - (Required) Absolute directory path to place images in the front-end or in the hosts pointed at by storage_bridge_list.
* `storage_bridge_list` - (Optional) List of servers to access the public directory.

### S3

This Marketplace uses an S3 API-capable service as the Back-end.

The following arguments are supported:

* `type` - (Optional) Type of the s3 backend: aws, ceph, minio.
* `access_key_id` - (Required) The access key of the S3 user.
* `secret_access_key` - (Required) The secret key of the S3 user.
* `bucket` - (Required) The bucket where the files will be stored.
* `region` - (Required) The region to connect to. Any value will work with Ceph-S3.
* `endpoint_url` - (Optional) Only required when connecteing to a service other than Amazon S3.
* `total_size` - (Optional) Define the total size of the marketplace in MB.
* `read_block_length` - (Optional) Split the file into chunks of this size in MB, never user a value larger than 100. 

### LXC

The Linux Containers image server hosts a public image server with container images for LXC and LXD.

The following arguments are supported:

* `endpoint_url` - (Optional) The base URL of the Market.
* `roofs_image_size` - (Optional) Size in MB for the image holding the rootfs.
* `filesystem` - (Optional) Filesystem used for the image.
* `image_block_file_format` - (Optional) Image block file format.
* `skip_untested` - (Optional) Include only apps with support for context.
* `cpu` - (Optional) VM template CPU.
* `vcpu` - (Optional) VM template VCPU.
* `memory` - (Optional) VM template memory.
* `privileged` - (Optional) Secrurity mode of the Linux Container.

## Turnkey

The following arguments are supported:

* `endpoint_url` - (Optional) The base URL of the Market.
* `roofs_image_size` - (Optional) Size in MB for the image holding the rootfs.
* `filesystem` - (Optional) Filesystem used for the image.
* `image_block_file_format` - (Optional) Image block file format.
* `skip_untested` - (Optional) Include only apps with support for context.

## Attribute Reference

The following attributes are exported:

* `id` - ID of the marketplace.
* `tags_all` - Result of the applied `default_tags` and then resource `tags`.
* `default_tags` - Default tags defined in the provider configuration.

## Import

`opennebula_marketplace` can be imported using its ID:

```shell
terraform import opennebula_marketplace.example 123
```
