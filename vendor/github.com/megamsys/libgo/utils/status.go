/*
** Copyright [2013-2017] [Megam Systems]
**
** Licensed under the Apache License, Version 2.0 (the "License");
** you may not use this file except in compliance with the License.
** You may obtain a copy of the License at
**
** http://www.apache.org/licenses/LICENSE-2.0
**
** Unless required by applicable law or agreed to in writing, software
** distributed under the License is distributed on an "AS IS" BASIS,
** WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
** See the License for the specific language governing permissions and
** limitations under the License.
 */

package utils

// Status represents the status of a unit in vertice
type Status string

func (s Status) String() string {
	return string(s)
}

type State string

func (s State) String() string {
	return string(s)
}

func (s Status) Event_type() string {
	switch s.String() {
	case LAUNCHING:
		return ONEINSTANCELAUNCHINGTYPE
	case BALANCECHECK:
		return ONEINSTANCESBALANCEVERIFYTYPE
	case INSUFFICIENT_FUND:
		return ONEINSTANCESINSUFFIENTFUNDTYPE
	case QUOTA_UNPAID:
		return ONEINSTANCESQUOTAUNPAID
	case QUOTAUPDATING:
		return ONEINSTANCEUSERQUOTAUPDATING
	case QUOTAUPDATED:
		return ONEINSTANCEUSERQUOTAUPDATED
	case VMBOOTING:
		return ONEINSTANCEBOOTINGTYPE
	case LAUNCHED:
		return ONEINSTANCELAUNCHEDTYPE
	case BOOTSTRAPPING:
		return ONEINSTANCEBOOTSTRAPPINGTYPE
	case BOOTSTRAPPED:
		return ONEINSTANCEBOOTSTRAPPEDTYPE
	case STATEUPPING:
		return ONEINSTANCESTATEUPPINGTYPE
	case STATEUPPED:
		return ONEINSTANCESTATEUPPEDTYPE
	case RUNNING:
		return ONEINSTANCERUNNINGTYPE
	case STARTING:
		return ONEINSTANCESTARTINGTYPE
	case STARTED:
		return ONEINSTANCESTARTEDTYPE
	case STOPPING:
		return ONEINSTANCESTOPPINGTYPE
	case STOPPED:
		return ONEINSTANCESTOPPEDTYPE
	case SUSPENDING:
		return ONEINSTANCESUSPENDINGTYPE
	case SUSPENDED:
		return ONEINSTANCESUSPENDEDTYPE
	case UPGRADED:
		return ONEINSTANCEUPGRADEDTYPE
	case DESTROYING:
		return ONEINSTANCEDESTROYINGTYPE
	case NUKED:
		return ONEINSTANCEDELETEDTYPE
	case SNAPSHOTTING:
		return ONEINSTANCESNAPSHOTTINGTYPE
	case SNAPSHOTTED:
		return ONEINSTANCESNAPSHOTTEDTYPE
	case COOKBOOKDOWNLOADING:
		return COOKBOOKDOWNLOADINGTYPE
	case COOKBOOKDOWNLOADED:
		return COOKBOOKDOWNLOADEDTYPE
	case COOKBOOKFAILURE:
		return COOKBOOKFAILURETYPE
	case APPDEPLOYING:
		return ONEINSTANCEAPPDEPLOYING
	case APPDEPLOYED:
		return ONEINSTANCEAPPDEPLOYED
	case NETWORK_UNAVAIL:
		return ONEINSTANCENETWORKUNAVAILABLE
	case VNCHOSTUPDATING:
		return ONEINSTANCEVNCHOSTUPDATING
	case VNCHOSTUPDATED:
		return ONEINSTANCEVNCHOSTUPDATED
	case AUTHKEYSUPDATING:
		return AUTHKEYSUPDATINGTYPE
	case AUTHKEYSUPDATED:
		return AUTHKEYSUPDATEDTYPE
	case AUTHKEYSFAILURE:
		return AUTHKEYSFAILURETYPE
	case CHEFCONFIGSETUPSTARTING:
		return ONEINSTANCECHEFCONFIGSETUPSTARTING
	case CHEFCONFIGSETUPSTARTED:
		return ONEINSTANCECHEFCONFIGSETUPSTARTED
	case INSTANCEIPSUPDATING:
		return INSTANCEIPSUPDATINGTYPE
	case INSTANCEIPSUPDATED:
		return INSTANCEIPSUPDATEDTYPE
	case INSTANCEIPSFAILURE:
		return INSTANCEIPSFAILURETYPE
	case CONTAINERNETWORKSUCCESS:
		return CONTAINERNETWORKSUCCESSTYPE
	case CONTAINERNETWORKFAILURE:
		return CONTAINERNETWORKFAILURETYPE
	case DNSNETWORKCREATING:
		return ONEINSTANCEDNSCNAMING
	case DNSNETWORKCREATED:
		return ONEINSTANCEDNSCNAMED
	case DNSNETWORKSKIPPED:
		return ONEINSTANCEDNSNETWORKSKIPPED
	case CLONING:
		return ONEINSTANCEGITCLONING
	case CLONED:
		return ONEINSTANCEGITCLONED
	case CONTAINERLAUNCHING:
		return CONTAINERINSTANCELAUNCHINGTYPE
	case CONTAINERBOOTSTRAPPING:
		return CONTAINERINSTANCEBOOTSTRAPPING
	case CONTAINERBOOTSTRAPPED:
		return CONTAINERINSTANCEBOOTSTRAPPED
	case CONTAINERLAUNCHED:
		return CONTAINERINSTANCELAUNCHEDTYPE
	case CONTAINEREXISTS:
		return CONTAINERINSTANCEEXISTS
	case CONTAINERDELETE:
		return CONTAINERINSTANCEDELETE
	case CONTAINERSTARTING:
		return CONTAINERINSTANCESTARTING
	case CONTAINERSTARTED:
		return CONTAINERINSTANCESTARTED
	case CONTAINERSTOPPING:
		return CONTAINERINSTANCESTOPPING
	case CONTAINERSTOPPED:
		return CONTAINERINSTANCESTOPPED
	case CONTAINERRESTARTING:
		return CONTAINERINSTANCERESTARTING
	case CONTAINERRESTARTED:
		return CONTAINERINSTANCERESTARTED
	case CONTAINERUPGRADED:
		return CONTAINERINSTANCEUPGRADED
	case CONTAINERRUNNING:
		return CONTAINERINSTANCERUNNING
	case CONTAINERERROR:
		return CONTAINERINSTANCEERROR

	case WAITUNTILL:
		return ONEINSTANCEWAITING
	case LCMSTATECHECK:
		return ONEINSTANCELCMSTATECHECKING
	case VMSTATECHECK:
		return ONEINSTANCEVMSTATECHECKING
	case PENDING:
		return ONEINSTANCEVMSTATEPENDING
	case HOLD:
		return ONEINSTANCEVMSTATEHOLD
	case ACTIVE + "_lcm_init":
		return ONEINSTANCELCMSTATEINIT
	case ACTIVE + "_boot":
		return ONEINSTANCELCMSTATEBOOT
	case ACTIVE + "_prolog":
		return ONEINSTANCELCMMSTATEPROLOG
	case RESETPASSWORD:
		return INSTANCERESETOLDPASSWORD
	case PREERROR:
		return ONEINSTANCEPREERRORTYPE
	case ERROR:
		return ONEINSTANCEERRORTYPE
	default:
		return "arrgh"
	}
}

