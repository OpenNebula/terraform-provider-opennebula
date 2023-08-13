---
layout: "opennebula"
page_title: "OpenNebula: opennebula_cluster"
sidebar_current: "docs-opennebula-resource-cluster"
description: |-
  Provides an OpenNebula cluster resource.
---

# opennebula_cluster

Provides an OpenNebula cluster resource.

## Example Usage

```hcl
resource "opennebula_cluster" "example" {
  name = "example"

  tags = {
    environment = "example"
  }

  template_section {
    name = "example"
    elements = {
      tag1 = "value"
    }
  }

}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the cluster.
* `tags` - (Optional) Map of tags (`key=value`) assigned to the resource. Override matching tags present in the `default_tags` atribute when configured in the `provider` block. See [tags usage related documentation](https://registry.terraform.io/providers/OpenNebula/opennebula/latest/docs#using-tags) for more information.
* `template_section` - (Optional) Allow to add a custom vector. See [Template section parameters](#template-section-parameters)

### Template section parameters

`template_section` supports the following arguments:

* `name` - (Optional) The vector name.
* `elements` - (Optional) Collection of custom tags.

## Attribute Reference

The following attribute is exported:

* `id` - ID of the cluster.
* `tags_all` - Result of the applied `default_tags` and then resource `tags`.
* `default_tags` - Default tags defined in the provider configuration.
* `hosts` - List of hosts user IDs part of the cluster.
* `datastores` - List of hosts user IDs part of the cluster.
* `virtual_networks` - List of hosts user IDs part of the cluster.

## Import

`opennebula_cluster` can be imported using its ID:

```shell
terraform import opennebula_cluster.example 123
```

<https://registry.terraform.io/docs/providers/aws/r/iam_service_linked_role.html>
/docs/providers/aws/r/iam_service_linked_role.html
