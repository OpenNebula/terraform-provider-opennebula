package opennebula

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/fatih/structs"
	"github.com/hashicorp/terraform/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	errs "github.com/OpenNebula/one/src/oca/go/src/goca/errors"
	vn "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualnetwork"
	vnk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualnetwork/keys"
)

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
				Computed:    true,
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
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "ID of the group that will own the vnet",
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
				Computed:      true,
				Description:   "List of cluster IDs hosting the virtual Network, if not set it uses the default cluster",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"vlan_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "VLAN ID. Only if 'Type' is : 802.1Q, vxlan or ovswich and if 'automatic_vlan_id' is not set",
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "automatic_vlan_id"},
			},
			"automatic_vlan_id": {
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true,
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
				Computed:      true,
				Description:   "Gateway IP if necessary",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"network_mask": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Network Mask",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"dns": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "DNS IP if necessary",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"ar": {
				Type:          schema.TypeList,
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
							Computed:    true,
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
							Computed:    true,
							Description: "Start IPv6 of the range to be allocated (Required if IP6_STATIC or IP4_6_STATIC)",
						},
						"mac": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Start MAC of the range to be allocated",
						},
						"global_prefix": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Global prefix for IP6 or IP4_6",
						},
						"ula_prefix": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "ULA prefix for IP6 or IP4_6",
						},
						"prefix_length": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Prefix lenght Only needed for IP6_STATIC or IP4_6_STATIC",
						},
					},
				},
			},
			"hold_ips": {
				Type:          schema.TypeList,
				Optional:      true,
				Computed:      true,
				Description:   "List of IPs to be held the VNET",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"hold_size": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				Description:   "Carve a network reservation of this size from the reservation starting from `ip_hold`",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
				Deprecated:    "use 'hold_ips' instead",
			},
			"ip_hold": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Start IP of the range to be held",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
				Deprecated:    "use 'hold_ips' instead",
			},
			"reservation_vnet": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				Description:   "Create a reservation from this VNET ID",
				ConflictsWith: []string{"bridge", "physical_device", "ar", "hold_ips", "hold_size", "ip_hold", "type", "vlan_id", "automatic_vlan_id", "mtu", "clusters", "dns", "gateway", "network_mask"},
			},
			"reservation_size": {
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				Description:   "Reserve this many IPs from reservation_vnet",
				ConflictsWith: []string{"bridge", "physical_device", "ar", "hold_ips", "hold_size", "ip_hold", "type", "vlan_id", "automatic_vlan_id", "mtu", "clusters", "dns", "gateway", "network_mask"},
			},
			"security_groups": {
				Type:        schema.TypeList,
				Optional:    true,
				Computed:    true,
				Description: "List of Security Group IDs to be applied to the VNET",
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"group": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Name of the Group that onws the Virtual Network, If empty, it uses caller group",
			},
			"tags": tagsSchema(),
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
		vnDef, err := generateVn(d)
		if err != nil {
			return err
		}

		// Get Clusters list
		clusters := getVnetClustersValue(d)

		// Create VNet
		vnetID, err := controller.VirtualNetworks().Create(vnDef, clusters[0])
		if err != nil {
			return err
		}
		vnc = controller.VirtualNetwork(vnetID)

		d.SetId(fmt.Sprintf("%v", vnetID))

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
		ars := d.Get("ar").([]interface{})

		for i, arinterface := range ars {
			armap := arinterface.(map[string]interface{})
			arstr := generateAR(armap, i)
			err := vnc.AddAR(arstr)
			if err != nil {
				return fmt.Errorf("Error: %s\nAR: %s", err, arstr)
			}
		}

		// Set Clusters (first in list is already set)
		if len(clusters) > 1 {
			err := setVnetClusters(clusters[1:], meta, vnetID)
			if err != nil {
				return err
			}
		}

		// Deprecated
		if d.Get("hold_size").(int) > 0 {
			// add address range and reservations
			ip := net.ParseIP(d.Get("ip_hold").(string))
			ip = ip.To4()

			for i := 0; i < d.Get("hold_size").(int); i++ {
				var address_reservation_string = `LEASES = [ IP = %s]`
				r_err := vnc.Hold(fmt.Sprintf(address_reservation_string, ip))
				if r_err != nil {
					return r_err
				}

				ip[3]++
			}
		}

		if hold_ips_list, ok := d.GetOk("hold_ips"); ok {
			for _, ip := range hold_ips_list.([]interface{}) {
				var address_reservation_string = `LEASES = [ IP = %s]`
				r_err := vnc.Hold(fmt.Sprintf(address_reservation_string, ip.(string)))
				if r_err != nil {
					return r_err
				}
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

	ar := vn.NewAddressRange()

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

	ar.Add(vnk.ARID, arid)
	ar.Add(vnk.Size, fmt.Sprint(arsize))
	ar.Add(vnk.Type, artype)

	if armac != "" {
		ar.Add(vnk.Mac, armac)
	}

	switch artype {
	case "IP4":
		ar.Add(vnk.IP, arip4)

	case "IP6":

		if argprefix != "" {
			ar.Add(vnk.GlobalPrefix, argprefix)
		}

		if arulaprefix != "" {
			ar.Add(vnk.UlaPrefix, arulaprefix)
		}

	case "IP6_STATIC":

		ar.Add("IP6", arip6)
		ar.Add(vnk.PrefixLength, arprefixlength)

	case "IP4_6":

		if argprefix != "" {
			ar.Add(vnk.GlobalPrefix, argprefix)
		}

		if arulaprefix != "" {
			ar.Add(vnk.UlaPrefix, arulaprefix)
		}

		ar.Add(vnk.IP, arip4)

	case "IP4_6_STATIC":

		ar.Add(vnk.IP, arip4)
		ar.Add("IP6", arip6)
		ar.Add(vnk.PrefixLength, arprefixlength)
	}

	return ar.String()
}

func generateVnTemplate(d *schema.ResourceData) (string, error) {

	tpl := vn.NewTemplate()

	mtu := d.Get("mtu").(int)
	guestmtu := d.Get("guest_mtu").(int)

	if guestmtu > mtu {
		return "", fmt.Errorf("Invalid: Guest MTU (%v) is greater than MTU (%v)", guestmtu, mtu)
	}

	tpl.AddPair("MTU", mtu)
	tpl.AddPair(string(vnk.GuestMTU), guestmtu)

	if dns, ok := d.GetOk("dns"); ok {
		tpl.Add(vnk.DNS, dns.(string))
	}
	if gw, ok := d.GetOk("gateway"); ok {
		tpl.Add(vnk.Gateway, gw.(string))
	}
	if netMask, ok := d.GetOk("network_mask"); ok {
		tpl.Add(vnk.NetworkMask, netMask.(string))
	}
	if desc, ok := d.GetOk("description"); ok {
		tpl.Add("DESCRIPTION", desc.(string))
	}

	tagsInterface := d.Get("tags").(map[string]interface{})
	for k, v := range tagsInterface {
		tpl.AddPair(strings.ToUpper(k), v)
	}

	tplStr := tpl.String()
	log.Printf("[INFO] VNET template: %s", tplStr)

	return tplStr, nil
}

func generateVn(d *schema.ResourceData) (string, error) {
	vnname := d.Get("name").(string)
	vnmad := d.Get("type").(string)

	if vnmad == "" {
		vnmad = "bridge"
	}

	tpl := vn.NewTemplate()

	tpl.Add(vnk.Name, vnname)
	tpl.Add(vnk.VNMad, vnmad)

	if validVlanType(vnmad) >= 0 {
		if d.Get("automatic_vlan_id") == true {
			tpl.Add("AUTOMATIC_VLAN_ID", "YES")
		} else if vlanid, ok := d.GetOk("vlan_id"); ok {
			tpl.Add(vnk.VlanID, vlanid.(string))
		} else {
			return "", fmt.Errorf("You must specify a 'vlan_id' or set the flag 'automatic_vlan_id'")
		}
	}
	if vnbridge, ok := d.GetOk("bridge"); ok {
		tpl.Add(vnk.Bridge, vnbridge.(string))
	}
	if vnphydev, ok := d.GetOk("physical_device"); ok {
		tpl.Add(vnk.PhyDev, vnphydev.(string))
	}

	tplStr := tpl.String()
	log.Printf("[INFO] VNET definition: %s", tplStr)

	return tplStr, nil
}

func getVnetClustersValue(d *schema.ResourceData) []int {
	var result = make([]int, 0)

	if clusters, ok := d.GetOk("clusters"); ok {
		clusterList := clusters.([]interface{})
		for i := 0; i < len(clusterList); i++ {
			result = append(result, clusterList[i].(int))
		}
	} else {
		result = append(result, -1)
	}
	return result
}

func setVnetClusters(clusters []int, meta interface{}, id int) error {
	controller := meta.(*goca.Controller)
	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	clusterPool, err := controller.Clusters().Info(false)
	if err != nil {
		return err
	}
	log.Printf("Number of clusters: %d", len(clusters))
	for i := 0; i < len(clusters); i++ {
		clusterid := clusters[i]
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
	d.Set("permissions", permissionsUnixString(*vn.Permissions))

	err = flattenVnetTemplate(d, &vn.Template)
	if err != nil {
		return err
	}

	if err := d.Set("ar", generateARMapFromStructs(vn.ARs)); err != nil {
		log.Printf("[WARN] Error setting ar for Virtual Network %x, error: %s", vn.ID, err)
	}
	return nil
}

func flattenVnetTemplate(d *schema.ResourceData, vnTpl *vn.Template) error {

	tags := make(map[string]interface{})
	for i, _ := range vnTpl.Elements {
		pair, ok := vnTpl.Elements[i].(*dyn.Pair)
		if !ok {
			continue
		}

		switch pair.Key() {
		case "SECURITY_GROUPS":
			secgrouplist, err := vnTpl.GetStr("SECURITY_GROUPS")
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
		case "MTU":
			mtustr, _ := vnTpl.Get("MTU")
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
		case "DESCRIPTION":
			desc, err := vnTpl.Get("DESCRIPTION")
			if desc != "" {
				err = d.Set("description", desc)
				if err != nil {
					return err
				}
			}
		default:
			// Get only tags from userTemplate
			if tagsInterface, ok := d.GetOk("tags"); ok {
				var err error
				for k, _ := range tagsInterface.(map[string]interface{}) {
					tags[k], err = vnTpl.GetStr(strings.ToUpper(k))
					if err != nil {
						return err
					}
				}
			}
		}
	}

	if len(tags) > 0 {
		err := d.Set("tags", tags)
		if err != nil {
			return err
		}
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

	tpl := vn.NewTemplate()
	changes := false

	if d.HasChange("description") {
		tpl.Add("DESCRIPTION", d.Get("description").(string))
		changes = true
	}

	if d.HasChange("gateway") {
		tpl.Add(vnk.Gateway, d.Get("gateway").(string))
		changes = true
	}

	if d.HasChange("dns") {
		tpl.Add(vnk.DNS, d.Get("dns").(string))
		changes = true
	}

	if d.HasChange("network_mask") {
		tpl.Add(vnk.NetworkMask, d.Get("network_mask").(string))
		changes = true
	}

	if d.HasChange("security_groups") {
		securitygroups := d.Get("security_groups")
		secgrouplist := ArrayToString(securitygroups.([]interface{}), ",")
		tpl.Add(vnk.SecGroups, secgrouplist)
		changes = true
	}

	if d.HasChange("tags") {
		tagsInterface := d.Get("tags").(map[string]interface{})
		for k, v := range tagsInterface {
			tpl.Del(strings.ToUpper(k))
			tpl.AddPair(strings.ToUpper(k), v)
		}
		changes = true
	}

	if changes {
		err := vnc.Update(tpl.String(), 1)
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

	if d.HasChange("hold_ips") {
		// Release all old Held IPs
		o_hold_ips_list, _ := d.GetChange("hold_ips")
		for _, ip := range o_hold_ips_list.([]interface{}) {
			var address_reservation_string = `LEASES = [ IP = %s]`
			r_err := vnc.Release(fmt.Sprintf(address_reservation_string, ip.(string)))
			if r_err != nil {
				return r_err
			}
		}
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vn, err = vnc.Info(false)
	if err != nil {
		return err
	}

	if d.HasChange("ar") {
		_, narsset := d.GetChange("ar")

		nars := narsset.([]interface{})
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

	if d.HasChange("hold_ips") {
		_, n_hold_ips_list := d.GetChange("hold_ips")
		// Hold only requested IPs
		for _, ip := range n_hold_ips_list.([]interface{}) {
			var address_reservation_string = `LEASES = [ IP = %s]`
			r_err := vnc.Hold(fmt.Sprintf(address_reservation_string, ip.(string)))
			if r_err != nil {
				return r_err
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

	// Deprecated
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
	}

	if hold_ips_list, ok := d.GetOk("hold_ips"); ok {
		for _, ip := range hold_ips_list.([]interface{}) {
			var address_reservation_string = `LEASES = [ IP = %s]`
			r_err := vnc.Release(fmt.Sprintf(address_reservation_string, ip.(string)))
			if r_err != nil {
				return r_err
			}
		}
	}
	log.Printf("[INFO] Successfully released reservered IP addresses.")

	err = vnc.Delete()
	if err != nil {
		return err
	}

	log.Printf("[INFO] Successfully deleted Vnet\n")
	return nil
}
