package opennebula

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	errs "github.com/OpenNebula/one/src/oca/go/src/goca/errors"
	vn "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualnetwork"
)

type vnetTemplate struct {
	Description     string `xml:"DESCRIPTION,omitempty"`
	Security_Groups string `xml:"SECURITY_GROUPS,omitempty"`
	Mtu             int    `xml:"MTU,omitempty"`
	Dns             string `xml:"DNS,omitempty"`
	Gateway         string `xml:"GATEWAY,omitempty"`
	Network_Mask    string `xml:"NETWORK_MASK,omitempty"`
	GuestMtu        int    `xml:"GUEST_MTU,omitempty"`
}

func resourceOpennebulaVirtualNetwork() *schema.Resource {
	return &schema.Resource{
		Create: resourceOpennebulaVirtualNetworkCreate,
		Read:   resourceOpennebulaVirtualNetworkRead,
		Exists: resourceOpennebulaVirtualNetworkExists,
		Update: resourceOpennebulaVirtualNetworkUpdate,
		Delete: resourceOpennebulaVirtualNetworkDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the vnet",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the vnet, in OpenNebula's XML or String format",
			},
			"permissions": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Permissions for the vnet (in Unix format, owner-group-other, use-manage-admin)",
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					value := v.(string)

					if len(value) != 3 {
						errors = append(errors, fmt.Errorf("%q has specify 3 permission sets: owner-group-other", k))
					}

					all := true
					for _, c := range strings.Split(value, "") {
						if c < "0" || c > "7" {
							all = false
						}
					}
					if !all {
						errors = append(errors, fmt.Errorf("Each character in %q should specify a Unix-like permission set with a number from 0 to 7", k))
					}

					return
				},
			},
			"uid": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the user that will own the vnet",
			},
			"gid": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"group"},
				Description:   "ID of the group that will own the vnet",
			},
			"uname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the user that will own the vnet",
			},
			"gname": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Name of the group that will own the vnet",
			},
			"bridge": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Name of the bridge interface to which the vnet should be associated",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"physical_device": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Name of the physical device to which the vnet should be associated",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"type": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "bridge",
				Description:   "Type of the Virtual Network: dummy, bridge, fw, ebtables, 802.1Q, vxlan, ovswitch. Default is 'bridge'",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
				ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
					validtypes := []string{"dummy", "bridge", "fw", "ebtables", "802.1Q", "vxlan", "ovswitch"}
					value := v.(string)

					if inArray(value, validtypes) < 0 {
						errors = append(errors, fmt.Errorf("Type %q must be one of: %s", k, strings.Join(validtypes, ",")))
					}

					return
				},
			},
			"clusters": {
				Type:          schema.TypeList,
				Optional:      true,
				Description:   "List of cluster IDs hosting the virtual Network, if not set it uses the default cluster",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"vlan_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "VLAN ID. Only if 'Type' is : 802.1Q, vxlan or ovswich and if 'automatic_vlan_id' is not set",
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "automatic_vlan_id"},
			},
			"automatic_vlan_id": {
				Type:          schema.TypeBool,
				Optional:      true,
				Description:   "If set, let OpenNebula to attribute VLAN ID",
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "vlan_id"},
			},
			"mtu": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "MTU of the vnet (defaut: 1500)",
				Default:       1500,
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"guest_mtu": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "MTU of the Guest interface. Must be lower or equal to 'mtu' (defaut: 1500)",
				Default:       1500,
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"gateway": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Gateway IP if necessary",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"network_mask": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Network Mask",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"dns": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "DNS IP if necessary",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"ar": {
				Type:          schema.TypeSet,
				Optional:      true,
				MinItems:      1,
				Description:   "List of Address Ranges to be part of the Virtual Network",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"ar_type": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "IP4",
							Description: "Type of the Address Range: IP4, IP6. Default is 'IP4'",
							ValidateFunc: func(v interface{}, k string) (ws []string, errors []error) {
								validtypes := []string{"IP4", "IP6", "IP6_STATIC", "IP4_6", "IP4_6_STATIC", "ETHER"}
								value := v.(string)

								if inArray(value, validtypes) < 0 {
									errors = append(errors, fmt.Errorf("Address Range type %q must be one of: %s", k, strings.Join(validtypes, ",")))
								}

								return
							},
						},
						"ip4": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Start IPv4 of the range to be allocated (Required if IP4 or IP4_6).",
						},
						"size": {
							Type:        schema.TypeInt,
							Required:    true,
							Description: "Size (in number) of the ip range",
						},
						"ip6": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Start IPv6 of the range to be allocated (Required if IP6_STATIC or IP4_6_STATIC)",
						},
						"mac": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Start MAC of the range to be allocated",
						},
						"global_prefix": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Global prefix for IP6 or IP4_6",
						},
						"ula_prefix": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "ULA prefix for IP6 or IP4_6",
						},
						"prefix_length": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Prefix lenght Only needed for IP6_STATIC or IP4_6_STATIC",
						},
					},
				},
			},
			"hold_size": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "Carve a network reservation of this size from the reservation starting from `ip_hold`",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"ip_hold": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Start IP of the range to be held",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"reservation_vnet": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "Create a reservation from this VNET ID",
				ConflictsWith: []string{"bridge", "physical_device", "ar", "hold_size", "ip_hold", "type", "vlan_id", "automatic_vlan_id", "mtu", "clusters", "dns", "gateway", "network_mask"},
			},
			"reservation_size": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "Reserve this many IPs from reservation_vnet",
				ConflictsWith: []string{"bridge", "physical_device", "ar", "hold_size", "ip_hold", "type", "vlan_id", "automatic_vlan_id", "mtu", "clusters", "dns", "gateway", "network_mask"},
			},
			"security_groups": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "List of Security Group IDs to be applied to the VNET",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"group": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"gid"},
				Description:   "Name of the Group that onws the Virtual Network, If empty, it uses caller group",
			},
		},
	}
}