func (s Status) MkEvent_type() string {
	switch s.String() {
	case DATABLOCK_CREATING:
		return MARKETPLACEBLOCKCREATING
	case DATABLOCK_CREATED:
		return MARKETPLACEBLOCKCREATED
	case LAUNCHING:
		return MARKETPLACEINSTANCELAUNCHINGTYPE
	case LAUNCHED:
		return MARKETPLACEINSTANCELAUNCHEDTYPE
	case VMBOOTING:
	case VNCHOSTUPDATING:
		return MARKETPLACEVNCHOSTUPDATING
	case VNCHOSTUPDATED:
		return MARKETPLACEVNCHOSTUPDATED
	case WAITUNTILL:
		return MARKETPLACEWAITING
	case LCMSTATECHECK:
		return MARKETPLACELCMSTATECHECKING
	case VMSTATECHECK:
		return MARKETPLACEVMSTATECHECKING
	case PENDING:
		return MARKETPLACEVMSTATEPENDING
	case HOLD:
		return MARKETPLACEVMSTATEHOLD
	case ACTIVE + "_lcm_init":
		return MARKETPLACELCMSTATEINIT
	case ACTIVE + "_boot":
		return MARKETPLACELCMSTATEBOOT
	case ACTIVE + "_prolog":
		return MARKETPLACELCMSTATEPROLOG
	case PREERROR:
		return MARKETPLACEPREERRORTYPE
	case IMAGE_SAVING:
		return MARKETPLACEIMAGESAVING
	case IMAGE_SAVED:
		return MARKETPLACEIMAGESAVED
	case IMAGE_READY:
		return MARKETPLACEIMAGEREADY
	}
	return "arrah"
}

