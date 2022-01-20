---
layout: "opennebula"
page_title: "Provider: OpenNebula"
sidebar_current: "docs-opennebula-index"
description: |-
  The OpenNebula provider is used to interact with OpenNebula cluster resources.
---

# OpenNebula Provider

The OpenNebula provider is used to interact with OpenNebula cluster resources.

The provider allow you to manage your OpenNebula clusters resources.
It needs to be configured with proper credentials before it can be used.

Use the navigation to the left to read about the available resources.

Current Stable release: 0.3.0
Beta Realease: 0.4.0

## Example Usage

**terraform.tfvars:**

```hcl
one_endpoint      = "http://frontzone:2633/RPC2"
one_username      = "USERNAME"
one_password      = "PASSWORD OR TOKEN"
one_flow_endpoint = "http://frontzone:2474/RPC2"
```

**terraform.tf:**

```hcl
variable "one_endpoint" {}
variable "one_username" {}
variable "one_password" {}
variable "one_flow_endpoint" {}

provider "opennebula" {
  endpoint      = "${var.one_endpoint}"
  flow_endpoint = "${var.one_flow_endpoint}"
  username      = "${var.one_username}"
  password      = "${var.one_password}"
}

# Create a new group of users to the OpenNebula cluster
resource "opennebula_group" "group" {
  # ...
}
```

## Argument Reference

The following arguments are mandatory in the `provider` block:

* `endpoint` - (Required) This is the URL of OpenNebula XML-RPC Endpoint API.
* `flow_endpoint` - (Optional) This is the OneFlow HTTP Endpoint API.
* `username` - (Required) This is the OpenNebula Username.
* `password` - (Required) This is the Opennebula Password of the username.
