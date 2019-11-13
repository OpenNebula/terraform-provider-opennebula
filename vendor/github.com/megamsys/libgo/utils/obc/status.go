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

package obc

import (
	"strings"
)

// Status represents the status of a unit in vertice
type Status string

func (s Status) String() string {
	return string(s)
}

type State string

func (s State) String() string {
	return string(s)
}

// **************************************
//find some generic way to get a string instead of switch case
//*************************************

func (s Status) Event_type() string {
	switch s.String() {
	case TPL_HOSTINFOS_RUN:
		return OBCHOSTINFOSTPLRUN
	case TPL_HOSTINFOS_FINISHED:
		return OBCHOSTINFOSTPLFINISHED
	case CMD_GETCONFIG_BEGIN:
		return OBCHOSTINFOSCONFIGPARSING
	case CMD_GETCONFIG_END:
		return OBCHOSTINFOSCONFIGPARSED
	case CMD_KVMCHECK_BEGIN:
		return OBCHOSTINFOSKVMCHECKPARSING
	case CMD_KVMCHECK_END:
		return OBCHOSTINFOSKVMCHECKPARSED
	case TPL_VERTICE_RUN:
		return OBCVERTICETPLRUN
	case VERTICE_INSTALL_BEGIN:
		return OBCVERTICEINSTALLINGTYPE
	case VERTICE_INSTALL_END:
		return OBCVERTICEINSTALLEDTYPE
	case TPL_VERTICE_FINISHED:
		return OBCVERTICETPLFINISHED
	case TPL_VERTICEGATEWAY_RUN:
		return OBCVERTICEGATEWAYTPLRUN
	case GATEWAY_INSTALL_BEGIN:
		return OBCVERTICEGATEWAYINSTALLINGTYPE
	case GATEWAY_INSTALL_END:
		return OBCVERTICEGATEWAYINSTALLEDTYPE
	case TPL_VERTICEGATEWAY_FINISHED:
		return OBCVERTICEGATEWAYTPLFINISHED
	case TPL_VERTICENILAVU_RUN:
		return OBCVERTICENILAVUTPLRUN
	case NILAVU_INSTALL_BEGIN:
		return OBCVERTICENILAVUINSTALLINGTYPE
	case NILAVU_INSTALL_END:
		return OBCVERTICENILAVUINSTALLEDTYPE
	case TPL_VERTICENILAVU_FINISHED:
		return OBCVERTICENILAVUTPLFINISHED
	case TPL_ONEMASTER_RUN:
		return OBCONEMASTERTPLRUN
	case ONEMASTER_INSTALL_BEGIN:
		return OBCONEMASTERINSTALLINGTYPE
	case ONEMASTER_INSTALL_END:
		return OBCONEMASTERINSTALLEDTYPE
	case ONEMASTER_ACTIVATE_BEGIN:
		return OBCONEMASTERACTIVATINGTYPE
	case ONEMASTER_ACTIVATE_END:
		return OBCONEMASTERACTIVATEDTYPE
	case TPL_ONEMASTER_FINISHED:
		return OBCONEMASTERTPLFINISHED
	case TPL_ONEHOST_RUN:
		return OBCONEHOSTTPLRUN
	case ONEHOST_PEPARE_BEGIN:
		return OBCONEHOSTPEPARINGTYPE
	case ONEHOST_PREPARE_END:
		return OBCONEHOST_PREPAREDTYPE
	case ONEHOST_INSTALL_BEGIN:
		return OBCONEHOSTINSTALLINGTYPE
	case ONEHOST_INSTALL_END:
		return OBCONEHOSTINSTALLEDTYPE
	case TPL_ONEHOST_FINISHED:
		return OBCONEHOSTTPLFINISHEDTYPE
	case TPL_CEPHCLUSTER_RUN:
		return OBCCEPHCLUSTERTPLRUN
	case CEPHCLUSTER_PREPARE_BEGIN:
		return OBCCEPHCLUSTERPREPAREINSTALLINGTYPE
	case CEPHCLUSTER_PREPARE_END:
		return OBCCEPHCLUSTERPREPAREINSTALLEDTYPE
	case CEPHCLUSTER_INSTALL_BEIGIN:
		return OBCCEPHCLUSTERINSTALLINGTYPE
	case CEPHCLUSTER_INSTALL_END:
		return OBCCEPHCLUSTERINSTALLEDTYPE
	case CEPHCLUSTER_ACCESS_BEGIN:
		return OBCCEPHCLUSTERACCESSENABLINGTYPE
	case CEPHCLUSTER_ACCESS_END:
		return OBCCEPHCLUSTERACCESSENABLEDTYPE
	case CEPHCLUSTER_NEW_BEGIN:
		return OBCCEPHCLUSTERNEWCREATINGTYPE
	case CEPHCLUSTER_NEW_END:
		return OBCCEPHCLUSTERNEWCREATEDTYPE
	case CEPHCLUSTER_CONFIG_BEGIN:
		return OBCCEPHCLUSTERCONFIGURINGTYPE
	case CEPHCLUSTER_CONFIG_END:
		return OBCCEPHCLUSTERCONFIGUREDTYPE
	case CEPHCLUSTER_MON_BEGIN:
		return OBCCEPHCLUSTERMONINITIALINGTYPE
	case CEPHCLUSTER_MON_END:
		return OBCCEPHCLUSTERMONINITIALIZEDTYPE
	case CEPHCLUSTER_CREATEPOOL_BEGIN:
		return OBCCEPHCLUSTERPOOLCREATINGTYPE
	case CEPHCLUSTER_CREATEPOOL_END:
		return OBCCEPHCLUSTERPOOLCREATEDTYPE
	case TPL_CEPHCLUSTER_FINISHED:
		return OBCCEPHCLUSTERTPLFINISHED
	case TPL_CEPHCLIENT_RUN:
		return OBCCEPHCLIENTTPLRUN
	case CEPHCLIENT_INSTALL_BEGIN:
		return OBCCEPHCLIENTINSTALLING
	case CEPHCLIENT_INSTALL_END:
		return OBCCEPHCLIENTINSTALLED
	case TPL_CEPHCLIENT_FINISHED:
		return OBCCEPHCLIENTTPLFINISHED
	case TPL_ADDOSDS_RUN:
		return OBCADDOSDSTPLRUN
	case OSDS_ACTIVATE_BEGIN:
		return OBCOSDSACTIVATINGTYPE
	case OSDS_ACTIVATE_END:
		return OBCOSDSACTIVATEDTYPE
	case OSDS_PREPARE_BEGIN:
		return OBCOSDSPREPARINGTYPE
	case OSDS_PREPARE_END:
		return OBCOSDSPREPAREDTYPE
	case TPL_ADDOSDS_FINISHED:
		return OBCADDOSDSTPLFINISHED
	case TPL_DSCONNECTION_RUN:
		return OBCCEPHDSCONNECTIONTPLRUN
	case CEPHDS_AUTHKEY_BEGIN:
		return OBCCEPHDSAUTHKEYCREATINGTYPE
	case CEPHDS_AUTHKEY_END:
		return OBCCEPHDSAUTHKEYCREATEDTYPE
	case CEPHDS_DEFINEKEY_BEGIN:
		return OBCCEPHDSAUTHKEYDEFININGTYPE
	case CEPHDS_DEFINEKEY_END:
		return OBCCEPHDSAUTHKEYDEFINEDTYPE
	case TPL_DSCONNECTION_FINISHED:
		return OBCCEPHDSCONNECTIONTPLFINISHED
	case TPL_CEPHACCESS_RUN:
		return OBCCEPHACCESSTPLRUN
	case CEPHACCESS_PASSWORD_BEGIN:
		return OBCCEPHACCESSPASSAUTHENDICATINGTYPE
	case CEPHACCESS_PASSWORD_END:
		return OBCCEPHACCESSPASSAUTHENDICATEDTYPE
	case CEPHACCESS_KEY_BEGIN:
		return OBCCEPHACCESSKEYAUTHENDICATINGTYPE
	case CEPHACCESS_KEY_END:
		return OBCCEPHACCESSKEYAUTHENDICATEDTYPE
	case TPL_CEPHACCESS_FINISHED:
		return OBCCEPHACCESSTPLFINISHED
	case TPL_ZAPDISK_RUN:
		return OBCZAPDISKTPLRUN
	case ZAPDISKS_CLEAN_BEGIN:
		return OBCZAPDISKSCLEANINGTYPE
	case ZAPDISKS_CLEAN_END:
		return OBCZAPDISKSCLEANEDTYPE
	case TPL_ZAPDISK_FINISHED:
		return OBCZAPDISKTPLFINISHED
	case TPL_KVMNETWORK_RUN:
		return OBCKVMNETWORKTPLRUN
	case KVMNETWORK_CONFIG_BEGIN:
		return OBCKVMNETWORKCONFIGURINGTYPE
	case KVMNETWORK_CONFIG_END:
		return OBCKVMNETWORKCONFIGUREDTYPE
	case TPL_KVMNETWORK_FINISHED:
		return OBCKVMNETWORKTPLFINISHED
	case TPL_LVMINSTALL_RUN:
		return OBCLVMTPLRUN
	case LVM_INSTALL_BEGIN:
		return OBCLVM_INSTALLINGTYPE
	case LVM_INSTALL_END:
		return OBCLVMINSTALLEDTYPE
	case TPL_LVMINSTALL_FINISHED:
		return OBCLVMTPLFINISHED
	case TPL_NETWORKINFO_RUN:
		return OBCNETWORKINFOTPLRUN
	case GETNETWORK_INFOS_BEGIN:
		return OBCGETNETWORKINFOSGATHERINGTYPE
	case GETNETWORK_INFOS_END:
		return OBCGETNETWORKINFOSGATHEREDTYPE
	case TPL_NETWORKINFO_FINISHED:
		return OBCNETWORKINFOTPLFINISHED
	case RUNNING:
		return OBCHOSTINSTALLRUNNINGTYPE
	case HOSTRUNNING:
		return OBCHOSTINSTALLRUNNINGTYPE
	case MASTERRUNNING:
		return OBCHOSTINSTALLRUNNINGTYPE
	case NETWORKERROR:
		return OBCNETWORKERRORTYPE
	case STORAGEERROR:
		return OBCSTORAGEERRORTYPE
	case COMPUTEERROR:
		return OBCCOMPUTEERRORTYPE
	case PREERROR:
		return OBCTEMPLATEPREERRORTYPE
	case POSTERROR:
		return OBCTEMPLATEPOSTERRORTYPE
	case ERROR:
		return OBCTEPMLATEERRORTYPE
	default:
		return "arrgh"
	}
}

