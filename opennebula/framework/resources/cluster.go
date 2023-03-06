package resources

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/OpenNebula/one/src/oca/go/src/goca"
	dyn "github.com/OpenNebula/one/src/oca/go/src/goca/dynamic"
	"github.com/OpenNebula/one/src/oca/go/src/goca/parameters"
	"github.com/OpenNebula/terraform-provider-opennebula/opennebula/framework/common"
	"github.com/OpenNebula/terraform-provider-opennebula/opennebula/framework/config"
)

var _ resource.Resource = &Cluster{}
var _ resource.ResourceWithImportState = &Cluster{}

type ClusterModel struct {
	Id              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	Tags            types.Map    `tfsdk:"tags"`
	DefaultTags     types.Map    `tfsdk:"default_tags"`
	TagsAll         types.Map    `tfsdk:"tags_all"`
	TemplateSection types.Set    `tfsdk:"template_section"`
}

// TODO: use cluster controller
type Cluster struct {
	defaultTags map[string]string
	controller  *goca.Controller
}

func NewCluster() resource.Resource {
	return &Cluster{}
}

func (r *Cluster) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (r *Cluster) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Cluster resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				//	MarkdownDescription: "Example identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Name of the Cluster",
				//MarkdownDescription: "",
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": common.TagsSchema(),
			"default_tags": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Default tags defined in the provider configuration",
			},
			"tags_all": schema.MapAttribute{
				Computed:    true,
				ElementType: types.StringType,
				Description: "Result of the applied default_tags and resource tags",
			},
		},
		Blocks: map[string]schema.Block{
			"template_section": common.TemplateSectionBlock(),
		},
	}
}

func (r *Cluster) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	config, ok := req.ProviderData.(*config.Provider)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Cluster Configure Type",
			fmt.Sprintf("Expected *config.Provider, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.controller = config.Controller
	r.defaultTags = config.DefaultTags
}

func (r *Cluster) ModifyPlan(ctx context.Context, req resource.ModifyPlanRequest, resp *resource.ModifyPlanResponse) {
	var data *ClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	// convert tags from terraform elements to Go types
	var tags common.Tags

	element, err := data.Tags.ToTerraformValue(ctx)
	if err != nil {
		log.Print("[DEBUG] ToTerraformValue tags err: ", err)
	}
	err = element.As(&tags)
	if err != nil {
		log.Print("[DEBUG] As err: ", err)
	}

	// build tags_all map content
	newtagsAll := make(map[string]string, len(r.defaultTags))

	// copy default tags map
	for k, v := range r.defaultTags {
		newtagsAll[k] = v
	}

	for k, v := range tags.Elements {
		newtagsAll[k] = v
	}

	diags := resp.Plan.SetAttribute(ctx, path.Root("tags_all"), newtagsAll)
	if len(diags) > 0 {
		resp.Diagnostics.Append(diags...)
	}
}

