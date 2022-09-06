---
layout: "opennebula"
page_title: "OpenNebula: opennebula_security_group"
sidebar_current: "docs-opennebula-resource-security-group"
description: |-
  Provides an OpenNebula security group resource.
---

# opennebula_security_group

Provides an OpenNebula security group resource.

This resource allows you to manage security groups on your OpenNebula clusters. When applied,
a new security group is created. When destroyed, this security group is removed.

## Example Usage

```hcl
resource "opennebula_security_group" "example" {
  name        = "security-group"
  description = "Terraform security group"
  group       = "terraform"

  rule {
    protocol  = "ALL"
    rule_type = "OUTBOUND"
  }

  rule {
    protocol  = "TCP"
    rule_type = "INBOUND"
    range     = "22"
  }

  rule {
    protocol  = "ICMP"
    rule_type = "INBOUND"
  }

  tags = {
    environment = "example"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the security group.
* `description` - (Optional) Description of the security group.
* `permissions` - (Optional) Permissions applied on security group. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin).
* `commit` - (Optional) Flag to commit changes on Virtual Machine on security group update. Defaults to `true`.
* `rule` - (Required) List of rules. See [Rule parameters](#rule-parameters) below for details
* `group` - (Optional) Name of the group which owns the security group. Defaults to the caller primary group.
* `tags` - (Optional) Security group tags (Key = Value).

### Rule parameters

`rule` supports the following arguments:

* `protocol` - (Required) Protocol for the rule. Supported values: `ALL`, `TCP`, `UDP`, `ICMP` or `IPSEC`.
* `rule_type` - (Required) Direction of the traffic flow to allow, must be `INBOUND` or `OUTBOUND`.
* `network_id` - (Optional) VNET ID to be used as the source/destination IP addresses.
* `ip` - (Optional) IP (or starting IP if used with 'size') to apply the rule to.
* `size` - (Optional) Number of IPs to apply the rule from, starting with `ip`.
* `range` - (Optional) Comma separated list of ports and port ranges.
* `icmp_type` - (Optional) Type of ICMP traffic to apply to when 'protocol' is `ICMP`.

See <https://docs.opennebula.org/5.12/operation/network_management/security_groups.html> for more details on allowed values.

## Attribute Reference

The following attribute are exported:

* `id` - ID of the security group.
* `uid` - User ID whom owns the security group.
* `gid` - Group ID which owns the security group.
* `uname` - User Name whom owns the security group.
* `gname` - Group Name which owns the security group.
* `tags_all` - Result of the applied `default_tags` and then resource `tags`.
* `default_tags` - Default tags defined in the provider configuration.

## Import

`opennebula_security_group` can be imported using its ID:

```shell
terraform import opennebula_security_group.example 123
```
