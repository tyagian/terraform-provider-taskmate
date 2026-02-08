package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure TaskMateProvider satisfies various provider interfaces.
var _ provider.Provider = &TaskMateProvider{}

// TaskMateProvider defines the provider implementation.
type TaskMateProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance testing.
	version string
}

// TaskMateProviderModel describes the provider data model.
type TaskMateProviderModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

func (p *TaskMateProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "taskmate"
	resp.Version = p.version
}

func (p *TaskMateProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "TaskMate provider for managing tasks",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "TaskMate API host URL. Defaults to http://localhost:8080",
				Optional:            true,
			},
			"token": schema.StringAttribute{
				MarkdownDescription: "API token for authentication. Generate using: curl -X POST http://localhost:8080/api/v1/auth/token",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *TaskMateProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data TaskMateProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Set defaults
	host := data.Host.ValueString()
	if host == "" {
		host = "http://localhost:8080"
	}

	token := data.Token.ValueString()

	// Create API client
	client := NewClient(host, token)

	// Make client available to resources and data sources
	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *TaskMateProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewTaskResource,
	}
}

func (p *TaskMateProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewTaskDataSource,
		NewTasksDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &TaskMateProvider{
			version: version,
		}
	}
}
