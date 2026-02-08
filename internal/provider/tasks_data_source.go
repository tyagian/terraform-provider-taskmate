package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &TasksDataSource{}

func NewTasksDataSource() datasource.DataSource {
	return &TasksDataSource{}
}

// TasksDataSource defines the data source implementation.
type TasksDataSource struct {
	client *Client
}

// TasksDataSourceModel describes the data source data model.
type TasksDataSourceModel struct {
	Tasks []TaskDataSourceModel `tfsdk:"tasks"`
	ID    types.String          `tfsdk:"id"`
}

func (d *TasksDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tasks"
}

func (d *TasksDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "TaskMate tasks data source - lists all tasks",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Placeholder identifier",
				Computed:            true,
			},
			"tasks": schema.ListNestedAttribute{
				MarkdownDescription: "List of all tasks",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "Task identifier",
							Computed:            true,
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
				},
			},
		},
	}
}

func (d *TasksDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *TasksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data TasksDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tasks, err := d.client.ListTasks()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to list tasks, got error: %s", err))
		return
	}

	// Convert tasks to data source model
	data.Tasks = make([]TaskDataSourceModel, len(tasks))
	for i, task := range tasks {
		data.Tasks[i] = TaskDataSourceModel{
			ID:          types.StringValue(fmt.Sprintf("%d", task.ID)),
			Title:       types.StringValue(task.Title),
			Description: types.StringValue(task.Description),
			DueDate:     types.StringValue(task.DueDate),
			Priority:    types.StringValue(task.Priority),
			Status:      types.StringValue(task.Status),
			CreatedAt:   types.StringValue(task.CreatedAt.Format("2006-01-02T15:04:05Z07:00")),
			UpdatedAt:   types.StringValue(task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00")),
		}
	}

	// Set a placeholder ID for the data source
	data.ID = types.StringValue("tasks")

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