func (s Status) Description(host string) string {
	tmpl := strings.ToUpper(strings.Split(s.String(), ".")[0])
	TPL_RUN := "Host " + host + " template [" + tmpl + "] running.. "
	TPL_FINISHED := "Host " + host + " template [" + tmpl + "] finished. "
	switch s.String() {
	case TPL_HOSTINFOS_RUN:
		return TPL_RUN
	case TPL_HOSTINFOS_FINISHED:
		return TPL_FINISHED
	case CMD_GETCONFIG_BEGIN:
		return "Host " + host + " configuration gathering.. "
	case CMD_GETCONFIG_END:
		return "Host " + host + " configuration gathered. "
	case CMD_KVMCHECK_BEGIN:
		return "Host " + host + " kvm checking for node setup"
	case CMD_KVMCHECK_END:
		return "Host " + host + " kvm verified for node setup"
	case TPL_VERTICE_RUN:
		return TPL_RUN
	case VERTICE_INSTALL_BEGIN:
		return "Host " + host + " vertice package installing"
	case VERTICE_INSTALL_END:
		return "Host " + host + " vertice package installed"
	case TPL_VERTICE_FINISHED:
		return TPL_FINISHED
	case TPL_VERTICEGATEWAY_RUN:
		return TPL_RUN
	case GATEWAY_INSTALL_BEGIN:
		return "Host " + host + " verticegateway package installing"
	case GATEWAY_INSTALL_END:
		return "Host " + host + " verticegateway package installed"
	case TPL_VERTICEGATEWAY_FINISHED:
		return TPL_FINISHED
	case TPL_VERTICENILAVU_RUN:
		return TPL_RUN
	case NILAVU_INSTALL_BEGIN:
		return "Host " + host + " verticenilavu package installing"
	case NILAVU_INSTALL_END:
		return "Host " + host + " verticegateway package installed"
	case TPL_VERTICENILAVU_FINISHED:
		return TPL_FINISHED
	case TPL_ONEMASTER_RUN:
		return TPL_RUN
	case ONEMASTER_INSTALL_BEGIN:
		return "Host " + host + " opennebula master (front end) installing"
	case ONEMASTER_INSTALL_END:
		return "Host " + host + " opennebula master (front end) installed"
	case ONEMASTER_ACTIVATE_BEGIN:
		return "Host " + host + " one master configure changes and activating"
	case ONEMASTER_ACTIVATE_END:
		return "Host " + host + " one master configure changes and activated"
	case TPL_ONEMASTER_FINISHED:
		return TPL_FINISHED
	case TPL_ONEHOST_RUN:
		return TPL_RUN
	case ONEHOST_PEPARE_BEGIN:
		return "Host " + host + " opennebula node dependancy packages installing"
	case ONEHOST_PREPARE_END:
		return "Host " + host + " opennebula node dependancy packages installed"
	case ONEHOST_INSTALL_BEGIN:
		return "Host " + host + " opennebula node installing"
	case ONEHOST_INSTALL_END:
		return "Host " + host + " opennebula node installed"
	case TPL_ONEHOST_FINISHED:
		return TPL_FINISHED
	case TPL_CEPHCLUSTER_RUN:
		return TPL_RUN
	case CEPHCLUSTER_PREPARE_BEGIN:
		return "Host " + host + " ceph cluster dependancy packages installing"
	case CEPHCLUSTER_PREPARE_END:
		return "Host " + host + " ceph cluster dependancy packages installed"
	case CEPHCLUSTER_INSTALL_BEIGIN:
		return "Host " + host + " ceph-deploy packages installing"
	case CEPHCLUSTER_INSTALL_END:
		return "Host " + host + " ceph-deploy packages installed"
	case CEPHCLUSTER_ACCESS_BEGIN:
		return "Host " + host + " ceph cluster user access configuring"
	case CEPHCLUSTER_ACCESS_END:
		return "Host " + host + " ceph cluster user access configuring"
	case CEPHCLUSTER_NEW_BEGIN:
		return "Host " + host + " ceph new cluster creating"
	case CEPHCLUSTER_NEW_END:
		return "Host " + host + " ceph new cluster created"
	case CEPHCLUSTER_CONFIG_BEGIN:
		return "Host " + host + " ceph cluster configuration changing"
	case CEPHCLUSTER_CONFIG_END:
		return "Host " + host + " ceph cluster configuration changed"
	case CEPHCLUSTER_MON_BEGIN:
		return "Host " + host + " ceph monitor initializing"
	case CEPHCLUSTER_MON_END:
		return "Host " + host + " ceph monitor initialized"
	case CEPHCLUSTER_CREATEPOOL_BEGIN:
		return "Host " + host + " ceph cluster storage pool creating"
	case CEPHCLUSTER_CREATEPOOL_END:
		return "Host " + host + " ceph cluster storage pool created"
	case TPL_CEPHCLUSTER_FINISHED:
		return TPL_FINISHED
	case TPL_CEPHCLIENT_RUN:
		return TPL_RUN
	case CEPHCLIENT_INSTALL_BEGIN:
		return "Host " + host + " ceph client packages installing"
	case CEPHCLIENT_INSTALL_END:
		return "Host " + host + " ceph client packages installed"
	case TPL_CEPHCLIENT_FINISHED:
		return TPL_FINISHED
	case TPL_ADDOSDS_RUN:
		return TPL_RUN
	case OSDS_ACTIVATE_BEGIN:
		return "Host " + host + " activating add osds to ceph cluster"
	case OSDS_ACTIVATE_END:
		return "Host " + host + " activated add osds to ceph cluster"
	case OSDS_PREPARE_BEGIN:
		return "Host " + host + " preparing add osds to ceph cluster"
	case OSDS_PREPARE_END:
		return "Host " + host + " prepared add osds to ceph cluster"
	case TPL_ADDOSDS_FINISHED:
		return TPL_FINISHED
	case TPL_DSCONNECTION_RUN:
		return TPL_RUN
	case CEPHDS_AUTHKEY_BEGIN:
		return "Host " + host + " ceph datastore user auth key creating"
	case CEPHDS_AUTHKEY_END:
		return "Host " + host + " ceph datastore user auth key created"
	case CEPHDS_DEFINEKEY_BEGIN:
		return "Host " + host + " secrekey defining to libvert"
	case CEPHDS_DEFINEKEY_END:
		return "Host " + host + " secrekey defined to libvert"
	case TPL_DSCONNECTION_FINISHED:
		return TPL_FINISHED
	case TPL_CEPHACCESS_RUN:
		return TPL_RUN
	case CEPHACCESS_PASSWORD_BEGIN:
		return "Host " + host + " sharing public key to client through client password"
	case CEPHACCESS_PASSWORD_END:
		return "Host " + host + " shared public key to client through client password"
	case CEPHACCESS_KEY_BEGIN:
		return "Host " + host + " sharing public key to client through client private key"
	case CEPHACCESS_KEY_END:
		return "Host " + host + " shared public key to client through client private key"
	case TPL_CEPHACCESS_FINISHED:
		return TPL_FINISHED
	case TPL_ZAPDISK_RUN:
		return TPL_RUN
	case ZAPDISKS_CLEAN_BEGIN:
		return "Host " + host + " ceph-deploy disk cleaning"
	case ZAPDISKS_CLEAN_END:
		return "Host " + host + " ceph-deploy disk cleaned"
	case TPL_ZAPDISK_FINISHED:
		return TPL_FINISHED
	case TPL_KVMNETWORK_RUN:
		return TPL_RUN
	case KVMNETWORK_CONFIG_BEGIN:
		return "Host " + host + " kvm network configuring"
	case KVMNETWORK_CONFIG_END:
		return "Host " + host + " kvm network configured"
	case TPL_KVMNETWORK_FINISHED:
		return TPL_FINISHED
	case TPL_LVMINSTALL_RUN:
		return TPL_RUN
	case LVM_INSTALL_BEGIN:
		return "Host " + host + " lvm packages installing"
	case LVM_INSTALL_END:
		return "Host " + host + " lvm packages installed"
	case TPL_LVMINSTALL_FINISHED:
		return TPL_FINISHED
	case TPL_NETWORKINFO_RUN:
		return TPL_RUN
	case GETNETWORK_INFOS_BEGIN:
		return "Host " + host + " getting network informations"
	case GETNETWORK_INFOS_END:
		return "Host " + host + " network informations collected"
	case TPL_NETWORKINFO_FINISHED:
		return TPL_FINISHED
	case RUNNING:
		return "Server " + host + " installed successfully"
	case HOSTRUNNING:
		return "Host " + host + " installed successfully"
	case MASTERRUNNING:
		return "Master " + host + " installed successfully"
	default:
		return "arrgh"
	}
}