func (r *Cluster) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data *ClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	clusterID, err := r.controller.Clusters().Create(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create the cluster",
			err.Error(),
		)
		return
	}

	controller := r.controller.Cluster(clusterID)

	// template management

	tpl := dyn.NewTemplate()

	sectionElements := data.TemplateSection.Elements()
	if len(sectionElements) > 0 {

		var templateSection common.TemplateSection
		for _, t := range sectionElements {
			element, err := t.ToTerraformValue(ctx)
			if err != nil {
				log.Print("[DEBUG] ToTerraformValue template section err: ", err)
				continue
			}
			err = element.As(&templateSection)
			if err != nil {
				log.Print("[DEBUG] As err: ", err)
				continue
			}
		}

		vec := tpl.AddVector(strings.ToUpper(templateSection.Name))
		for k, v := range templateSection.Elements {
			vec.AddPair(k, v)
		}
	}

	// add tags
	var tags common.Tags
	element, err := data.Tags.ToTerraformValue(ctx)
	if err != nil {
		log.Print("[DEBUG] ToTerraformValue tags err: ", err)
	}

	err = element.As(&tags)
	if err != nil {
		log.Print("[DEBUG] As err: ", err)
	}
	for k, v := range tags.Elements {
		key := strings.ToUpper(k)
		tpl.AddPair(key, v)
	}

	// add default tags if they aren't overriden
	if len(r.defaultTags) > 0 {
		for k, v := range r.defaultTags {
			key := strings.ToUpper(k)
			p, _ := tpl.GetPair(key)
			if p != nil {
				continue
			}
			tpl.AddPair(key, v)
		}
	}

	if len(tpl.Elements) > 0 {
		log.Printf("[DEBUG] Cluster update: %s", tpl.String())
		err = controller.Update(tpl.String(), parameters.Merge)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update the cluster content",
				fmt.Sprintf("cluster (ID: %d): %s", clusterID, err),
			)
			return
		}
	}

	// Convert from the API data model to the Terraform data model
	// and set any unknown attribute values.
	data.Id = types.StringValue(fmt.Sprint(clusterID))

	// fill default tags
	// TODO: store an intermediary representation of default tags
	// instead of converting values back here ?
	var diags diag.Diagnostics
	var defaultTags map[string]attr.Value

	if len(r.defaultTags) > 0 {
		defaultTags = make(map[string]attr.Value)
		for k, v := range r.defaultTags {
			defaultTags[k] = types.StringValue(v)
		}
	}

	data.DefaultTags, diags = types.MapValue(types.StringType, defaultTags)
	if len(diags) > 0 {
		resp.Diagnostics = append(resp.Diagnostics, diags...)
		return
	}
	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *Cluster) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data *ClusterModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	log.Print("[DEBUG] Cluster reading...")

	id64, err := strconv.ParseInt(data.Id.ValueString(), 10, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse cluster ID",
			err.Error(),
		)
		return
	}
	id := int(id64)
	clusterInfos, err := r.controller.Cluster(id).Info()
	if err != nil {

		if NoExists(err) {
			resp.Diagnostics.AddError(
				"Failed to retrieve the cluster",
				fmt.Sprintf("cluster (ID: %d): %s", id, err),
			)
			log.Printf("[WARN] Removing cluster %s from state because it no longer exists in", data.Name)
			return

		}
		resp.Diagnostics.AddError(
			"Failed retrieve cluster informations",
			fmt.Sprintf("cluster (ID: %d): %s", id, err),
		)
		return
	}

	data.Name = types.StringValue(clusterInfos.Name)

	// read tags
	// Retrieve and copy the tags names from the configuration then fill value with thoses from remote cluster
	var stateTags common.Tags
	element, err := data.Tags.ToTerraformValue(ctx)
	if err != nil {
		log.Print("[DEBUG] ToTerraformValue state tags err: ", err)
	}
	err = element.As(&stateTags)
	if err != nil {
		log.Print("[DEBUG] As err: ", err)
	}

	if len(stateTags.Elements) > 0 {
		readTags := make(map[string]attr.Value)

		for k, _ := range stateTags.Elements {
			v, err := clusterInfos.Template.GetStr(strings.ToUpper(k))
			if err != nil {
				continue
			}
			readTags[k] = types.StringValue(v)
		}

		var diags diag.Diagnostics
		data.Tags, diags = types.MapValue(types.StringType, readTags)
		if len(diags) > 0 {
			resp.Diagnostics = append(resp.Diagnostics, diags...)
			return
		}
	}

	//data.TagsAll
	//data.TemplateSection

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r *Cluster) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state *ClusterModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	log.Print("[DEBUG] Cluster reading...")

	id64, err := strconv.ParseInt(plan.Id.ValueString(), 10, 0)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to parse cluster ID",
			err.Error(),
		)
		return
	}
	id := int(id64)
	controller := r.controller.Cluster(id)
	clusterInfos, err := controller.Info()
	if err != nil {

		if NoExists(err) {
			resp.Diagnostics.AddError(
				"Failed to retrieve the cluster",
				fmt.Sprintf("cluster (ID: %d): %s", id, err),
			)
			log.Printf("[WARN] Removing cluster %s from state because it no longer exists in", plan.Name)
			return

		}
		resp.Diagnostics.AddError(
			"Failed retrieve cluster informations",
			fmt.Sprintf("cluster (ID: %d): %s", id, err),
		)
		return
	}

	// XXX
	update := false
	newTpl := clusterInfos.Template
	if !plan.TemplateSection.Equal(state.TemplateSection) {
		// updateTemplateSection(d, &newTpl.Template)
		update = true
	}

	// tags
	// tags_all

	if update {
		err = controller.Update(newTpl.String(), parameters.Replace)
		if err != nil {
			resp.Diagnostics.AddError(
				"Failed to update cluster content",
				fmt.Sprintf("cluster (ID: %d): %s", id, err),
			)
			return
		}

	}

}

func (r *Cluster) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}

func (r *Cluster) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
}
