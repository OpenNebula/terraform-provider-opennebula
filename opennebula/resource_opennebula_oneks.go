package opennebula

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	ctyjson "github.com/zclconf/go-cty/cty/json"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceOpennebulaOneKS() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOpennebulaOneKSCreate,
		ReadContext:   resourceOpennebulaOneKSRead,
		UpdateContext: resourceOpennebulaOneKSUpdate,
		DeleteContext: resourceOpennebulaOneKSDelete,
		Importer:      &schema.ResourceImporter{StateContext: resourceOpennebulaOneKSImport},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(time.Hour),
			Delete: schema.DefaultTimeout(time.Hour),
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the OneKS cluster",
				ValidateFunc: validation.All(
					validation.StringLenBetween(1, 63),
					validation.StringMatch(regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`), "must be a valid name"),
				),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Computed:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Description:  "Cluster description. When omitted, preserves the existing description.",
			},
			"kubernetes_version": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				ValidateFunc: validation.StringIsNotWhiteSpace,
				Description:  "Kubernetes version supported by the selected family",
			},
			"deployment": oneKSDeploymentSchema(false),
			"spec": {
				Type:        schema.TypeList,
				Required:    true,
				MinItems:    1,
				MaxItems:    1,
				ForceNew:    true,
				Sensitive:   true,
				Description: "Control-plane creation specification",
				Elem: &schema.Resource{Schema: map[string]*schema.Schema{
					"family": {
						Type:         schema.TypeString,
						Optional:     true,
						ForceNew:     true,
						ValidateFunc: validation.StringIsNotWhiteSpace,
					},
					"flavour": {
						Type:         schema.TypeString,
						Required:     true,
						ForceNew:     true,
						ValidateFunc: validation.StringIsNotWhiteSpace,
					},
					"user_inputs_values": {
						Type:        schema.TypeString,
						Optional:    true,
						Sensitive:   true,
						ForceNew:    true,
						Description: "User input overrides as a JSON object; use jsonencode for dynamic values",
						ValidateFunc: validation.All(
							validation.StringIsJSON,
							validation.StringMatch(regexp.MustCompile(`^\s*\{`), "must be a JSON object"),
						),
						DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
							normalizedOld, oldErr := normalizeOneKSJSON(old)
							normalizedNew, newErr := normalizeOneKSJSON(new)
							return oldErr == nil && newErr == nil && normalizedOld == normalizedNew
						},
					},
				}},
			},
			"force_delete": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Force cleanup of a failed lifecycle operation during deletion",
			},
			"state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current OneKS lifecycle state",
			},
		},
	}
}

func resourceOpennebulaOneKSCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getOneKSClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	spec := &oneKSSpec{
		Family:  d.Get("spec.0.family").(string),
		Flavour: d.Get("spec.0.flavour").(string),
	}
	if inputs := d.Get("spec.0.user_inputs_values").(string); inputs != "" {
		if !oneKSJSONObject([]byte(inputs)) {
			return diag.Errorf("OneKS user_inputs_values must be a JSON object")
		}
		spec.UserInputs = json.RawMessage(inputs)
	}
	deployment := &oneKSDeployment{
		Cluster: &oneKSReference{ID: d.Get("deployment.0.cluster.0.id").(int)},
		Networks: &oneKSNetworks{
			Public:  &oneKSReference{ID: d.Get("deployment.0.networks.0.public.0.id").(int)},
			Private: &oneKSReference{ID: d.Get("deployment.0.networks.0.private.0.id").(int)},
		},
	}
	if _, ok := d.GetOk("deployment.0.datastores"); ok {
		deployment.Datastores = &oneKSDatastores{}
		if _, ok := d.GetOk("deployment.0.datastores.0.image"); ok {
			deployment.Datastores.Image = &oneKSReference{ID: d.Get("deployment.0.datastores.0.image.0.id").(int)}
		}
	}
	input := map[string]interface{}{
		"name":               d.Get("name"),
		"kubernetes_version": d.Get("kubernetes_version"),
		"deployment":         deployment,
		"spec":               spec,
	}
	if description, ok := d.GetOk("description"); ok {
		input["description"] = description
	}
	var result oneKSResponse
	err = client.request(ctx, http.MethodPost, "/clusters", input, &result)
	if result.Document != nil {
		if id, err := result.Document.ID.Int64(); err == nil && id >= 0 {
			d.SetId(result.Document.ID.String())
		}
	}
	if err != nil {
		return diag.FromErr(err)
	}
	if err := result.validate(); err != nil {
		return diag.FromErr(err)
	}
	if err := setOneKSState(d, &result); err != nil {
		return diag.FromErr(err)
	}
	cluster, err := waitOneKSCluster(ctx, client, d.Id(), false, d.Timeout(schema.TimeoutCreate))
	if cluster != nil {
		if stateErr := setOneKSState(d, cluster); stateErr != nil {
			return diag.FromErr(stateErr)
		}
	}
	if err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceOpennebulaOneKSRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getOneKSClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	result, err := client.cluster(ctx, d.Id())
	if oneKSNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	return diag.FromErr(setOneKSState(d, result))
}

func setOneKSState(d *schema.ResourceData, result *oneKSResponse) error {
	body := result.Document.Template.Body
	reference := func(ref *oneKSReference) []interface{} {
		return []interface{}{map[string]interface{}{"id": ref.ID}}
	}
	placement := body.Deployment
	deployment := map[string]interface{}{
		"cluster": reference(placement.Cluster),
		"networks": []interface{}{map[string]interface{}{
			"public":  reference(placement.Networks.Public),
			"private": reference(placement.Networks.Private),
		}},
	}
	if placement.Datastores != nil {
		datastores := map[string]interface{}{}
		if placement.Datastores.Image != nil {
			datastores["image"] = reference(placement.Datastores.Image)
		}
		deployment["datastores"] = []interface{}{datastores}
	}
	for key, value := range map[string]interface{}{
		"name":               body.Name,
		"description":        body.Description,
		"state":              body.State,
		"kubernetes_version": body.KubernetesVersion,
		"deployment":         []interface{}{deployment},
	} {
		if err := d.Set(key, value); err != nil {
			return err
		}
	}
	return nil
}

func resourceOpennebulaOneKSImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	client, err := getOneKSClient(meta)
	if err != nil {
		return nil, err
	}
	result, err := client.clusterWithQuery(ctx, d.Id(), "?expand=true&include_sensitive=true")
	if err != nil {
		return nil, err
	}
	controlPlane := result.Document.Template.Body.ControlPlane
	if controlPlane == nil || controlPlane.Flavour == "" {
		return nil, fmt.Errorf("cannot import OneKS cluster without an expanded control-plane flavour")
	}
	spec := map[string]interface{}{"family": controlPlane.Family, "flavour": controlPlane.Flavour}
	if controlPlane.UserInputs != nil {
		if !oneKSJSONObject(controlPlane.UserInputs) {
			return nil, fmt.Errorf("invalid expanded OneKS user input values")
		}
		inputs, err := normalizeOneKSJSON(string(controlPlane.UserInputs))
		if err != nil {
			return nil, err
		}
		spec["user_inputs_values"] = inputs
	}
	if err := d.Set("spec", []interface{}{spec}); err != nil {
		return nil, err
	}
	if err := d.Set("force_delete", false); err != nil {
		return nil, err
	}
	if err := setOneKSState(d, result); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func resourceOpennebulaOneKSUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getOneKSClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	input := map[string]interface{}{}
	for _, key := range []string{"name", "description"} {
		if d.HasChange(key) {
			input[key] = d.Get(key)
		}
	}
	if len(input) == 0 {
		return nil
	}
	var result oneKSResponse
	if err := client.request(ctx, http.MethodPatch, "/clusters/"+d.Id(), input, &result); err != nil {
		return diag.FromErr(err)
	}
	if err := result.validate(); err != nil {
		return diag.FromErr(err)
	}
	if result.Document.ID.String() != d.Id() {
		return diag.Errorf("OneKS API returned a different cluster ID")
	}
	return diag.FromErr(setOneKSState(d, &result))
}

func resourceOpennebulaOneKSDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := getOneKSClient(meta)
	if err != nil {
		return diag.FromErr(err)
	}
	current, err := client.cluster(ctx, d.Id())
	if oneKSNotFound(err) {
		d.SetId("")
		return nil
	}
	if err != nil {
		return diag.FromErr(err)
	}
	if current.Document.Template.Body.State != "DEPROVISIONING" {
		path := "/clusters/" + d.Id()
		if d.Get("force_delete").(bool) {
			path += "?force=true"
		}
		if err := client.request(ctx, http.MethodDelete, path, nil, nil); err != nil {
			if oneKSNotFound(err) {
				d.SetId("")
				return nil
			}
			return diag.FromErr(err)
		}
	}
	_, err = waitOneKSCluster(ctx, client, d.Id(), true, d.Timeout(schema.TimeoutDelete))
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId("")
	return nil
}

func waitOneKSCluster(ctx context.Context, client *oneKSClient, id string, deleting bool, timeout time.Duration) (*oneKSResponse, error) {
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	waiter := &resource.StateChangeConf{
		Pending:    []string{"waiting"},
		Target:     []string{"complete"},
		Timeout:    timeout,
		MinTimeout: 5 * time.Second,
		Refresh: func() (interface{}, string, error) {
			result, err := client.cluster(waitCtx, id)
			if deleting && oneKSNotFound(err) {
				return true, "complete", nil
			}
			if err != nil {
				return nil, "", err
			}
			state := result.Document.Template.Body.State
			if strings.HasSuffix(state, "_FAILURE") || state == "WARNING" {
				return result, state, fmt.Errorf("OneKS cluster %s entered %s; inspect its OneKS logs and recover the lifecycle operation before retrying", id, state)
			}
			if !deleting && state == "RUNNING" {
				return result, "complete", nil
			}
			if !deleting && (state == "DONE" || state == "DEPROVISIONING") {
				return result, state, fmt.Errorf("OneKS cluster %s entered unexpected state %s during creation", id, state)
			}
			return result, "waiting", nil
		},
	}
	raw, err := waiter.WaitForStateContext(waitCtx)
	result, _ := raw.(*oneKSResponse)
	if err != nil {
		return result, fmt.Errorf("waiting for OneKS cluster %s: %w", id, err)
	}
	return result, nil
}

func oneKSDeploymentSchema(computed bool) *schema.Schema {
	block := func(fields map[string]*schema.Schema, required bool) *schema.Schema {
		s := &schema.Schema{
			Type:     schema.TypeList,
			ForceNew: !computed,
			Elem:     &schema.Resource{Schema: fields},
		}
		if computed {
			s.Computed = true
		} else {
			s.Required = required
			s.Optional = !required
			s.MaxItems = 1
			if required {
				s.MinItems = 1
			}
		}
		return s
	}
	reference := func(required bool) *schema.Schema {
		id := &schema.Schema{
			Type:     schema.TypeInt,
			Required: !computed,
			Computed: computed,
			ForceNew: !computed,
		}
		if !computed {
			id.ValidateFunc = validation.IntAtLeast(0)
		}
		return block(map[string]*schema.Schema{"id": id}, required)
	}
	deployment := block(map[string]*schema.Schema{
		"cluster": reference(true),
		"networks": block(map[string]*schema.Schema{
			"public":  reference(true),
			"private": reference(true),
		}, true),
		"datastores": block(map[string]*schema.Schema{"image": reference(false)}, false),
	}, true)
	deployment.ForceNew = !computed
	deployment.Description = "OneKS deployment placement"
	return deployment
}

func normalizeOneKSJSON(value string) (string, error) {
	var decoded ctyjson.SimpleJSONValue
	if err := json.Unmarshal([]byte(value), &decoded); err != nil {
		return "", err
	}
	normalized, err := decoded.MarshalJSON()
	return string(normalized), err
}