func getVirtualNetworkController(d *schema.ResourceData, meta interface{}, args ...int) (*goca.VirtualNetworkController, error) {
	controller := meta.(*goca.Controller)
	var vnc *goca.VirtualNetworkController

	// Try to find the VNet by ID, if specified
	if d.Id() != "" {
		id, err := strconv.ParseUint(d.Id(), 10, 64)
		if err != nil {
			return nil, err
		}
		vnc = controller.VirtualNetwork(int(id))
	}

	// Otherwise, try to find the VNet by name as the de facto compound primary key
	if d.Id() == "" {
		id, err := controller.VirtualNetworks().ByName(d.Get("name").(string), args...)
		if err != nil {
			return nil, err
		}
		vnc = controller.VirtualNetwork(id)
	}

	return vnc, nil
}

func changeVNetGroup(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)
	var gid int

	vnc, err := getVirtualNetworkController(d, meta)
	if err != nil {
		return err
	}

	if d.Get("group") != "" {
		gid, err = controller.Groups().ByName(d.Get("group").(string))
		if err != nil {
			return err
		}
	} else {
		gid = d.Get("gid").(int)
	}

	err = vnc.Chown(-1, gid)
	if err != nil {
		return err
	}

	return nil
}

func validVlanType(intype string) int {
	vlanType := []string{"802.1Q", "vxlan", "ovswitch"}
	return inArray(intype, vlanType)
}

