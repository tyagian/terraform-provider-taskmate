package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &TaskResource{}
var _ resource.ResourceWithImportState = &TaskResource{}

func NewTaskResource() resource.Resource {
	return &TaskResource{}
}

// TaskResource defines the resource implementation.
type TaskResource struct {
	client *Client
}

// TaskResourceModel describes the resource data model.
type TaskResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Title       types.String `tfsdk:"title"`
	Description types.String `tfsdk:"description"`
	DueDate     types.String `tfsdk:"due_date"`
	Priority    types.String `tfsdk:"priority"`
	Status      types.String `tfsdk:"status"`
	CreatedAt   types.String `tfsdk:"created_at"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
}

func (r *TaskResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_task"
}

func (r *TaskResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `TaskMate task resource

Manages a task in the TaskMate application.

## Example Usage

` + "```hcl" + `
resource "taskmate_task" "example" {
  title       = "Deploy application"
  description = "Deploy v2.0 to production"
  due_date    = "2024-12-31"
  priority    = "high"
  status      = "pending"
}
` + "```" + `

## Import

Tasks can be imported using their ID. To find task IDs, use the ` + "`taskmate_tasks`" + ` data source or query the API directly.

` + "```bash" + `
# List all tasks to find IDs
terraform apply -target=data.taskmate_tasks.all

# Import a task by ID
terraform import taskmate_task.example 1
` + "```" + `

See the examples/import directory for detailed import workflows.
`,

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Task identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"title": schema.StringAttribute{
				MarkdownDescription: "Task title",
				Required:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Task description",
				Optional:            true,
			},
			"due_date": schema.StringAttribute{
				MarkdownDescription: "Task due date (YYYY-MM-DD)",
				Optional:            true,
			},
			"priority": schema.StringAttribute{
				MarkdownDescription: "Task priority (low, medium, high)",
				Optional:            true,
				Computed:            true,
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "Task status (pending, completed)",
				Optional:            true,
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

func (r *TaskResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *TaskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data TaskResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	task, err := r.client.CreateTask(
		data.Title.ValueString(),
		data.Description.ValueString(),
		data.DueDate.ValueString(),
		data.Priority.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create task, got error: %s", err))
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%d", task.ID))
	data.Title = types.StringValue(task.Title)
	data.Description = types.StringValue(task.Description)
	data.DueDate = types.StringValue(task.DueDate)
	data.Priority = types.StringValue(task.Priority)
	data.Status = types.StringValue(task.Status)
	data.CreatedAt = types.StringValue(task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
	data.UpdatedAt = types.StringValue(task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *TaskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data TaskResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var id int
	_, err := fmt.Sscanf(data.ID.ValueString(), "%d", &id)
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse task ID: %s", err))
		return
	}

	task, err := r.client.GetTask(id)
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

func (r *TaskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data TaskResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var id int
	_, err := fmt.Sscanf(data.ID.ValueString(), "%d", &id)
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse task ID: %s", err))
		return
	}

	task, err := r.client.UpdateTask(
		id,
		data.Title.ValueString(),
		data.Description.ValueString(),
		data.DueDate.ValueString(),
		data.Priority.ValueString(),
		data.Status.ValueString(),
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update task, got error: %s", err))
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

func (r *TaskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data TaskResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	var id int
	_, err := fmt.Sscanf(data.ID.ValueString(), "%d", &id)
	if err != nil {
		resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse task ID: %s", err))
		return
	}

	err = r.client.DeleteTask(id)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete task, got error: %s", err))
		return
	}
}

func (r *TaskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
