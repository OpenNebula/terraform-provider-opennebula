## 0.2.1 (Unreleased)

BUG FIXES:
* resources/opennebula_virtual_machine: Revert regression introduced by b071b27b4b9f722e881f3954531a192e3cd99275 ([#52](https://github.com/terraform-providers/terraform-provider-opennebula/issues/52))
* resources/opennebula_template: Revert regression introduced by b071b27b4b9f722e881f3954531a192e3cd99275 ([#52](https://github.com/terraform-providers/terraform-provider-opennebula/issues/52))
* resources/opennebula_virtual_machine_group: Remove Computed for tags ([#53](https://github.com/terraform-providers/terraform-provider-opennebula/issues/53))
* resources/opennebula_virtual_machine: Remove Computed for tags ([#53](https://github.com/terraform-providers/terraform-provider-opennebula/issues/53))
* resources/opennebula_virtual_template: Remove Computed for tags ([#53](https://github.com/terraform-providers/terraform-provider-opennebula/issues/53))

## 0.2.0 (July 02, 2020)

NOTES:
* OpenNebula version used by CI updated to 5.12

FEATURES:
* **New Data Source**: `opennebula_virtual_machine_group`: First implementation
* **New Resource**: `opennebula_virtual_machine_group`: First implementation ([onevmgroup](http://docs.opennebula.org/5.10/integration/system_interfaces/api.html#onevmgroup)),
* **New Resource**: `opennebula_acl`: First implementation ([oneacl](http://docs.opennebula.org/5.10/integration/system_interfaces/api.html#oneacl)),
OpenNebula provider issue: ([#16](https://github.com/terraform-providers/terraform-provider-opennebula/issues/16))
* resource/opennebula_virtual_machine: Associate a VM group (only during VM creation) ([#16](https://github.com/terraform-providers/terraform-provider-opennebula/issues/16))
* resource/opennebula_template: Associate a VM group.
* resource/opennebula_image: Add support for tags ([#22](https://github.com/terraform-providers/terraform-provider-opennebula/issues/22))
* resource/opennebula_security_group: Add support for tags ([#22](https://github.com/terraform-providers/terraform-provider-opennebula/issues/22))
* resource/opennebula_template: Add support for tags ([#22](https://github.com/terraform-providers/terraform-provider-opennebula/issues/22))
* resource/opennebula_virtual_machine: Add support for tags ([#22](https://github.com/terraform-providers/terraform-provider-opennebula/issues/22))
* resource/opennebula_virtual_machine_group: Add support for tags ([#22](https://github.com/terraform-providers/terraform-provider-opennebula/issues/22))
* resource/opennebula_virtual_network: Add support for tags ([#22](https://github.com/terraform-providers/terraform-provider-opennebula/issues/22))
* resource/opennebula_virtual_machine: Add timeout parameter ([#36](https://github.com/terraform-providers/terraform-provider-opennebula/issues/36))
* resource/opennebula_mage: Add timeout parameter ([#36](https://github.com/terraform-providers/terraform-provider-opennebula/issues/36))

ENHANCEMENTS:
* all resources: use Goca dynamic templates to build entities
* all resources: update permissions to follow Goca changes
* resource/opennebula_virtual_machine: keep context from template, then override redefined pairs
* resource/opennebula_template: share with VM resource the schemas parts: cpu, vcpu, memory, context, graphics, os, disk, nic, vmgroup

DEPRECATION:
* resource/opennebula_template: Remove `template` parameter to reproduce resource/opennebula_virtual_machine details schema

BUG FIXES:
* data/opennebula_template: Fix missing parameters on Read ([#29](https://github.com/terraform-providers/terraform-provider-opennebula/issues/29))

## 0.1.1 (January 06, 2020)

BUG FIXES:
* resource/opennebula_virtual_machine: Start VM on Hold ([#1](https://github.com/terraform-providers/terraform-provider-opennebula/issues/1))
* resource/opennbula_virtual_machine: Attach nic or disk in the declared order ([#5](https://github.com/terraform-providers/terraform-provider-opennebula/issues/5))
* all ressources: Fix changes detected on update while parameters are not set ([#2](https://github.com/terraform-providers/terraform-provider-opennebula/issues/2))
* resource/opennebula_virtual_network: Fix setting of cluster id on Virtual Network Creation ([#6](https://github.com/terraform-providers/terraform-provider-opennebula/issues/6))

DEPRECATION:
* resource/opennebula_virtual_machine: Remove `instance` parameter as it is redundant with `name`

## 0.1.0 (November 25, 2019)

NOTES:
* First implementation of the provider
* Basic Tests + CI + Coverage


FEATURES:
* **New Data Source**: `opennebula_group`: First implementation
* **New Data Source**: `opennebula_image`: First implementation
* **New Data Source**: `opennebula_security_group`: First implementation
* **New Data Source**: `opennebula_template`: First implementation
* **New Data Source**: `opennebula_virtual_data_center`: First implementation
* **New Data Source**: `opennebula_virtual_network`: First implementation
* **New Resource**: `opennebula_group`: First implementation ([onegroup](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onegroup))
* **New Resource**: `opennebula_image`: First implementation ([oneimage](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#oneimage))
* **New Resource**: `opennebula_security_group`: First implementation ([onesecgroup](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onesecgroup))
* **New Resource**: `opennebula_template`: First implementation ([onetemplate](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onetemplate))
* **New Resource**: `opennebula_virtual_data_center`: First implementation ([onevdc](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onevdc))
* **New Resource**: `opennebula_virtual_machine`: First implementation ([onevm](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onevm))
* **New Resource**: `opennebula_virtual_network`: First implementation ([onevnet](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onevnet))