func resourceOpennebulaVirtualNetworkCreate(d *schema.ResourceData, meta interface{}) error {
	controller := meta.(*goca.Controller)
	var vnc *goca.VirtualNetworkController

	//VNET reservation
	if rvnet, ok := d.GetOk("reservation_vnet"); ok {
		reservation_vnet := rvnet.(int)
		reservation_name := d.Get("name").(string)
		reservation_size := d.Get("reservation_size").(int)

		if reservation_vnet <= 0 {
			return fmt.Errorf("Reservation VNET ID must be greater than 0!")
		} else if reservation_size <= 0 {
			return fmt.Errorf("Reservation size must be greater than 0!")
		}

		//The API only takes ATTRIBUTE=VALUE for VNET reservations...
		reservation_string := "SIZE=%d\nNAME=\"%s\""

		// Get VNet Controller to reserve from
		vnc = controller.VirtualNetwork(reservation_vnet)

		rID, err := vnc.Reserve(fmt.Sprintf(reservation_string, reservation_size, reservation_name))
		if err != nil {
			return err
		}

		vnc = controller.VirtualNetwork(rID)

		// TODO: fix it after 5.10 release
		// Force the "decrypt" bool to false to keep ONE 5.8 behavior
		vnet, err := vnc.Info(false)
		if err != nil {
			return err
		}

		d.SetId(fmt.Sprintf("%v", vnet.ID))

		log.Printf("[DEBUG] New VNET reservation ID: %d", vnet.ID)

	} else { //New VNET
		vntpl, err := generateVnXML(d)
		if err != nil {
			return err
		}

		// Create VNet
		vnetID, err := controller.VirtualNetworks().Create(vntpl, -1)
		if err != nil {
			return err
		}
		vnc = controller.VirtualNetwork(vnetID)

		d.SetId(fmt.Sprintf("%v", vnetID))

		if d.Get("hold_size").(int) > 0 {
			// add address range and reservations
			ip := net.ParseIP(d.Get("ip_start").(string))
			ip = ip.To4()

			for i := 0; i < d.Get("hold_size").(int); i++ {
				var address_reservation_string = `LEASES=[IP=%s]`
				r_err := vnc.Hold(fmt.Sprintf(address_reservation_string, ip))
				if r_err != nil {
					return r_err
				}

				ip[3]++
			}
		}

		// Call API once
		update, err := generateVnTemplate(d)
		if err != nil {
			return err
		}
		err = vnc.Update(update, 1)
		if err != nil {
			return err
		}

		// Address Ranges
		ars := d.Get("ar").(*schema.Set).List()

		for i, arinterface := range ars {
			armap := arinterface.(map[string]interface{})
			arstr := generateAR(armap, i)
			err := vnc.AddAR(arstr)
			if err != nil {
				return fmt.Errorf("Error: %s\nAR: %s", err, arstr)
			}
		}

		// Set Clusters
		if _, ok := d.GetOk("clusters"); ok {
			err := setVnetClusters(d, meta, vnetID)
			if err != nil {
				return err
			}
		}
	}

	// Set Security Groups
	if securitygroups, ok := d.GetOk("security_groups"); ok {
		secgrouplist := ArrayToString(securitygroups.([]interface{}), ",")

		err := vnc.Update(fmt.Sprintf("SECURITY_GROUPS=\"%s\"", secgrouplist), 1)
		if err != nil {
			return err
		}
	}
	// update permisions
	if perms, ok := d.GetOk("permissions"); ok {
		err := vnc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			log.Printf("[ERROR] template permissions change failed, error: %s", err)
			return err
		}
	}

	if d.Get("group") != "" || d.Get("gid") != "" {
		err := changeVNetGroup(d, meta)
		if err != nil {
			return err
		}
	}

	return resourceOpennebulaVirtualNetworkRead(d, meta)
}

