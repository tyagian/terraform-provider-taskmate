package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &TaskDataSource{}

func NewTaskDataSource() datasource.DataSource {
	return &TaskDataSource{}
}

// TaskDataSource defines the data source implementation.
type TaskDataSource struct {
	client *Client
}

// TaskDataSourceModel describes the data source data model.
type TaskDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	DueDate     types.String `tfsdk:"due_date"`
	Priority    types.String `tfsdk:"priority"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (d *TaskDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task"
}

func (d *TaskDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "TaskMate task data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Task identifier",
				Required:            true,
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Task title",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Task description",
				Computed:            true,
			},
			"due_date": schema.StringAttribute{
				MarkdownDescription: "Task due date",
				Computed:            true,
			},
			"priority": schema.StringAttribute{
				MarkdownDescription: "Task priority",
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Task status",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "Creation timestamp",
				Computed:            true,
			},
			"updated_at": schema.StringAttribute{
				MarkdownDescription: "Last update timestamp",
				Computed:            true,
			},
		},
	}
}

func (d *TaskDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *TaskDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TaskDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var id int
	_, err := fmt.Sscanf(data.ID.ValueString(), "%d", &id)
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse task ID: %s", err))
		return
	}

	task, err := d.client.GetTask(id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read task, got error: %s", err))
		return
	}

	data.Title = types.StringValue(task.Title)
	data.Description = types.StringValue(task.Description)
	data.DueDate = types.StringValue(task.DueDate)
	data.Priority = types.StringValue(task.Priority)
	data.Status = types.StringValue(task.Status)
	data.CreatedAt = types.StringValue(task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	data.UpdatedAt = types.StringValue(task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
