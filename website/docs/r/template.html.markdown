---
layout: "opennebula"
page_title: "OpenNebula: opennebula_template"
sidebar_current: "docs-opennebula-resource-template"
description: |-
  Provides an OpenNebula template resource.
---

# opennebula_template

Provides an OpenNebula template resource.

This resource allows you to manage templates on your OpenNebula clusters. When applied,
a new template will be created. When destroyed, that template will be removed.

## Example Usage

```hcl
data "template_file" "templatetpl" {
  template = "${file("template-tpl.txt")}"
}

resource "opennebula_template" "mytemplate" {
  name = "mytemplate"
  template = "${data.template_file.templatetpl.rendered}"
  group = "terraform"
  permissions = "660"
}
```

with template file `template-tpl.txt`:
```php
CPU = 1
VCPU = 1
MEMORY = 512
Context =  [ 
  DNS_HOSTNAME = "YES",
  NETWORK = "YES",
  SSH_PUBLIC_KEY = "$USER[SSH_PUBLIC_KEY]"
]
NIC_DEFAULT = [
  MODEL = "virtio-net-pci" ]
OS = [
  ARCH = "x86_64",
  BOOT = "" ]
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) The name of the virtual machine template.
* `template` - (Required) Text describing the OpenNebula template object, in Opennebula's XML string format.
* `permissions` - (Optional) Permissions applied on template. Defaults to the UMASK in OpenNebula (in UNIX Format: owner-group-other => Use-Manage-Admin.
* `group` - (Optional) Name of the group which owns the template. Defaults to the caller primary group.

## Attribute Reference

The following attribute are exported:
* `id` - ID of the template.
* `uid` - User ID whom owns the template.
* `gid` - Group ID which owns the template.
* `uname` - User Name whom owns the template.
* `gname` - Group Name which owns the template.
* `reg_time` - Registration time of the template.

## Import

Not tested yet