func generateAR(armap map[string]interface{}, id int) string {
	// Generate AR depending on the AR Type
	artype := armap["ar_type"].(string)
	arip4 := armap["ip4"].(string)
	arip6 := armap["ip6"].(string)
	armac := armap["mac"].(string)
	arsize := armap["size"].(int)
	argprefix := armap["global_prefix"].(string)
	arulaprefix := armap["ula_prefix"].(string)
	arprefixlength := armap["prefix_length"].(string)
	arid := strconv.Itoa(id)

	if artype == "IP4" {
		if armac != "" {
			var arstring = `AR = [
                AR_ID = %s,
		        TYPE = IP4,
		        IP = %s,
		        MAC = %s,
		        SIZE = %d ]`
			return fmt.Sprintf(arstring, arid, arip4, armac, arsize)
		}
		var arstring = `AR = [
            AR_ID = %s,
            TYPE = IP4,
            IP = %s,
            SIZE = %d ]`
		return fmt.Sprintf(arstring, arid, arip4, arsize)
	}
	if artype == "IP6" {
		if armac != "" {
			if argprefix != "" {
				if arulaprefix != "" {
					var arstring = `AR = [
                        AR_ID = %s,
		                TYPE = IP6,
		                MAC = %s,
		                GLOBAL_PREFIX = %s,
		                ULA_PREFIX = %s,
		                SIZE = %d ]`
					return fmt.Sprintf(arstring, arid, armac, argprefix, arulaprefix, arsize)
				}
				var arstring = `AR = [
                    AR_ID = %s,
		            TYPE = IP6,
		            MAC = %s,
		            GLOBAL_PREFIX = %s,
		            SIZE = %d ]`
				return fmt.Sprintf(arstring, arid, armac, argprefix, arsize)
			}
			var arstring = `AR = [
                AR_ID = %s,
		        TYPE = IP6,
		        MAC = %s,
		        SIZE = %d ]`
			return fmt.Sprintf(arstring, arid, armac, arsize)
		}
		var arstring = `AR = [
            AR_ID = %s,
		    TYPE = IP6,
		    SIZE = %d ]`
		return fmt.Sprintf(arstring, arid, arsize)
	}
	if artype == "IP6_STATIC" {
		if armac != "" {
			var arstring = `AR = [
                AR_ID = %s,
		        TYPE = IP6_STATIC,
		        MAC = %s,
		        IP6 = %s,
		        PREFIX_LENGTH = %s,
		        SIZE = %d ]`
			return fmt.Sprintf(arstring, arid, armac, arip6, arprefixlength, arsize)
		}
		var arstring = `AR = [
            AR_ID = %s,
		    TYPE = IP6_STATIC,
		    IP6 = %s,
		    PREFIX_LENGTH = %s,
		    SIZE = %d ]`
		return fmt.Sprintf(arstring, arid, arip6, arprefixlength, arsize)
	}
	if artype == "IP4_6" {
		if armac != "" {
			if argprefix != "" {
				if arulaprefix != "" {
					var arstring = `AR = [
                        AR_ID = %s,
		                TYPE = IP4_6,
		                IP = %s,
		                MAC = %s,
		                GLOBAL_PREFIX = %s,
		                ULA_PREFIX = %s,
		                SIZE = %d ]`
					return fmt.Sprintf(arstring, arid, arip4, armac, argprefix, arulaprefix, arsize)
				}
				var arstring = `AR = [
                    AR_ID = %s,
		            TYPE = IP4_6,
		            IP = %s,
		            MAC = %s,
		            GLOBAL_PREFIX = %s,
		            SIZE = %d ]`
				return fmt.Sprintf(arstring, arid, arip4, armac, argprefix, arsize)
			}
			var arstring = `AR = [
                AR_ID = %s,
		        TYPE = IP4_6,
		        IP = %s,
		        MAC = %s,
		        SIZE = %d ]`
			return fmt.Sprintf(arstring, arid, arip4, armac, arsize)
		}
		var arstring = `AR = [
            AR_ID = %s,
		    TYPE = IP4_6,
		    IP = %s,
		    SIZE = %d ]`
		return fmt.Sprintf(arstring, arid, arip4, arsize)
	}
	if artype == "IP4_6_STATIC" {
		if armac != "" {
			var arstring = `AR = [
                AR_ID = %s,
		        TYPE = IP4_6_STATIC,
		        IP = %s,
		        MAC = %s,
		        IP6 = %s,
		        PREFIX_LENGTH = %s,
		        SIZE = %d ]`
			return fmt.Sprintf(arstring, arid, arip4, armac, arip6, arprefixlength, arsize)
		}
		var arstring = `AR = [
            AR_ID = %s,
		    TYPE = IP4_6_STATIC,
		    IP = %s,
		    IP6 = %s,
		    PREFIX_LENGTH = %s,
		    SIZE = %d ]`
		return fmt.Sprintf(arstring, arid, arip4, arip6, arprefixlength, arsize)
	}
	if artype == "ETHER" {
		if armac != "" {
			var arstring = `AR = [
                AR_ID = %s,
		        TYPE = ETHER,
		        MAC = %s,
		        SIZE = %d ]`
			return fmt.Sprintf(arstring, arid, armac, arsize)
		}
		var arstring = `AR = [
            AR_ID = %s,
		    TYPE = ETHER,
		    SIZE = %d ]`
		return fmt.Sprintf(arstring, arid, arsize)
	}

	return ""
}

func generateVnTemplate(d *schema.ResourceData) (string, error) {
	mtu := d.Get("mtu").(int)
	dns := d.Get("dns").(string)
	gateway := d.Get("gateway").(string)
	netmask := d.Get("network_mask").(string)
	description := d.Get("description").(string)
	guestmtu := d.Get("guest_mtu").(int)

	if guestmtu > mtu {
		return "", fmt.Errorf("Invalid: Guest MTU (%v) is greater than MTU (%v)", guestmtu, mtu)
	}

	vntpl := &vnetTemplate{
		Description:  description,
		Mtu:          mtu,
		GuestMtu:     guestmtu,
		Dns:          dns,
		Gateway:      gateway,
		Network_Mask: netmask,
	}

	w := &bytes.Buffer{}

	//Encode the VN template schema to XML
	enc := xml.NewEncoder(w)
	//enc.Indent("", "  ")
	if err := enc.Encode(vntpl); err != nil {
		return "", err
	}

	log.Printf("template XML: %s", w.String())
	return w.String(), nil
}

