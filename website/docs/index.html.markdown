---
layout: "opennebula"
page_title: "Provider: OpenNebula"
sidebar_current: "docs-opennebula-index"
description: |-
  The OpenNebula provider is used to interact with OpenNebula cluster resources.
---

# OpenNebula Provider

The OpenNebula provider is used to interact with OpenNebula cluster resources.

The provider allows you to manage your OpenNebula clusters resources.
It needs to be configured with proper credentials before it can be used.

## Example Usage

```hcl
terraform {
  required_providers {
    opennebula = {
      source = "OpenNebula/opennebula"
      version = "~> 1.0"
    }
  }
}

provider "opennebula" {
  endpoint      = "https://example.com:2633/RPC2"
}

resource "opennebula_group" "group" {
  # ...
}
```

## Authentication and Configuration

The configuration of the OpenNebula Provider can be set by the `provider` block attributes or by the environment variables.

### Provider configuration

* `endpoint` - (Required) The URL of OpenNebula XML-RPC Endpoint API (for example, `http://example.com:2633/RPC2`).
* `flow_endpoint` - (Optional) The OneFlow HTTP Endpoint API (for example, `http://example.com:2474/RPC2`).
* `username` - (Required) The OpenNebula username.
* `password` - (Required) The Opennebula password matching the username.
* `insecure` - (Optional) Allow insecure connexion (skip TLS verification).
* `default_tags` - (Optional) Apply default custom tags to created resources: group, image, security_group, template, vm_group, user, virtual_machine, virtual_network, virtual_router, virtual_router_instance. Theses tags could be overriden in the tag section of the resource. See [Default Tags parameters](#default-tags-parameters) below for details.

#### Default Tags parameters

`default_tags` supports the following arguments:

* `tags` - (Optional) Map of tags.

#### Example

```hcl
provider "opennebula" {
  endpoint      = "https://example.com:2633/RPC2"
  flow_endpoint = "https://example.com:2474/RPC2"
  username      = "me"
  password      = "p@s5w0rD"
  insecure      = true

  default_tags {
    tags = {
      environment = "default"
    }
  }
}

resource "opennebula_group" "group" {
  # ...
}
```

```bash
terraform init
terraform plan
```

!> **Warning:** Hard-coded credentials are not recommended in any Terraform configuration file and should not be commited in a public repository.

### Environment variables

The provider can also read the following environment variables if no value is set in the the `provider` block attributes:

* `OPENNEBULA_ENDPOINT`
* `OPENNEBULA_FLOW_ENDPOINT`
* `OPENNEBULA_USERNAME`
* `OPENNEBULA_PASSWORD`
* `OPENNEBULA_INSECURE`

#### Example

```bash
export OPENNEBULA_ENDPOINT="https://example.com:2633/RPC2"
export OPENNEBULA_FLOW_ENDPOINT="https://example.com:2474/RPC2"
export OPENNEBULA_USERNAME="me"
export OPENNEBULA_PASSWORD="p@s5w0rD"
export OPENNEBULA_INSECURE="true"
```

```hcl
provider "opennebula" {}

resource "opennebula_group" "group" {
  # ...
}
```
