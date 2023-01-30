# 1.1.1 (January 30th, 2023)

BUG FIXES:

* resources/opennebula_datastore: fix system type values (#382)
* resources/opennebula_host: fix host resource type case (#385)
* resources/opennebula_virtual_machine: import more sections and attributes: `os`, `graphics`, `cpu_model`, `features`, `sched_requirements`, `sched_ds_requirements`, `description`, `template_id` (#377, #312)
* resources/opennebula_virtual_router_instance: import more sections and attributes: `os`, `graphics`, `cpu_model`, `features`, `sched_requirements`, `sched_ds_requirements`, `description`, `template_id` (#377, #312)
* resources/opennebula_template: import more sections and attributes: `os`, `graphics`, `cpu_model`, `features`, `sched_requirements`, `sched_ds_requirements`, `description` (#377)
* resources/opennebula_virtual_router_instance_template: import more sections and attributes: `os`, `graphics`, `cpu_model`, `features`, `sched_requirements`, `sched_ds_requirements`, `description` (#377)
* resources/opennebula_virtual_machine: set empty values instead of null for `template_disk`, `template_nic`, `template_tags` (#312, #369)
* resources/opennebula_virtual_router_instance: set empty values instead of null for `template_disk`, `template_nic`, `template_tags` (#312, #369)
* resources/opennebula_datastore: add argument `cluster_ids` (#389)
* resources/opennebula_virtual_network: add argument `cluster_ids` (#389)
* resources/opennebula_datastore: conditional reading of `datastore` argument from `custom`. (#382)
* resources/opennebula_virtual_network_address_range: modify `hold_ips` content reading and introduce `helds_ips` attribute (#397)
* resources/opennebula_virtual_network: for reservation, fix `type` and `reservation_ar_id` reading. (#397)
* resources/opennebula_host: set overcommit map only when not empty (#399)

DEPRECATION:

* resources/opennebula_cluster: deprecate `hosts`, `datastores`, `virtual_networks` (#389)
* resources/opennebula_datastore: deprecate `cluster_id` (#389)
* resources/opennebula_virtual_network: deprecate `clusters` (#389)

# 1.1.0 (December 6th, 2022)

FEATURES:

* **New Resource**: `opennebula_cluster` (#227)
* **New Resource**: `opennebula_datastore` (#299)
* **New Data Source**: `opennebula_datastore` (#299)
* resources/opennebula_cluster: add `template_section` to manage vectors with an unique key (#359)
* resources/opennebula_group: add `template_section` to manage vectors with an unique key (#359)
* resources/opennebula_image: add `template_section` to manage vectors with an unique key (#359)
* resources/opennebula_security_group: add `template_section` to manage vectors with an unique key (#359)
* resources/opennebula_template: add `template_section` to manage vectors with an unique key (#359)
* resources/opennebula_vm_group: add `template_section` to manage vectors with an unique key (#359)
* resources/opennebula_user: add `template_section` to manage vectors with an unique key (#359)
* resources/opennebula_virtual_machine: add `template_section` to manage vectors with an unique key (#359)
* resources/opennebula_virtual_network: add `template_section` to manage vectors with an unique key (#359)
* resources/opennebula_virtual_router: add `template_section` to manage vectors with an unique key (#359)
* resources/opennebula_virtual_router_instance: add `template_section` to manage vectors with an unique key (#359)
* resources/opennebula_virtual_router_instance_template: add `template_section` to manage vectors with an unique key (#359)
* **New Resource**: `opennebula_host` (#300)
* **New Data source**: `opennebula_host` (#300)

DEPRECATION:

* resources/opennebula_group: remove deprecated attribute `delete_on_destruction` (#297)
* resources/opennebula_group: remove deprecated attribute `template` (#297)
* data/opennebula_group: remove deprecated attribute `quotas` (#297)
* data/opennebula_user remove deprecated attribute `quotas` (#297)
* data/opennebula_template: remove deprecated attribute `context` (#297)
* data/opennebula_template: remove deprecated attribute `graphics` (#297)
* data/opennebula_template: remove deprecated attribute `os` (#297)
* data/opennebula_virtual_network: remove deprecated attribute `description` (#297)

BUG FIXES:

* data/opennebula_template: simplify `hasXXX` filter handling (#370)
* data/opennebula_image: goca dependency update: pool info method retrieve all (#331)
* data/opennebula_security_group: goca dependency update: pool info method retrieve all (#331)
* data/opennebula_vm_group: goca dependency update: pool info method retrieve all (#331)
* data/opennebula_template: goca dependency update: pool info method retrieve all (#331)
* data/opennebula_virtual_network: goca dependency update: pool info method retrieve all (#331)
* resources/opennebula_cluster: fix resource existence test at read (#373)
* resources/opennebula_group: fix resource existence test at read (#373)
* resources/opennebula_image: fix resource existence test at read (#373)
* resources/opennebula_security_group: fix resource existence test at read (#373)
* resources/opennebula_template: fix resource existence test at read (#373)
* resources/opennebula_vm_group: fix resource existence test at read (#373)
* resources/opennebula_user: fix resource existence test at read (#373)
* resources/opennebula_data_center: fix resource existence test at read (#373)
* resources/opennebula_virtual_machine: fix resource existence test at read (#373)
* resources/opennebula_virtual_network: fix resource existence test at read (#373)
* resources/opennebula_virtual_router: fix resource existence test at read (#373)

# 1.0.2 (November 8th, 2022)

BUG FIXES:

* resources/opennebula_group: Add `opennebula` section (#358)
* resource/opennebula_virtual_machine: Fix ignored NIC with `security_groups` configured (#342)

# 1.0.1 (October 3rd, 2022)

BUG FIXES:

* resources/opennebula_user: Fix ignored renaming (#343)
* resources/opennebula_group: Fix ignored renaming (#343)

ENHANCEMENTS:

* resources/opennebula_group_admins: Replace Typelist by Typeset on `users_ids` (#352)
* resources/opennebula_user: Replace Typelist by Typeset on `groups` (#352)
* resources/opennebula_virtual_data_center: Replace Typelist by Typeset on `group_ids`, `host_ids`, `datastore_ids`, `vnet_ids`, `cluster_ids` (#352)
* resources/opennebula_network: Replace Typelist by Typeset on `clusters`, `security_groups` (#352)

## 1.0.0 (September 19th, 2022)

BUG FIXES:

* resource/opennebula_virtual_machine: Fix diff on template inherited attributes `sched_requirements` and `sched_ds_requirements` (#330)

FEATURES:

* **New Resource**: `opennebula_virtual_network_address_range` (#279)
* resources/opennebula_virtual_network: add attributes `reservation_first_ip` and `reservation_ar_id` (#274)

ENHANCEMENTS:

* resources/opennebula_virtual_machine: add `dev_prefix`, `cache`, `discard` and `io` to `disk` (#291)
* resources/opennebula_virtual_network: add `network_address` and `search_domain` attributes (#292)
* provider: add attribute `insecure` to allow skipping TLS verification (#328)
* data/opennebula_template: add `has_cpu`, `has_vcpu`, `has_memory` (#287)
* provider: add section `default_tags` for group, image, security_group, template, vm_group, user, virtual_machine, virtual_network, virtual_router, virtual_router_instance, virtual_router_instance_template resources (#324)

DEPRECATION:

* resource/opennebula_virtual_network: remove deprecated attributes `hold_size` and `ip_hold` (#296)
* resource/opennebula_virtual_machine: remove deprecated attribute `instance` (#296)
* resources/opennebula_virtual_network: deprecated `ar` and `hold_ips` (#279)

## 0.5.2 (August 10th, 2022)

BUG FIXES:

* resources/opennebula_virtual_machine: allow to delete a VM in PENDING state (#315)
* resources/opennebula_virtual_machine: read disk, description and sched_requirements even if empty (#304)
* resources/opennebula_template: read disk, nic, description and sched_requirements even if empty (#304)
* resources/opennebula_virtual_machine: read disk even if empty (#304)
* data/opennebula_virtual_data_center: read tags even if emtpy (#304)
* data/opennebula_virtual_network: read tags even if emtpy (#304)
* resources/opennebula_group: read tags even if emtpy (#304)
* resources/opennebula_image: read tags even if emtpy (#304)
* resources/opennebula_security_group: read tags even if emtpy (#304)
* resources/opennebula_template_vm_group: read tags even if emtpy (#304)
* resources/opennebula_user: read tags even if emtpy (#304)
* resources/opennebula_virtual_machine: read tags even if emtpy (#304)
* resources/opennebula_template: read tags even if emtpy (#304)
* resources/opennebula_virtual_network: read tags even if emtpy (#304)
* resources/opennebula_virtual_router: read tags even if emtpy (#304)
* data/opennebula_virtual_network: MTU is optional (#284)
* resources/opennebula_virtual_machine: fix multiline regression (#309)

## 0.5.1 (July 4th, 2022)

ENHANCEMENTS:

* provider: replace several deprecated SDK functions (#269)
* resources/opennebula_virtual_machine: deprecate custom `timeout` attribute in favor of the SDK timeout facilities (#267)
* resources/opennebula_virtual_router_instance: deprecate custom `timeout` attribute in favor of the SDK timeout facilities (#267)
* resources/opennebula_image: deprecate custom `timeout` attribute in favor of the SDK timeout facilities (#267)

BUG FIXES:

* provider: fix incorrect conversions between integer types (#278)
* provider: fail on bad credentials (#288)
* data/opennebula_template: fix an error where `cpu`, `vcpu` or `memory` are undefined (#284)
* resources/opennebula_virtual_machine: fix missing NIC generation (#289)
* resources/opennebula_virtual_machine: fix VM state management failures (#132)

## 0.5.0 (June 7th, 2022)

NOTES:

* All datasources schemas have been reworked and an independant read method has been added for each (#229)
* The provider has been migrated to use the SDK v2 (#161)
* OpenNebula binding (goca) dependency has been updated to the 6.4 release (#270)

FEATURES:

* **New Data Source**: `opennebula_cluster`: allow filtering based on `name` and `tags` (#234)
* **New Resources**: `opennebula_virtual_router`, `opennebula_virtual_router_instance`, `opennebula_virtual_router_instance_template`, `opennebula_virtual_router_nic` (#170)
* resources/opennebula_virtual_machine: Add 'on_disk_change' property to opennebula_virtual_machine (#184)
* **New Resource**: `opennebula_group_admins` (#245)
* resources/opennebula_template: add `features` section (#237)

ENHANCEMENTS:

* data/opennebula_group: make `name` optional and add `tags` filtering (#268)
* data/opennebula_image: make `name` optional and enable `tags` filtering (#229)
* data/opennebula_security_group: make `name` optional and enable `tags` filtering (#229)
* data/opennebula_template: make `name` optional and enable `tags` filtering (#229)
* data/opennebula_template_vm_group: make `name` optional and enable `tags` filtering (#229)
* data/opennebula_user: make `name` optional and enable `tags` filtering (#229)
* data/opennebula_virtual_data_center: make `name` optional and add `tags` filtering (#229)
* data/opennebula_virtual_network: make `name` optional and enable `tags` filtering (#229)
* resources/opennebula_group: add `sunstone` and `tags` sections (#251)
* resources/opennebula_virtual_network: compatibility added for network states (#270)
* resources/opennebula_virtal_machine: enable VM vcpu, cpu and memory update (#273)
* resources/opennebula_user: add `tags` sections (#275)
* resources/opennebula_acl: enable `zone` parameter (#277)

DEPRECATION:

* data/opennebula_group: deprecate `quotas`, `template`, remove `users` (#251, #229)
* data/opennebula_group: deprecate `delete_on_destruction` and set its default value to `true` (#253)
* data/opennebula_template: deprecate `context`, `graphics` and `os`. Make `disk`, `nic` and `vmgroup` computed. Remove `template` (#229)
* data/opennebula_user: deprecate `quotas` and `auth_driver` (#229)
* data/opennebula_virtual_network: deprecate `description`. Make `mtu` computed (#229)

BUG FIXES:

* resources/opennebula_security_group: read `name` (#229)
* resources/opennebula_virtual_machine: fix volatile disk update adding `computed_volatile_format` (#260)
* resources/opennebula_virtual_machine: fix template quotes escaping (#270)
* resources/opennebula_template: fix template quotes escaping (#270)
* resources/opennebula_template: fix reading and update of `cpu`, `vcpu`, `memory` (#236)
* resources/opennebula_virtual_machine: fix reading of `cpu`, `vcpu`, `memory` (#236)
* resources/openenbula_image: fix key deletions in `tags`  (#275)
* resources/opennebula_security_group: fix key deletions in `tags`  (#275)

## 0.4.3 (March 23th, 2022)

* Support for `darwin/arm64` and `windows/arm64` platforms.
* Documentation update.

## 0.4.2 (March 10th, 2022)

BUG FIXES:

* resources/opennebula_virtual_machine: fix description duplication
* resources/opennebula_template: check fields at read
* resources/opennebula_template: fix template update

## 0.4.1 (February 15th, 2022)

BUG FIXES:

* resources/opennebula_service_template: Fix `template` diff method to perform deep equal check over `ServiceTemplate` struct instead of binary file diff.
* resources/opennebula_virtual_network: check empty ar at read
* resources/opennebula_template: check `user_inputs` at read
* resources/opennebula_virtual_machine: fix update of user_template related attributes
* resources/opennebula_image: remove `computed_size` attribute
* resources/opennebula_virtual_network: remove ar ordering code
* resources/opennebula_group: detailed error messages

## 0.4.0 (January 20th, 2022)

*/!\ DISCALAIMER:*
This release is *NOT* stable, it is considered as a BETA for 0.4 validation purpose
The current stable release remains 0.3.0.

FEATURES:

* resources/opennebula_virtual_machine: Optionally preserve NIC ordering
* resources/opennebula_virtual_machine: Enable locking
* resources/opennebula_virtual_network: Enable locking
* resources/opennebula_template: Enable locking

BUG FIXES:

* resources/opennebula_virtual_network: fix type at read for reservation_vnet
* resources/opennebula_virtual_network: reservation_vnet: Zero is a valid ID
* resources/opennebula_virtual_machine: Fix several disks attached to the same images
* resources/opennebula_virtual_data_center: Fix `zones` flattening and associated tests
* resources/opennebula_user: Fix crash on quota datas reading
* resources/opennebula_group: Fix crash on quota datas reading
* resources/opennebula_virtual_machine: Fix several NICs attached to the same network
* resources/opennebula_security_group: fix rule conversion from struct to config
* resources/opennebula_virtual_machine_group: make `role` reading conditional
* resources/opennebula_virtual_machine_group: remove `vms` field
* resources/opennebula_service: add compatibility with OneFlow server > `5.12.x`
* data/opennebula_user: remove password field

FEATURES:

* resources/opennebula_virtual_machine: Add description, sched_requirements, sched_ds_requirements
* resources/opennebula_template: Add description, user_inputs, sched_requirements, sched_ds_requirements

ENHANCEMENTS:

* resources/opennebula_virtual_network: Enhance address range update
* resources/opennebula_virtual_machine: enable context, os, graphics sections update
* resources/opennebula_virtual_machine: Allow VM deletion from other states than RUNNING
* resources/opennebula_image: Enable description update
* resources/opennebula_virtual_machine: Enable volatile disks

DEPRECATION:

* resources/opennebula_group: deprecate `users` to move group membership management on user resource side

## 0.3.0 (December 17, 2020)

FEATURES:

* **New Resource**** New Data Source**: opennebula_user : First implementation ([#69](https://github.com/OpenNebula/terraform-provider-opennebula/issues/69))
* resources/opennebula_virtual_machine: Enable VM disk update ([#64](https://github.com/OpenNebula/terraform-provider-opennebula/issues/64))
* resources/opennebula_virtual_machine: Change 'image_id' disk attribute from Required to Optional ([#71](https://github.com/OpenNebula/terraform-provider-opennebula/issues/71))
* **New Resource**: `opennebula_service`: First implementation ([oneflow](http://docs.opennebula.io/5.12/integration/system_interfaces/appflow_api.html#service)),
* **New Resource**: `opennebula_service_template`: First implementation ([oneflow-template](http://docs.opennebula.io/5.12/integration/system_interfaces/appflow_api.html#service-template)),
* resources/opennebula_virtual_machine: Enable VM NIC update ([#72](https://github.com/OpenNebula/terraform-provider-opennebula/issues/72))

BUG FIXES:

* resources/opennebula_virtual_network: Fix Hold IPs crash ([#67](https://github.com/OpenNebula/terraform-provider-opennebula/issues/67))
* resources/opennebula_virtual_network: Fix Documentation about AR usage ([#66](https://github.com/OpenNebula/terraform-provider-opennebula/issues/66))

DEPRECATION:

* resource/opennebula_virtual_network: Replace `hold_size` and `ip_hold` parameters by `hold_ips`

## 0.2.2 (October 16, 2020)

New release only for Terraform Registry migration

## 0.2.1 (July 03, 2020)

BUG FIXES:

* resources/opennebula_virtual_machine: Revert regression introduced by b071b27b4b9f722e881f3954531a192e3cd99275 ([#52](https://github.com/OpenNebula/terraform-provider-opennebula/issues/52))
* resources/opennebula_template: Revert regression introduced by b071b27b4b9f722e881f3954531a192e3cd99275 ([#52](https://github.com/OpenNebula/terraform-provider-opennebula/issues/52))
* resources/opennebula_virtual_machine_group: Remove Computed for tags ([#53](https://github.com/OpenNebula/terraform-provider-opennebula/issues/53))
* resources/opennebula_virtual_machine: Remove Computed for tags ([#53](https://github.com/OpenNebula/terraform-provider-opennebula/issues/53))
* resources/opennebula_virtual_template: Remove Computed for tags ([#53](https://github.com/OpenNebula/terraform-provider-opennebula/issues/53))

## 0.2.0 (July 02, 2020)

NOTES:

* OpenNebula version used by CI updated to 5.12

FEATURES:

* **New Data Source**: `opennebula_virtual_machine_group`: First implementation
* **New Resource**: `opennebula_virtual_machine_group`: First implementation ([onevmgroup](http://docs.opennebula.org/5.10/integration/system_interfaces/api.html#onevmgroup)),
* **New Resource**: `opennebula_acl`: First implementation ([oneacl](http://docs.opennebula.org/5.10/integration/system_interfaces/api.html#oneacl)),
OpenNebula provider issue: ([#16](https://github.com/OpenNebula/terraform-provider-opennebula/issues/16))
* resource/opennebula_virtual_machine: Associate a VM group (only during VM creation) ([#16](https://github.com/OpenNebula/terraform-provider-opennebula/issues/16))

OpenNebula provider issue: ([#16](https://github.com/terraform-providers/terraform-provider-opennebula/issues/16))

* resource/opennebula_virtual_machine: Associate a VM group (only during VM creation) ([#16](https://github.com/terraform-providers/terraform-provider-opennebula/issues/16))
* resource/opennebula_template: Associate a VM group.
* resource/opennebula_image: Add support for tags ([#22](https://github.com/OpenNebula/terraform-provider-opennebula/issues/22))
* resource/opennebula_security_group: Add support for tags ([#22](https://github.com/OpenNebula/terraform-provider-opennebula/issues/22))
* resource/opennebula_template: Add support for tags ([#22](https://github.com/OpenNebula/terraform-provider-opennebula/issues/22))
* resource/opennebula_virtual_machine: Add support for tags ([#22](https://github.com/OpenNebula/terraform-provider-opennebula/issues/22))
* resource/opennebula_virtual_machine_group: Add support for tags ([#22](https://github.com/OpenNebula/terraform-provider-opennebula/issues/22))
* resource/opennebula_virtual_network: Add support for tags ([#22](https://github.com/OpenNebula/terraform-provider-opennebula/issues/22))
* resource/opennebula_virtual_machine: Add timeout parameter ([#36](https://github.com/OpenNebula/terraform-provider-opennebula/issues/36))
* resource/opennebula_mage: Add timeout parameter ([#36](https://github.com/OpenNebula/terraform-provider-opennebula/issues/36))

ENHANCEMENTS:

* all resources: use Goca dynamic templates to build entities
* all resources: update permissions to follow Goca changes
* resource/opennebula_virtual_machine: keep context from template, then override redefined pairs
* resource/opennebula_template: share with VM resource the schemas parts: cpu, vcpu, memory, context, graphics, os, disk, nic, vmgroup

DEPRECATION:

* resource/opennebula_template: Remove `template` parameter to reproduce resource/opennebula_virtual_machine details schema

BUG FIXES:

* data/opennebula_template: Fix missing parameters on Read ([#29](https://github.com/OpenNebula/terraform-provider-opennebula/issues/29))

## 0.1.1 (January 06, 2020)

BUG FIXES:

* resource/opennebula_virtual_machine: Start VM on Hold ([#1](https://github.com/OpenNebula/terraform-provider-opennebula/issues/1))
* resource/opennbula_virtual_machine: Attach nic or disk in the declared order ([#5](https://github.com/OpenNebula/terraform-provider-opennebula/issues/5))
* all ressources: Fix changes detected on update while parameters are not set ([#2](https://github.com/OpenNebula/terraform-provider-opennebula/issues/2))
* resource/opennebula_virtual_network: Fix setting of cluster id on Virtual Network Creation ([#6](https://github.com/OpenNebula/terraform-provider-opennebula/issues/6))

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
