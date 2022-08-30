package opennebula

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/schemas/shared"
	vn "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualnetwork"
	vnk "github.com/OpenNebula/one/src/oca/go/src/goca/schemas/virtualnetwork/keys"
)

var defaultVNetTimeout = time.Duration(5) * time.Minute

type Template string

const (
	ReservationSize           Template = "SIZE"
	ReservationFirstIP        Template = "IP"
	ReservationAddressRangeID Template = "AR_ID"
)

func resourceOpennebulaVirtualNetwork() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaVirtualNetworkCreate,
		ReadContext:   resourceOpennebulaVirtualNetworkRead,
		Exists:        resourceOpennebulaVirtualNetworkExists,
		UpdateContext: resourceOpennebulaVirtualNetworkUpdate,
		DeleteContext: resourceOpennebulaVirtualNetworkDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(defaultVNetTimeout),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
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
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "reservation_ar_id", "reservation_first_ip"},
			},
			"physical_device": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Name of the physical device to which the vnet should be associated",
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "reservation_ar_id", "reservation_first_ip"},
			},
			"type": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "bridge",
				Description:   "Type of the Virtual Network: dummy, bridge, fw, ebtables, 802.1Q, vxlan, ovswitch. Default is 'bridge'",
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "reservation_ar_id", "reservation_first_ip"},
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
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "reservation_ar_id", "reservation_first_ip"},
				Elem: &schema.Schema{
					Type: schema.TypeInt,
				},
			},
			"vlan_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "VLAN ID. Only if 'Type' is : 802.1Q, vxlan or ovswich and if 'automatic_vlan_id' is not set",
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "reservation_ar_id", "reservation_first_ip", "automatic_vlan_id"},
			},
			"automatic_vlan_id": {
				Type:          schema.TypeBool,
				Optional:      true,
				Computed:      true,
				Description:   "If set, let OpenNebula to attribute VLAN ID",
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "reservation_ar_id", "reservation_first_ip", "vlan_id"},
			},
			"mtu": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "MTU of the vnet (defaut: 1500)",
				Default:       1500,
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "reservation_ar_id", "reservation_first_ip"},
			},
			"guest_mtu": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "MTU of the Guest interface. Must be lower or equal to 'mtu' (defaut: 1500)",
				Default:       1500,
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "reservation_ar_id", "reservation_first_ip"},
			},
			"gateway": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Gateway IP if necessary",
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "reservation_ar_id", "reservation_first_ip"},
			},
			"network_mask": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Network Mask",
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "reservation_ar_id", "reservation_first_ip"},
			},
			"network_address": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Network Address",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"search_domain": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "Search Domain",
				ConflictsWith: []string{"reservation_vnet", "reservation_size"},
			},
			"dns": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				Description:   "DNS IP if necessary",
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "reservation_ar_id", "reservation_first_ip"},
			},
			"ar": {
				Type:          schema.TypeSet,
				Optional:      true,
				MinItems:      1,
				Description:   "List of Address Ranges to be part of the Virtual Network",
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "reservation_ar_id", "reservation_first_ip"},
				Elem: &schema.Resource{
					Schema: ARFields(),
				},
				Deprecated: "use virtual network address range resource instead",
			},
			"hold_ips": {
				Type:          schema.TypeList,
				Optional:      true,
				Description:   "List of IPs to be held the VNET",
				ConflictsWith: []string{"reservation_vnet", "reservation_size", "reservation_ar_id", "reservation_first_ip"},
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Deprecated: "use 'hold ips' in the related virtual network address range resource instead",
			},
			"reservation_vnet": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "Create a reservation from this VNET ID",
				ConflictsWith: []string{"bridge", "physical_device", "ar", "hold_ips", "type", "vlan_id", "automatic_vlan_id", "mtu", "clusters", "dns", "gateway", "network_mask", "network_address", "search_domain"},
				Default:       -1,
			},
			"reservation_size": {
				Type:          schema.TypeInt,
				Optional:      true,
				Description:   "Reserve this many IPs from reservation_vnet",
				ConflictsWith: []string{"bridge", "physical_device", "ar", "hold_ips", "type", "vlan_id", "automatic_vlan_id", "mtu", "clusters", "dns", "gateway", "network_mask", "network_address", "search_domain"},
			},
			"reservation_first_ip": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "First IP of the reservation",
				ConflictsWith: []string{"bridge", "physical_device", "ar", "hold_ips", "type", "vlan_id", "automatic_vlan_id", "mtu", "clusters", "dns", "gateway", "network_mask"},
			},
			"reservation_ar_id": {
				Type:          schema.TypeInt,
				Optional:      true,
				Default:       -1,
				Description:   "Address Range ID to be used for the reservation",
				ConflictsWith: []string{"bridge", "physical_device", "ar", "hold_ips", "type", "vlan_id", "automatic_vlan_id", "mtu", "clusters", "dns", "gateway", "network_mask"},
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
			"lock": lockSchema(),
		},
	}
}

