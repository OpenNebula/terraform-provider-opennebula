# addon-terraform

## Description

[OpenNebula](https://opennebula.org/) provider for [Terraform](https://www.terraform.io/).

## Compatibility

* Leverages [OpenNebula's XML/RPC API](https://docs.opennebula.org/5.88/integration/system_interfaces/api.html)
* Tested for version 5.8

This provider has been initiated to use official Goca from [OpenNebula](https://github.com/OpenNebula/one)
The version 1.0 of the provider is indended to be used with Terraform version Terraform 0.12+

For Older OpenNebula and Terraform releases, you can use non official provider from [Runtastic](https://github.com/runtastic/terraform-provider-opennebula) and enhanced by [BlackBerry](https://github.com/blackberry/terraform-provider-opennebula).

## Contributors

* OpenNebula:

* Iguane Solutions:
- Jean-Philippe Fourès (https://github.com/jaypif)
- Pierre Lafièvre (https://github.com/treywelsh)
- Edouard Hur (https://github.com/hekmon)
- Benjamin Gustin (https://github.com/aloababa)

## Contributing

Bug reports and pull requests are welcome on GitHub at
https://github.com/OpenNebula/addon-terraform/issues.

Please follow [How to Contribute](https://github.com/OpenNebula/one/wiki/How-to-Contribute-to-Development) rules for any pull request.

## Resources and Data sources:

Current definition of these resource types are supported yet:
* Groups
* Image
* Security Groups
* Template
* Virtual Data Center

As well as data sources for:
* Groups
* Image
* Security Groups
* Template
* Virtual Data Center
* Virtual Machine

## DOCUMENTATION
TODO Update Wiki page

## ROADMAP

The following list represent's all of OpenNebula's resources reachable through their API. The checked items are the ones that are functional and tested:

* [X] [onevm](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onevm)
* [X] [onetemplate](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onetemplate)
* [ ] [onehost](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onehost)
* [ ] [onecluster](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onecluster)
* [X] [onegroup](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onegroup)
* [X] [onevdc](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onevdc)
* [ ] [onevnet](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onevnet)
* [ ] [oneuser](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#oneuser)
* [ ] [onedatastore](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onedatastore)
* [X] [oneimage](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#oneimage)
* [ ] [onemarket](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onemarket)
* [ ] [onemarketapp](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onemarketapp)
* [ ] [onevrouter](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onevrouter)
* [ ] [onezone](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onezone)
* [X] [onesecgroup](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#onesecgroup)
* [ ] [oneacl](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#oneacl)
* [ ] [oneacct](https://docs.opennebula.org/5.8/integration/system_interfaces/api.html#oneacct)