func generateVnXML(d *schema.ResourceData) (string, error) {
	vnname := d.Get("name").(string)
	vnmad := d.Get("type").(string)
	vnbridge := d.Get("bridge").(string)
	vnphydev := d.Get("physical_device").(string)

	if vnmad == "" {
		vnmad = "bridge"
	}
	vnautovlan := "0"
	var vnvlan string

	if validVlanType(vnmad) >= 0 {
		if d.Get("automatic_vlan_id") == true {
			vnautovlan = "1"
		} else if vlanid, ok := d.GetOk("vlan_id"); ok {
			vnvlan = vlanid.(string)
		} else {
			return "", fmt.Errorf("You must specify a 'vlan_id' or set the flag 'automatic_vlan_id'")
		}
	}

	vntpl := &vn.VirtualNetwork{
		Name:            vnname,
		Bridge:          vnbridge,
		PhyDev:          vnphydev,
		VNMad:           vnmad,
		VlanIDAutomatic: vnautovlan,
		VlanID:          vnvlan,
	}

	w := &bytes.Buffer{}

	//Encode the VN template schema to XML
	enc := xml.NewEncoder(w)
	//enc.Indent("", "  ")
	if err := enc.Encode(vntpl); err != nil {
		return "", err
	}

	log.Printf("VNET XML: %s", w.String())
	return w.String(), nil
}

func setVnetClusters(d *schema.ResourceData, meta interface{}, id int) error {
	controller := meta.(*goca.Controller)
	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	clusterPool, err := controller.Clusters().Info(false)
	if err != nil {
		return err
	}
	clusters := d.Get("clusters").([]interface{})
	log.Printf("Number of clusters: %d", len(clusters))
	for i := 0; i < len(clusters); i++ {
		clusterid := clusters[i].(int)
		for j := 0; j < len(clusterPool.Clusters); j++ {
			if clusterid == clusterPool.Clusters[j].ID {
				cc := controller.Cluster(clusterPool.Clusters[j].ID)
				cc.AddVnet(id)
			}
		}
	}

	return nil
}

