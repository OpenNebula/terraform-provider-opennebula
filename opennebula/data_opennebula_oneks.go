package opennebula

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func dataOpennebulaOneKS() *schema.Resource {
	return &schema.Resource{
		ReadContext: datasourceOpennebulaOneKSRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeInt,
				Optional:    true,
				Computed:    true,
				Description: "OneKS cluster ID; zero is valid",
			},
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Computed:    true,
				Description: "Name of the OneKS cluster",
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the OneKS cluster",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "State of the OneKS cluster",
			},
			"kubernetes_version": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Kubernetes version of the OneKS cluster",
			},
			"deployment": oneKSDeploymentSchema(true),
		},
	}
}

func oneKSFilter(ctx context.Context, d *schema.ResourceData, meta interface{}) (*oneKSResponse, error) {
	client, err := getOneKSClient(meta)
	if err != nil {
		return nil, err
	}

	nameConfigured := oneKSSelectorConfigured(d, "name")
	if oneKSSelectorConfigured(d, "id") {
		id := d.Get("id")
		result, err := client.cluster(ctx, strconv.Itoa(id.(int)))
		if err != nil {
			return nil, err
		}
		if nameConfigured && d.Get("name").(string) != result.Document.Template.Body.Name {
			return nil, fmt.Errorf("OneKS cluster does not match the supplied name")
		}
		return result, nil
	}
	if !nameConfigured {
		return nil, fmt.Errorf("a OneKS cluster id or name must be supplied")
	}
	var clusters []oneKSResponse
	if err := client.request(ctx, http.MethodGet, "/clusters", nil, &clusters); err != nil {
		return nil, err
	}
	var match *oneKSResponse
	for i := range clusters {
		if err := clusters[i].validate(); err != nil {
			return nil, err
		}
		if clusters[i].Document.Template.Body.Name != d.Get("name").(string) {
			continue
		}
		if match != nil {
			return nil, fmt.Errorf("several OneKS clusters match the constraints")
		}
		match = &clusters[i]
	}
	if match == nil {
		return nil, fmt.Errorf("no OneKS clusters match the constraints")
	}
	return match, nil
}

func oneKSSelectorConfigured(d *schema.ResourceData, key string) bool {
	config := d.GetRawConfig()
	if config.IsKnown() && !config.IsNull() {
		value := config.GetAttr(key)
		return value.IsKnown() && !value.IsNull()
	}
	_, ok := d.GetOkExists(key)
	return ok
}

func datasourceOpennebulaOneKSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	result, err := oneKSFilter(ctx, d, meta)
	if err != nil {
		return diag.FromErr(err)
	}
	if err := setOneKSState(d, result); err != nil {
		return diag.FromErr(err)
	}
	id, _ := strconv.Atoi(result.Document.ID.String())
	if err := d.Set("id", id); err != nil {
		return diag.FromErr(err)
	}
	d.SetId(result.Document.ID.String())
	return nil
}