func ARFields() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"id": {
			Type:     schema.TypeString,
			Computed: true,
		},
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
		"computed_ip6": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Start IPv6 of the range to be allocated (Required if IP6_STATIC or IP4_6_STATIC)",
		},
		"computed_mac": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Start MAC of the range to be allocated",
		},
		"computed_global_prefix": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Global prefix for IP6 or IP4_6",
		},
		"computed_ula_prefix": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "ULA prefix for IP6 or IP4_6",
		},
	}
}

func getVirtualNetworkController(d *schema.ResourceData, meta interface{}, args ...int) (*goca.VirtualNetworkController, error) {
	config := meta.(*Configuration)
	controller := config.Controller
	var vnc *goca.VirtualNetworkController

	// Try to find the VNet by ID, if specified
	if d.Id() != "" {
		id, err := strconv.ParseUint(d.Id(), 10, 0)
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
	config := meta.(*Configuration)
	controller := config.Controller
	var gid int

	vnc, err := getVirtualNetworkController(d, meta)
	if err != nil {
		return err
	}

	if d.Get("group") != "" {
		group := d.Get("group").(string)
		gid, err = controller.Groups().ByName(group)
		if err != nil {
			return fmt.Errorf("Can't find a group with name `%s`: %s", group, err)
		}
	} else {
		gid = d.Get("gid").(int)
	}

	err = vnc.Chown(-1, gid)
	if err != nil {
		return fmt.Errorf("Can't find a group with ID `%d`: %s", gid, err)
	}

	return nil
}

func validVlanType(intype string) int {
	vlanType := []string{"802.1Q", "vxlan", "ovswitch"}
	return inArray(intype, vlanType)
}

func resourceOpennebulaVirtualNetworkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	config := meta.(*Configuration)
	controller := config.Controller
	var vnc *goca.VirtualNetworkController
	var diags diag.Diagnostics

	reservationVNet := d.Get("reservation_vnet").(int)

	// VNET reservation
	if reservationVNet > -1 {
		reservationTemplate := dyn.NewTemplate()

		if reservationName, ok := d.GetOk("name"); ok {
			reservationTemplate.AddPair("NAME", reservationName.(string))
		}
		if reservationFirstIP, ok := d.GetOk("reservation_first_ip"); ok {
			reservationTemplate.AddPair("IP", reservationFirstIP.(string))
		}
		if reservationARID, ok := d.GetOk("reservation_ar_id"); ok && reservationARID != -1 {
			reservationTemplate.AddPair("AR_ID", reservationARID.(int))
		}
		reservationSize, ok := d.GetOk("reservation_size")
		if ok {
			reservationTemplate.AddPair("SIZE", reservationSize.(int))
		}

		if reservationSize.(int) <= 0 {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Wrong size value",
				Detail:   "Reservation size must be strictly greater than 0",
			})
			return diags
		}

		// Get VNet Controller to reserve from
		vnc = controller.VirtualNetwork(reservationVNet)

		// Call .Info to check if the Network exists
		_, err := vnc.Info(false)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve informations",
				Detail:   fmt.Sprintf("Virtual network (ID: %d) reservation: %s", reservationVNet, err),
			})
			return diags
		}

		rID, err := vnc.Reserve(reservationTemplate.String())
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to reserve network addresses",
				Detail:   fmt.Sprintf("Virtual network (ID: %d) reservation: %s", reservationVNet, err),
			})
			return diags
		}

		vnc = controller.VirtualNetwork(rID)

		// TODO: fix it after 5.10 release
		// Force the "decrypt" bool to false to keep ONE 5.8 behavior
		vnet, err := vnc.Info(false)

		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to retrieve informations",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		d.SetId(fmt.Sprintf("%v", vnet.ID))

		log.Printf("[DEBUG] New VNET reservation ID: %d", vnet.ID)

	} else { //New VNET
		vnDef, err := generateVn(d)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to generate description",
				Detail:   err.Error(),
			})
			return diags
		}

		// Get Clusters list
		clusters := getVnetClustersValue(d)

		// Create VNet
		vnetID, err := controller.VirtualNetworks().Create(vnDef, clusters[0])
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to create the virtual network",
				Detail:   err.Error(),
			})
			return diags
		}
		vnc = controller.VirtualNetwork(vnetID)

		// virtual network states were introduce with OpenNebula 6.4 release
		requiredVersion, _ := version.NewVersion("6.4.0")

		if config.OneVersion.GreaterThanOrEqual(requiredVersion) {
			timeout := d.Timeout(schema.TimeoutCreate)
			transient := []string{vn.Init.String(), vn.LockCreate.String()}
			_, err = waitForVNetworkState(ctx, vnc, timeout, transient, []string{vn.Ready.String()})
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to wait virtual network to be in READY state",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}

		d.SetId(fmt.Sprintf("%v", vnetID))

		// Call API once
		update, err := generateVnTemplate(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to generate template description",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		err = vnc.Update(update, 1)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update content",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		// Address Ranges
		ars := d.Get("ar").(*schema.Set).List()

		for _, arinterface := range ars {
			armap := arinterface.(map[string]interface{})
			arstr := vnetGenerateAR(armap).String()
			err := vnc.AddAR(arstr)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add an address range",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}

		// Set Clusters (first in list is already set)
		if len(clusters) > 1 {
			err := setVnetClusters(clusters[1:], meta, vnetID)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to set cluster",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}

		if hold_ips_list, ok := d.GetOk("hold_ips"); ok {
			for _, ip := range hold_ips_list.([]interface{}) {
				var address_reservation_string = `LEASES = [ IP = %s]`
				err := vnc.Hold(fmt.Sprintf(address_reservation_string, ip.(string)))
				if err != nil {
					diags = append(diags, diag.Diagnostic{
						Severity: diag.Error,
						Summary:  "Failed to hold a lease",
						Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
					})
					return diags
				}
			}
		}

	}

	// Set Security Groups
	if securitygroups, ok := d.GetOk("security_groups"); ok {
		secgrouplist := ArrayToString(securitygroups.([]interface{}), ",")

		err := vnc.Update(fmt.Sprintf("SECURITY_GROUPS=\"%s\"", secgrouplist), 1)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to add security groups",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}
	// update permisions
	if perms, ok := d.GetOk("permissions"); ok {
		err := vnc.Chmod(permissionUnix(perms.(string)))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change permissions",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.Get("group") != "" || d.Get("gid") != "" {
		err := changeVNetGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if lock, ok := d.GetOk("lock"); ok && lock.(string) != "UNLOCK" {

		var level shared.LockLevel
		err := StringToLockLevel(lock.(string), &level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to convert lock level",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		err = vnc.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to lock",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return resourceOpennebulaVirtualNetworkRead(ctx, d, meta)
}

func vnetGenerateAR(armap map[string]interface{}) *vn.AddressRange {

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

	return ar
}

func generateVnTemplate(d *schema.ResourceData, meta interface{}) (string, error) {

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
	if netAddr, ok := d.GetOk("network_address"); ok {
		tpl.Add(vnk.NetworkAddress, netAddr.(string))
	}
	if searchDom, ok := d.GetOk("search_domain"); ok {
		tpl.Add(vnk.SearchDomain, searchDom.(string))
	}
	if desc, ok := d.GetOk("description"); ok {
		tpl.Add("DESCRIPTION", desc.(string))
	}

	tagsInterface := d.Get("tags").(map[string]interface{})
	for k, v := range tagsInterface {
		tpl.AddPair(strings.ToUpper(k), v)
	}

	// add default tags if they aren't overriden
	config := meta.(*Configuration)

	if len(config.defaultTags) > 0 {
		for k, v := range config.defaultTags {
			key := strings.ToUpper(k)
			p, _ := tpl.GetPair(key)
			if p != nil {
				continue
			}
			tpl.AddPair(key, v)
		}
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
	config := meta.(*Configuration)
	controller := config.Controller
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

func matchARs(ARConfig map[string]interface{}, AR vn.AR) bool {

	return AR.Type == ARConfig["ar_type"].(string) &&
		AR.Size == ARConfig["size"].(int) &&
		emptyOrEqual(ARConfig["ip4"], AR.IP) &&
		emptyOrEqual(ARConfig["ip6"], AR.IP6) &&
		emptyOrEqual(ARConfig["mac"], AR.MAC) &&
		emptyOrEqual(ARConfig["global_prefix"], AR.GlobalPrefix) &&
		emptyOrEqual(ARConfig["ula_prefix"], AR.ULAPrefix)
}

func resourceOpennebulaVirtualNetworkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	vnc, err := getVirtualNetworkController(d, meta, -2, -1, -1)
	if err != nil {
		if NoExists(err) {
			log.Printf("[WARN] Removing virtual network %s from state because it no longer exists in", d.Get("name"))
			d.SetId("")
			return nil
		}

		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the virtual network controller",
			Detail:   err.Error(),
		})
		return diags
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vn, err := vnc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
		})
		return diags
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
	d.Set("permissions", permissionsUnixString(*vn.Permissions))

	err = flattenVnetTemplate(d, &vn.Template)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to flatten template",
			Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	ARIf := d.Get("ar")
	if ARIf != nil {
		err = flattenVnetARs(d, vn)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to flatten address ranges",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if len(vn.ParentNetworkID) > 0 {
		parentNetworkID, err := strconv.ParseInt(vn.ParentNetworkID, 10, 0)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to parse parent network ID",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		d.Set("reservation_vnet", parentNetworkID)
		if len(vn.ARs) > 0 {
			arID, err := strconv.ParseInt(vn.ARs[0].ID, 10, 0)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to parse address range ID",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
			d.Set("reservation_ar_id", arID)
			d.Set("reservation_size", vn.ARs[0].Size)
			d.Set("reservation_first_ip", vn.ARs[0].IP)
		}
	}

	if vn.Lock != nil {
		d.Set("lock", LockLevelToString(vn.Lock.Locked))
	}

	return nil
}

func flattenVnetARs(d *schema.ResourceData, vn *vn.VirtualNetwork) error {

	ARSet := make([]map[string]interface{}, 0, len(vn.ARs))
	ARConfigs := d.Get("ar").(*schema.Set).List()
	log.Printf("[INFO] ARs: %+v", vn.ARs)
	log.Printf("[INFO] ARConfigs: %+v", ARConfigs)
	for _, AR := range vn.ARs {

		match := false

		// retrieve the associated AR config
		for _, ARConfigIf := range ARConfigs {

			ARConfig := ARConfigIf.(map[string]interface{})

			if !matchARs(ARConfig, AR) {
				continue
			}

			match = true
			ARMap := flattenAR(ARConfig, AR)
			ARSet = append(ARSet, ARMap)

			break
		}

		if !match {
			log.Printf("[WARN] Configuration for AR ID %s not found.", AR.ID)
		}

	}

	if err := d.Set("ar", ARSet); err != nil {
		log.Printf("[WARN] Error setting ar for Virtual Network %x, error: %s", vn.ID, err)
	}

	return nil
}

func flattenVnetTemplate(d *schema.ResourceData, vnTpl *vn.Template) error {

	tags := make(map[string]interface{})
	tagsInterface, tagsOk := d.GetOk("tags")
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
				mtu, err := strconv.ParseInt(mtustr, 10, 0)
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
			if tagsOk {
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

	if _, ok := d.GetOk("tags"); ok {
		err := d.Set("tags", tags)
		if err != nil {
			return err
		}
	}

	return nil
}

func flattenAR(config map[string]interface{}, AR vn.AR) map[string]interface{} {

	ARMap := map[string]interface{}{
		"id":                     AR.ID,
		"ar_type":                AR.Type,
		"ip4":                    AR.IP,
		"size":                   AR.Size,
		"computed_ip6":           AR.IP6,
		"computed_mac":           AR.MAC,
		"computed_global_prefix": AR.GlobalPrefix,
		"computed_ula_prefix":    AR.ULAPrefix,
	}

	// if attribute set by the user, set read value
	if len(config["ip6"].(string)) > 0 {
		ARMap["ip6"] = AR.IP6
	}
	if len(config["mac"].(string)) > 0 {
		ARMap["mac"] = AR.MAC
	}
	if len(config["global_prefix"].(string)) > 0 {
		ARMap["global_prefix"] = AR.GlobalPrefix
	}
	if len(config["ula_prefix"].(string)) > 0 {
		ARMap["ula_prefix"] = AR.ULAPrefix
	}

	return ARMap
}

func resourceOpennebulaVirtualNetworkExists(d *schema.ResourceData, meta interface{}) (bool, error) {

	config := meta.(*Configuration)
	controller := config.Controller

	imageID, err := strconv.ParseInt(d.Id(), 10, 0)
	if err != nil {
		return false, err
	}

	_, err = controller.VirtualNetwork(int(imageID)).Info(false)
	if NoExists(err) {
		return false, err
	}

	return true, err
}

func resourceOpennebulaVirtualNetworkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	//Get Virtual Network Controller
	vnc, err := getVirtualNetworkController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the virtual network controller",
			Detail:   err.Error(),
		})
		return diags
	}

	lock, lockOk := d.GetOk("lock")
	if d.HasChange("lock") && lockOk && lock.(string) == "UNLOCK" {

		err = vnc.Unlock()
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to unlock",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
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

	if d.HasChange("network_address") {
		tpl.Add(vnk.NetworkAddress, d.Get("network_address").(string))
		changes = true
	}

	if d.HasChange("search_domain") {
		tpl.Add(vnk.SearchDomain, d.Get("search_domain").(string))
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
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to update content",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	if d.HasChange("name") {
		err := vnc.Rename(d.Get("name").(string))
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to rename",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated name for Vnet\n")
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vn, err := vnc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	if d.HasChange("permissions") {
		if perms, ok := d.GetOk("permissions"); ok {
			err = vnc.Chmod(permissionUnix(perms.(string)))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to change permissions",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
		log.Printf("[INFO] Successfully updated Vnet\n")
	}

	if d.HasChange("group") {
		err = changeVNetGroup(d, meta)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to change group",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
		log.Printf("[INFO] Successfully updated group for Vnet %s\n", vn.Name)
	}

	if d.HasChange("hold_ips") {
		// Release all old Held IPs
		o_hold_ips_list, _ := d.GetChange("hold_ips")
		for _, ip := range o_hold_ips_list.([]interface{}) {
			var address_reservation_string = `LEASES = [ IP = %s]`
			err := vnc.Release(fmt.Sprintf(address_reservation_string, ip.(string)))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to release a lease on hold",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
	}

	// TODO: fix it after 5.10 release
	// Force the "decrypt" bool to false to keep ONE 5.8 behavior
	vn, err = vnc.Info(false)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to retrieve informations",
			Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	if d.HasChange("ar") {
		log.Println("[DEBUG] AR changed")
		old, new := d.GetChange("ar")
		existingARsCfg := old.(*schema.Set).List()
		newARsCfg := new.(*schema.Set).List()

		ARToRem, ARToAdd := diffListConfig(newARsCfg, existingARsCfg,
			&schema.Resource{
				Schema: ARFields(),
			},
			"ar_type",
			"ip4",
			"size",
			"ip6",
			"global_prefix",
			"ula_prefix",
			"prefix_length",
		)

		// remove ARs
		for _, ARIf := range ARToRem {
			ARConfig := ARIf.(map[string]interface{})

			ARID, err := strconv.ParseInt(ARConfig["id"].(string), 10, 0)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to parse address range ID",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}

			err = vnc.RmAR(int(ARID))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to remove address range",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}

		// Add new ARs
		for _, ARIf := range ARToAdd {
			ARConfig := ARIf.(map[string]interface{})

			ARTemplateStr := vnetGenerateAR(ARConfig).String()

			err = vnc.AddAR(ARTemplateStr)
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to add address range",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}

	}

	if d.HasChange("hold_ips") {
		_, n_hold_ips_list := d.GetChange("hold_ips")
		// Hold only requested IPs
		for _, ip := range n_hold_ips_list.([]interface{}) {
			var address_reservation_string = `LEASES = [ IP = %s]`
			err := vnc.Hold(fmt.Sprintf(address_reservation_string, ip.(string)))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to hold a lease",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
	}

	if d.HasChange("lock") && lockOk && lock.(string) != "UNLOCK" {

		var level shared.LockLevel

		err = StringToLockLevel(lock.(string), &level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to convert lock level",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}

		err = vnc.Lock(level)
		if err != nil {
			diags = append(diags, diag.Diagnostic{
				Severity: diag.Error,
				Summary:  "Failed to lock",
				Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
			})
			return diags
		}
	}

	return resourceOpennebulaVirtualNetworkRead(ctx, d, meta)
}

func resourceOpennebulaVirtualNetworkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {

	var diags diag.Diagnostics

	vnc, err := getVirtualNetworkController(d, meta)
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to get the virtual network controller",
			Detail:   err.Error(),
		})
		return diags
	}

	if hold_ips_list, ok := d.GetOk("hold_ips"); ok {
		for _, ip := range hold_ips_list.([]interface{}) {
			var address_reservation_string = `LEASES = [ IP = %s]`
			err := vnc.Release(fmt.Sprintf(address_reservation_string, ip.(string)))
			if err != nil {
				diags = append(diags, diag.Diagnostic{
					Severity: diag.Error,
					Summary:  "Failed to release a lease on hold",
					Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
				})
				return diags
			}
		}
	}
	log.Printf("[INFO] Successfully released reservered IP addresses.")

	err = vnc.Delete()
	if err != nil {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "Failed to delete",
			Detail:   fmt.Sprintf("virtual network (ID: %s): %s", d.Id(), err),
		})
		return diags
	}

	log.Printf("[INFO] Successfully deleted Vnet\n")
	return nil
}

func waitForVNetworkState(ctx context.Context, vnc *goca.VirtualNetworkController, timeout time.Duration, transient, final []string) (interface{}, error) {

	stateConf := &resource.StateChangeConf{
		Pending: transient,
		Target:  final,
		Refresh: func() (interface{}, string, error) {

			log.Println("Refreshing virtual network state...")

			var vNetInfos *vn.VirtualNetwork
			vNetInfos, err := vnc.Info(false)
			if err != nil {
				if NoExists(err) {
					return vNetInfos, "notfound", nil
				}
				return vNetInfos, "", err
			}
			state, err := vNetInfos.State()
			if err != nil {
				return vNetInfos, "", err
			}

			log.Printf("virtual network (ID:%d, name:%s) is currently in state %v", vNetInfos.ID, vNetInfos.Name, state.String())

			if state == vn.Error {
				return vNetInfos, state.String(), fmt.Errorf("virtual network (ID:%d) entered error state.", vNetInfos.ID)
			}

			return vNetInfos, state.String(), nil
		},
		Timeout:    timeout,
		Delay:      5 * time.Second,
		MinTimeout: 2 * time.Second,
	}

	return stateConf.WaitForStateContext(ctx)

}