func (s Status) Description(name string) string {
	error_common := "Oops something went wrong on .."
	switch s.String() {
	case LAUNCHING:
		return "Your machine is initializing.."
	case BALANCECHECK:
		return "Verifying credit balance.."
	case INSUFFICIENT_FUND:
		return "Insuffient funds on you wallet to launch machine .."
	case QUOTA_UNPAID:
		return "You have unpaid invoice. Pay your invoice, and redeploy."
	case VMBOOTING:
		return "Created machine by the hypervisor, watch the console for boot ..."
	case LAUNCHED:
		return "Machine was initialized on cloud.."
	case QUOTAUPDATING:
		return "Machine is updating into quota.."
	case QUOTAUPDATED:
		return "Machine is updated into quota.."
	case VNCHOSTUPDATING:
		return "Enabling vnc console access.."
	case VNCHOSTUPDATED:
		return "Enabled, you can access your machine console.."
	case BOOTSTRAPPING:
		return "Bootstrapping your machine.."
	case BOOTSTRAPPED:
		return "Bootstrapped your machine.."
	case STATEUPPING:
		return "Submitting request to initiate creation of DNS record"
	case STATEUPPED:
		return "Submission accepted to create DNS record... "
	case RUNNING:
		return "Your machine is running.."
	case APPDEPLOYING:
		return "Your application is deploying.."
	case APPDEPLOYED:
		return "Your application is deployed.."
	case STARTING:
		return "Your machine is  starting.."
	case STARTED:
		return "Your machine was started.."
	case STOPPING:
		return "Stopping process initializing on .."
	case STOPPED:
		return "Your machine was stopped.."
	case SUSPENDING:
		return "Suspend process initializing on .."
	case SUSPENDED:
		return "Your machine was suspended.."
	case UPGRADED:
		return "Your machine was was upgraded.."
	case DESTROYING:
		return "Your machine is getting removed."
	case NUKED:
		return "Your machine was removed.."
	case SNAPSHOTTING:
		return "Snapshot in progress."
	case SNAPSHOTTED:
		return "Snapshot created.."
	case DISKATTACHING:
		return "Attaching a volume to your machine"
	case DISKATTACHED:
		return "Attached a volume storage to your machine"
	case DISKDETACHING:
		return "Removing a volume storage from your machine"
	case DISKDETACHED:
		return "Removed a volume storage from your machine"
	case COOKBOOKDOWNLOADING:
		return "Downloading infrastructure automation instructions."
	case COOKBOOKDOWNLOADED:
		return "Downloaded infrastructure automation instructions.."
	case COOKBOOKFAILURE:
		return error_common
	case CHEFCONFIGSETUPSTARTING:
		return "Preparing configuration parameters."
	case CHEFCONFIGSETUPSTARTED:
		return "Prepared configuration paramter for the install.."
	case CLONING:
		return "Cloning your git repository .."
	case CLONED:
		return "Cloned your git repository .."
	case DNSNETWORKCREATING:
		return "Creating DNS CNAME entry.."
	case DNSNETWORKCREATED:
		return "Created DNS CNAME entry.."
	case DNSNETWORKSKIPPED:
		return "Skipped DNS CNAME."
	case AUTHKEYSUPDATING:
		return "Configuring, ssh with access credentials.."
	case AUTHKEYSUPDATED:
		return "Configured, ssh with access credentials.."
	case AUTHKEYSFAILURE:
		return error_common
	case INSTANCEIPSUPDATING:
		return "Updating public and private ips"
	case INSTANCEIPSUPDATED:
		return "Updated public and private ips"
	case INSTANCEIPSFAILURE:
		return error_common
	case CONTAINERNETWORKSUCCESS:
		return "Private and public ips are updated on your " + name
	case CONTAINERNETWORKFAILURE:
		return error_common + "Updating private and public ips on .."
	case CONTAINERLAUNCHING:
		return "Your  container is initializing.."
	case CONTAINERBOOTSTRAPPING:
		return name + " was bootstrapping.."
	case CONTAINERBOOTSTRAPPED:
		return name + " was bootstrapped.."
	case CONTAINERLAUNCHED:
		return "Container  was initialized on cloud.."
	case CONTAINEREXISTS:
		return name + "was exists.."
	case CONTAINERDELETE:
		return name + "was deleted.."
	case CONTAINERSTARTING:
		return "Starting process initializing on .."
	case CONTAINERSTARTED:
		return name + " was started.."
	case CONTAINERSTOPPING:
		return "Stopping process initializing on .."
	case CONTAINERSTOPPED:
		return name + " was stopped.."
	case CONTAINERRESTARTING:
		return "Restarting process initializing on .."
	case CONTAINERRESTARTED:
		return name + " was restarted.."
	case CONTAINERUPGRADED:
		return name + " was upgraded.."
	case CONTAINERRUNNING:
		return name + "was running.."
	case VMSTATECHECK:
		return "Machine state checking"
	case WAITUNTILL:
		return "[20 seconds] machine is deploying.."
	case PENDING:
		return "Selecting the node to deploy"
	case HOLD:
		return "Scheduling for deployment"
	case ACTIVE + "_lcm_init":
		return "Internally initialzing the machine for deployment."
	case ACTIVE + "_boot":
		return "Waiting for the hypervisor to create the machine"
	case ACTIVE + "_prolog":
		return "Transferring the disk images the host in which the machine will be running"
	case RESETPASSWORD:
		return "Machine root password updating"
	case IMAGE_SAVING:
		return "Saving marketplaces image"
	case IMAGE_SAVED:
		return "Marketplaces image saved successfully"
	case IMAGE_READY:
		return "Marketplaces image ready to publish"
	case CONTAINERERROR:
		return error_common
	case ERROR:
		return error_common
	case PREERROR:
		return name
	case NETWORK_UNAVAIL:
		return name
	case POST_ERROR:
		return error_common
	default:
		return "arrgh"
	}
}