func resourceOpennebulaVirtualNetworkRead(d *schema.ResourceData, meta interface{}) error {
	vnc, err := getVirtualNetworkController(d, meta, -2, -1, -1)
	if err != nil {
		switch err.(type) {
		case *errs.ClientError:
			clientErr, _ := err.(*errs.ClientError)
			if clientErr.Code == errs.ClientRespHTTP {
				response := clientErr.GetHTTPResponse()
				if response.StatusCode == http.StatusNotFound {
					log.Printf("[WARN] Removing virtual network %s from state because it no longer exists in", d.Get("name"))
					d.SetId("")
					return nil
				}
			}
			return err
		default:
			return err
		}
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vn, err := vnc.Info(false)
	if err != nil {
		return err
	}

	d.SetId(strconv.Itoa(vn.ID))
	d.Set("name", vn.Name)
	d.Set("uid", vn.UID)
	d.Set("gid", vn.GID)
	d.Set("uname", vn.UName)
	d.Set("gname", vn.GName)
	d.Set("bridge", vn.Bridge)
	d.Set("physical_device", vn.PhyDev)
	if vn.VlanID != "" {
		d.Set("vlan_id", vn.VlanID)
	}
	if vn.VlanIDAutomatic == "1" {
		d.Set("automatic_vlan_id", true)
	}
	d.Set("type", vn.VNMad)
	d.Set("reservation_vnet", vn.ParentNetworkID)
	d.Set("permissions", permissionsUnixString(vn.Permissions))

	secgrouplist, err := vn.Template.Dynamic.GetContentByName("SECURITY_GROUPS")
	if err != nil {
		return err
	}
	secgroups_str := strings.Split(secgrouplist, ",")
	secgroups_int := []int{}

	for _, i := range secgroups_str {
		if i != "" {
			j, err := strconv.Atoi(i)
			if err != nil {
				return err
			}
			secgroups_int = append(secgroups_int, j)
		}
	}

	err = d.Set("security_groups", secgroups_int)
	if err != nil {
		log.Printf("[DEBUG] Error setting security groups on vnet: %s", err)
	}
	mtustr, _ := vn.Template.Dynamic.GetContentByName("MTU")
	if mtustr != "" {
		mtu, err := strconv.ParseInt(mtustr, 10, 64)
		if err != nil {
			return err
		}
		err = d.Set("mtu", mtu)
		if err != nil {
			return err
		}
	}
	desc, _ := vn.Template.Dynamic.GetContentByName("DESCRIPTION")
	if desc != "" {
		err = d.Set("description", fmt.Sprintf("%v", desc))
		if err != nil {
			return err
		}
	}

	if err := d.Set("ar", generateARMapFromStructs(vn.ARs)); err != nil {
		log.Printf("[WARN] Error setting ar for Virtual Network %x, error: %s", vn.ID, err)
	}
	return nil
}

func generateARMapFromStructs(slice []vn.AR) []map[string]interface{} {

	armap := make([]map[string]interface{}, 0)

	for i := 0; i < len(slice); i++ {
		armap = append(armap, structs.Map(slice[i]))
	}

	return armap
}

func resourceOpennebulaVirtualNetworkExists(d *schema.ResourceData, meta interface{}) (bool, error) {
	err := resourceOpennebulaVirtualNetworkRead(d, meta)
	if err != nil || d.Id() == "" {
		return false, err
	}

	return true, nil
}

func resourceOpennebulaVirtualNetworkUpdate(d *schema.ResourceData, meta interface{}) error {
	//Get Virtual Network Controller
	vnc, err := getVirtualNetworkController(d, meta)
	if err != nil {
		return err
	}

	if d.HasChange("description") {
		err := vnc.Update(fmt.Sprintf("DESCRIPTION =\"%s\"", d.Get("description").(string)), 1)
		if err != nil {
			return err
		}
	}

	if d.HasChange("gateway") {
		err := vnc.Update(fmt.Sprintf("GATEWAY = \"%s\"", d.Get("gateway").(string)), 1)
		if err != nil {
			return err
		}
	}

	if d.HasChange("dns") {
		err := vnc.Update(fmt.Sprintf("DNS = \"%s\"", d.Get("dns").(string)), 1)
		if err != nil {
			return err
		}
	}

	if d.HasChange("network_mask") {
		err := vnc.Update(fmt.Sprintf("NETWORK_MASK = \"%s\"", d.Get("network_mask").(string)), 1)
		if err != nil {
			return err
		}

	}

	if d.HasChange("security_groups") {
		securitygroups := d.Get("security_groups")
		secgrouplist := ArrayToString(securitygroups.([]interface{}), ",")

		err := vnc.Update(fmt.Sprintf("SECURITY_GROUPS=\"%s\"", secgrouplist), 1)
		if err != nil {
			return err
		}
	}

	if d.HasChange("name") {
		err := vnc.Rename(d.Get("name").(string))
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated name for Vnet\n")
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vn, err := vnc.Info(false)
	if err != nil {
		return err
	}

	if d.HasChange("permissions") {
		if perms, ok := d.GetOk("permissions"); ok {
			err = vnc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				return err
			}
		}
		log.Printf("[INFO] Successfully updated Vnet\n")
	}

	if d.HasChange("group") {
		err = changeVNetGroup(d, meta)
		if err != nil {
			return err
		}
		log.Printf("[INFO] Successfully updated group for Vnet %s\n", vn.Name)
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vn, err = vnc.Info(false)
	if err != nil {
		return err
	}

	if d.HasChange("ar") {
		_, narsset := d.GetChange("ar")

		nars := narsset.(*schema.Set).List()
		vnars := vn.ARs

		// Delete Old ARs
		for _, vnar := range vnars {
			arid, err := strconv.Atoi(vnar.ID)
			if err != nil {
				return err
			}
			err = vnc.RmAR(arid)
			if err != nil {
				return err
			}
		}

		// Add All ARs
		for i, ar := range nars {
			armap := ar.(map[string]interface{})

			arstr := generateAR(armap, i)
			err := vnc.AddAR(arstr)
			if err != nil {
				return fmt.Errorf("Error: %s\nAR: %s", err, arstr)
			}
		}

	}

	return nil
}

func resourceOpennebulaVirtualNetworkDelete(d *schema.ResourceData, meta interface{}) error {
	vnc, err := getVirtualNetworkController(d, meta)
	if err != nil {
		return err
	}

	if d.Get("hold_size").(int) > 0 {
		// add address range and reservations
		ip := net.ParseIP(d.Get("ip_hold").(string))
		ip = ip.To4()

		for i := 0; i < d.Get("reservation_size").(int); i++ {
			var address_reservation_string = `LEASES=[IP=%s]`
			r_err := vnc.Release(fmt.Sprintf(address_reservation_string, ip))

			if r_err != nil {
				return r_err
			}

			ip[3]++
		}
		log.Printf("[INFO] Successfully released reservered IP addresses.")
	}

	err = vnc.Delete()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Successfully deleted Vnet\n")
	return nil
}
