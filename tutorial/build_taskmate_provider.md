# Building a Terraform Provider: Complete Guide

## Table of Contents
1. [Introduction](#introduction)
2. [Architecture Overview](#architecture-overview)
3. [Provider Entry Point](#provider-entry-point)
4. [Provider Configuration](#provider-configuration)
5. [API Client](#api-client)
6. [Resource Implementation](#resource-implementation)
7. [Data Source Implementation](#data-source-implementation)
8. [Request Flow Diagrams](#request-flow-diagrams)
9. [Parameter Reference](#parameter-reference)

---

## Introduction

This tutorial explains how to build a Terraform provider using the **Terraform Plugin Framework**. We'll use the TaskMate provider as an example, which manages tasks through a REST API.

### What You'll Learn
- Provider architecture and component interaction
- Request/response flow through the provider
- Parameter meanings and usage
- CRUD operation implementation
- State management patterns

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Terraform Core                          │
│  (Reads .tf files, manages state, orchestrates operations)  │
└────────────────────────┬────────────────────────────────────┘
                         │ gRPC Protocol
                         │
┌────────────────────────▼────────────────────────────────────┐
│                  Provider Binary                            │
│  ┌──────────────────────────────────────────────────────┐   │
│  │  main.go - Entry Point                               │   │
│  │  • Starts gRPC server                                │   │
│  │  • Registers provider with Terraform                 │   │
│  └──────────────────────┬───────────────────────────────┘   │
│                         │                                    │
│  ┌──────────────────────▼───────────────────────────────┐   │
│  │  provider.go - Provider Logic                        │   │
│  │  • Schema definition (host, api_key)                 │   │
│  │  • Configuration parsing                             │   │
│  │  • Client initialization                             │   │
│  │  • Resource/DataSource registration                  │   │
│  └──────────────────────┬───────────────────────────────┘   │
│                         │                                    │
│         ┌───────────────┴───────────────┐                   │
│         │                               │                   │
│  ┌──────▼──────────┐           ┌───────▼──────────┐        │
│  │ task_resource.go│           │task_data_source.go│        │
│  │ • CRUD ops      │           │ • Read-only      │        │
│  │ • Schema        │           │ • Schema         │        │
│  │ • State mgmt    │           │ • State mgmt     │        │
│  └──────┬──────────┘           └───────┬──────────┘        │
│         │                               │                   │
│         └───────────────┬───────────────┘                   │
│                         │                                    │
│  ┌──────────────────────▼───────────────────────────────┐   │
│  │  client.go - API Client                              │   │
│  │  • HTTP request builder                              │   │
│  │  • Authentication (API key)                          │   │
│  │  • JSON marshaling/unmarshaling                      │   │
│  │  • Error handling                                    │   │
│  └──────────────────────┬───────────────────────────────┘   │
└─────────────────────────┼───────────────────────────────────┘
                          │ HTTP/REST
                          │
┌─────────────────────────▼───────────────────────────────────┐
│                  TaskMate API Server                        │
│  • Handles CRUD operations                                  │
│  • Persists data to tasks.json                              │
│  • Returns JSON responses                                   │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

| Component | Purpose | Key Functions |
|-----------|---------|---------------|
| **main.go** | Provider entry point | Start gRPC server, register provider |
| **provider.go** | Provider configuration | Parse config, create client, register resources |
| **client.go** | API communication | HTTP requests, authentication, error handling |
| **task_resource.go** | Task lifecycle | Create, Read, Update, Delete, Import |
| **task_data_source.go** | Read-only access | Fetch existing task data |

---

## Provider Entry Point

### File: `main.go`

```go
package main

import (
    "context"
    "flag"
    "log"
    
    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/hashicorp/terraform-provider-taskmate/internal/provider"
)

var (
    version string = "dev"
)

func main() {
    var debug bool
    
    flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
    flag.Parse()
    
    opts := providerserver.ServeOpts{
        Address: "registry.terraform.io/hashitalk/taskmate",
        Debug:   debug,
    }
    
    err := providerserver.Serve(context.Background(), provider.New(version), opts)
    
    if err != nil {
        log.Fatal(err.Error())
    }
}
```

### Parameter Breakdown

| Parameter | Type | Purpose | Example |
|-----------|------|---------|---------|
| `debug` | bool | Enable debugger support (delve) | `--debug` flag |
| `Address` | string | Provider registry address | `registry.terraform.io/hashitalk/taskmate` |
| `Debug` | bool | Enable debug mode in server | `true` or `false` |
| `version` | string | Provider version | `"1.0.0"` or `"dev"` |

### What Happens Here?

1. **Parse Flags**: Check if `--debug` flag is passed
2. **Configure Server**: Set provider address and debug mode
3. **Start Server**: Call `providerserver.Serve()` which:
   - Creates a gRPC server
   - Registers the provider with Terraform
   - Listens for Terraform commands
4. **Block**: Server runs until Terraform terminates it

---

## Provider Configuration

### File: `provider.go`

```go
type TaskMateProvider struct {
    version string
}

type TaskMateProviderModel struct {
    Host   types.String `tfsdk:"host"`
    ApiKey types.String `tfsdk:"api_key"`
}
```

### Schema Definition

```go
func (p *TaskMateProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "TaskMate provider for managing tasks",
        Attributes: map[string]schema.Attribute{
            "host": schema.StringAttribute{
                MarkdownDescription: "TaskMate API host URL. Defaults to http://localhost:8080",
                Optional:            true,
            },
            "api_key": schema.StringAttribute{
                MarkdownDescription: "API key for authentication. Defaults to todo-secret-key",
                Optional:            true,
                Sensitive:           true,
            },
        },
    }
}
```

### Schema Attribute Properties

| Property | Purpose | Values |
|----------|---------|--------|
| `Optional` | Field not required | `true` = can be omitted |
| `Required` | Field must be provided | `true` = must be set |
| `Computed` | Value set by provider | `true` = read from API |
| `Sensitive` | Hide in logs/output | `true` = masked in output |
| `MarkdownDescription` | Documentation | String describing the field |

### Configure Method

```go
func (p *TaskMateProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
    var data TaskMateProviderModel
    
    // Parse configuration from .tf file
    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    
    if resp.Diagnostics.HasError() {
        return
    }
    
    // Set defaults
    host := data.Host.ValueString()
    if host == "" {
        host = "http://localhost:8080"
    }
    
    apiKey := data.ApiKey.ValueString()
    if apiKey == "" {
        apiKey = "todo-secret-key"
    }
    
    // Create API client
    client := NewClient(host, apiKey)
    
    // Make client available to resources and data sources
    resp.DataSourceData = client
    resp.ResourceData = client
}
```

### Configure Parameters

| Parameter | Type | Purpose |
|-----------|------|---------|
| `ctx` | context.Context | Request context for cancellation |
| `req.Config` | Config | User's provider block from .tf file |
| `resp.Diagnostics` | Diagnostics | Error/warning collection |
| `resp.ResourceData` | interface{} | Data passed to resources |
| `resp.DataSourceData` | interface{} | Data passed to data sources |

### Configuration Flow

```
User's .tf file
    ↓
provider "taskmate" {
  host    = "http://localhost:8080"
  api_key = "todo-secret-key"
}
    ↓
Configure() method called
    ↓
Parse config into TaskMateProviderModel
    ↓
Apply defaults if values missing
    ↓
Create Client with host + api_key
    ↓
Store client in resp.ResourceData
    ↓
Resources/DataSources receive client
```

---


## API Client

### File: `client.go`

The client handles all HTTP communication with the TaskMate API.

```go
type Client struct {
    Host   string
    ApiKey string
    client *http.Client
}

type Task struct {
    ID          int       `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    DueDate     string    `json:"due_date"`
    Priority    string    `json:"priority"`
    Status      string    `json:"status"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

### Client Initialization

```go
func NewClient(host, apiKey string) *Client {
    return &Client{
        Host:   host,
        ApiKey: apiKey,
        client: &http.Client{
            Timeout: 30 * time.Second,
        },
    }
}
```

| Parameter | Purpose | Example |
|-----------|---------|---------|
| `host` | API base URL | `"http://localhost:8080"` |
| `apiKey` | Authentication token | `"todo-secret-key"` |
| `Timeout` | Max request duration | `30 * time.Second` |

### makeRequest Helper

```go
func (c *Client) makeRequest(method, path string, body interface{}) (*http.Response, error) {
    url := c.Host + "/api/v1" + path
    
    var reqBody io.Reader
    if body != nil {
        jsonBody, err := json.Marshal(body)
        if err != nil {
            return nil, fmt.Errorf("failed to marshal request body: %w", err)
        }
        reqBody = bytes.NewBuffer(jsonBody)
    }
    
    req, err := http.NewRequest(method, url, reqBody)
    if err != nil {
        return nil, fmt.Errorf("failed to create request: %w", err)
    }
    
    req.Header.Set("X-API-Key", c.ApiKey)
    req.Header.Set("Content-Type", "application/json")
    
    return c.client.Do(req)
}
```

### makeRequest Parameters

| Parameter | Type | Purpose | Example |
|-----------|------|---------|---------|
| `method` | string | HTTP method | `"GET"`, `"POST"`, `"PUT"`, `"DELETE"` |
| `path` | string | API endpoint path | `"/tasks"`, `"/tasks/1"` |
| `body` | interface{} | Request payload | `map[string]string{"title": "Task"}` |

### CreateTask

```go
func (c *Client) CreateTask(title, description, dueDate, priority string) (*Task, error) {
    reqBody := map[string]string{
        "title":       title,
        "description": description,
        "due_date":    dueDate,
        "priority":    priority,
    }
    
    resp, err := c.makeRequest("POST", "/tasks", reqBody)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusCreated {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API error: %s (status: %d)", string(body), resp.StatusCode)
    }
    
    var task Task
    if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &task, nil
}
```

### CreateTask Flow

```
Input: title, description, due_date, priority
    ↓
Build request body map
    ↓
POST /api/v1/tasks with JSON body
    ↓
Check status code (expect 201 Created)
    ↓
Decode JSON response into Task struct
    ↓
Return Task with ID assigned by API
```

### GetTask

```go
func (c *Client) GetTask(id int) (*Task, error) {
    resp, err := c.makeRequest("GET", fmt.Sprintf("/tasks/%d", id), nil)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("task with ID %d not found", id)
    }
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API error: %s (status: %d)", string(body), resp.StatusCode)
    }
    
    var task Task
    if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &task, nil
}
```

### GetTask Flow

```
Input: task ID (integer)
    ↓
GET /api/v1/tasks/{id}
    ↓
Check status code
  • 404 → Task not found error
  • 200 → Success, decode response
  • Other → API error
    ↓
Return Task struct
```

### UpdateTask

```go
func (c *Client) UpdateTask(id int, title, description, dueDate, priority, status string) (*Task, error) {
    reqBody := map[string]string{
        "title":       title,
        "description": description,
        "due_date":    dueDate,
        "priority":    priority,
        "status":      status,
    }
    
    resp, err := c.makeRequest("PUT", fmt.Sprintf("/tasks/%d", id), reqBody)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusNotFound {
        return nil, fmt.Errorf("task with ID %d not found", id)
    }
    
    if resp.StatusCode != http.StatusOK {
        body, _ := io.ReadAll(resp.Body)
        return nil, fmt.Errorf("API error: %s (status: %d)", string(body), resp.StatusCode)
    }
    
    var task Task
    if err := json.NewDecoder(resp.Body).Decode(&task); err != nil {
        return nil, fmt.Errorf("failed to decode response: %w", err)
    }
    
    return &task, nil
}
```

### UpdateTask Parameters

| Parameter | Purpose | Example |
|-----------|---------|---------|
| `id` | Task identifier | `1`, `42` |
| `title` | Task title | `"Deploy to Production"` |
| `description` | Task details | `"Deploy v2.0 release"` |
| `dueDate` | Due date (YYYY-MM-DD) | `"2024-12-31"` |
| `priority` | Priority level | `"low"`, `"medium"`, `"high"` |
| `status` | Task status | `"pending"`, `"completed"` |

### DeleteTask

```go
func (c *Client) DeleteTask(id int) error {
    resp, err := c.makeRequest("DELETE", fmt.Sprintf("/tasks/%d", id), nil)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusNotFound {
        return fmt.Errorf("task with ID %d not found", id)
    }
    
    if resp.StatusCode != http.StatusNoContent {
        body, _ := io.ReadAll(resp.Body)
        return fmt.Errorf("API error: %s (status: %d)", string(body), resp.StatusCode)
    }
    
    return nil
}
```

### DeleteTask Flow

```
Input: task ID
    ↓
DELETE /api/v1/tasks/{id}
    ↓
Check status code
  • 404 → Task not found error
  • 204 → Success (No Content)
  • Other → API error
    ↓
Return nil (success) or error
```

### HTTP Status Codes

| Code | Meaning | Used In |
|------|---------|---------|
| 200 OK | Success with body | GET, PUT |
| 201 Created | Resource created | POST |
| 204 No Content | Success, no body | DELETE |
| 404 Not Found | Resource doesn't exist | GET, PUT, DELETE |
| 400 Bad Request | Invalid input | POST, PUT |
| 401 Unauthorized | Invalid API key | All methods |

---


## Resource Implementation

### File: `task_resource.go`

Resources manage the full lifecycle of infrastructure: Create, Read, Update, Delete (CRUD).

```go
type TaskResource struct {
    client *Client
}

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
```

### Model Field Tags

The `tfsdk` tag maps Go struct fields to Terraform attributes:

```go
ID types.String `tfsdk:"id"`
```

Maps to Terraform:
```hcl
resource "taskmate_task" "example" {
  id = "1"  # This field
}
```

### Schema Definition

```go
func (r *TaskResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "TaskMate task resource",
        
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
```

### Attribute Flags Explained

| Flag | Meaning | Example Field |
|------|---------|---------------|
| `Required: true` | User must provide value | `title` |
| `Optional: true` | User can provide value | `description`, `due_date` |
| `Computed: true` | Provider sets value | `id`, `created_at`, `updated_at` |
| `Optional + Computed` | User can set or provider defaults | `priority`, `status` |

### Plan Modifiers

```go
PlanModifiers: []planmodifier.String{
    stringplanmodifier.UseStateForUnknown(),
}
```

**Purpose**: During `terraform plan`, if the value is unknown (e.g., ID not yet created), use the value from state instead of showing it as changing.

**Effect**: Prevents unnecessary "forces replacement" messages for computed fields.

### Configure Method

```go
func (r *TaskResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }
    
    client, ok := req.ProviderData.(*Client)
    
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *Client, got: %T", req.ProviderData),
        )
        return
    }
    
    r.client = client
}
```

### Configure Flow

```
Provider.Configure() sets resp.ResourceData = client
    ↓
Terraform calls Resource.Configure()
    ↓
req.ProviderData contains the client
    ↓
Type assert to *Client
    ↓
Store in r.client for use in CRUD methods
```

### Create Method

```go
func (r *TaskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data TaskResourceModel
    
    // Read Terraform plan data into the model
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    
    if resp.Diagnostics.HasError() {
        return
    }
    
    // Call API to create task
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
    
    // Map API response to model
    data.ID = types.StringValue(fmt.Sprintf("%d", task.ID))
    data.Title = types.StringValue(task.Title)
    data.Description = types.StringValue(task.Description)
    data.DueDate = types.StringValue(task.DueDate)
    data.Priority = types.StringValue(task.Priority)
    data.Status = types.StringValue(task.Status)
    data.CreatedAt = types.StringValue(task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
    data.UpdatedAt = types.StringValue(task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
    
    // Save data into Terraform state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### Create Parameters

| Parameter | Type | Purpose |
|-----------|------|---------|
| `ctx` | context.Context | Request context |
| `req.Plan` | Plan | User's desired state from .tf file |
| `resp.State` | State | Where to save actual state |
| `resp.Diagnostics` | Diagnostics | Error/warning collection |

### Create Flow

```
User runs: terraform apply
    ↓
Terraform reads resource block from .tf file
    ↓
Calls Resource.Create()
    ↓
req.Plan contains user's config:
  resource "taskmate_task" "example" {
    title       = "Deploy"
    description = "Deploy v2.0"
    due_date    = "2024-12-31"
    priority    = "high"
  }
    ↓
Extract values from req.Plan into TaskResourceModel
    ↓
Call client.CreateTask() with values
    ↓
API returns Task with ID=1, status="pending", timestamps
    ↓
Map all fields (including computed ones) to model
    ↓
Save model to resp.State
    ↓
Terraform stores state in terraform.tfstate
```

### Read Method

```go
func (r *TaskResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var data TaskResourceModel
    
    // Read current state
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    
    if resp.Diagnostics.HasError() {
        return
    }
    
    // Parse ID from state
    var id int
    _, err := fmt.Sscanf(data.ID.ValueString(), "%d", &id)
    if err != nil {
        resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse task ID: %s", err))
        return
    }
    
    // Fetch current state from API
    task, err := r.client.GetTask(id)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read task, got error: %s", err))
        return
    }
    
    // Update model with API response
    data.Title = types.StringValue(task.Title)
    data.Description = types.StringValue(task.Description)
    data.DueDate = types.StringValue(task.DueDate)
    data.Priority = types.StringValue(task.Priority)
    data.Status = types.StringValue(task.Status)
    data.CreatedAt = types.StringValue(task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
    data.UpdatedAt = types.StringValue(task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
    
    // Save updated state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### Read Flow

```
User runs: terraform refresh or terraform plan
    ↓
Terraform calls Resource.Read()
    ↓
req.State contains current state (ID=1)
    ↓
Extract ID from state
    ↓
Call client.GetTask(1)
    ↓
API returns current task data
    ↓
Update model with fresh data
    ↓
Save to resp.State
    ↓
Terraform detects drift if values changed
```

### When Read is Called

1. **terraform refresh**: Explicitly sync state with reality
2. **terraform plan**: Before showing changes
3. **After Create/Update**: Verify operation succeeded
4. **Before Update/Delete**: Get current state

### Update Method

```go
func (r *TaskResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var data TaskResourceModel
    
    // Read Terraform plan data
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    
    if resp.Diagnostics.HasError() {
        return
    }
    
    // Parse ID
    var id int
    _, err := fmt.Sscanf(data.ID.ValueString(), "%d", &id)
    if err != nil {
        resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse task ID: %s", err))
        return
    }
    
    // Call API to update task
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
    
    // Update model with API response
    data.Title = types.StringValue(task.Title)
    data.Description = types.StringValue(task.Description)
    data.DueDate = types.StringValue(task.DueDate)
    data.Priority = types.StringValue(task.Priority)
    data.Status = types.StringValue(task.Status)
    data.CreatedAt = types.StringValue(task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
    data.UpdatedAt = types.StringValue(task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
    
    // Save updated state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### Update Flow

```
User modifies .tf file:
  resource "taskmate_task" "example" {
    title  = "Deploy"
    status = "completed"  # Changed from "pending"
  }
    ↓
User runs: terraform apply
    ↓
Terraform detects change in plan
    ↓
Calls Resource.Update()
    ↓
req.Plan contains new desired state
    ↓
Extract ID and all fields from plan
    ↓
Call client.UpdateTask() with new values
    ↓
API returns updated task
    ↓
Save to resp.State
```

### Delete Method

```go
func (r *TaskResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var data TaskResourceModel
    
    // Read current state
    resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
    
    if resp.Diagnostics.HasError() {
        return
    }
    
    // Parse ID
    var id int
    _, err := fmt.Sscanf(data.ID.ValueString(), "%d", &id)
    if err != nil {
        resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse task ID: %s", err))
        return
    }
    
    // Call API to delete task
    err = r.client.DeleteTask(id)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete task, got error: %s", err))
        return
    }
    
    // State is automatically removed by Terraform
}
```

### Delete Flow

```
User removes resource from .tf file or runs: terraform destroy
    ↓
Terraform calls Resource.Delete()
    ↓
req.State contains current state (ID=1)
    ↓
Extract ID from state
    ↓
Call client.DeleteTask(1)
    ↓
API deletes task, returns 204 No Content
    ↓
Return without error
    ↓
Terraform removes resource from state
```

### ImportState Method

```go
func (r *TaskResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

### Import Flow

```
User runs: terraform import taskmate_task.example 1
    ↓
Terraform calls Resource.ImportState()
    ↓
req.ID = "1" (from command line)
    ↓
ImportStatePassthroughID sets state.id = "1"
    ↓
Terraform calls Resource.Read()
    ↓
Read fetches full task data from API
    ↓
State is populated with all fields
```

### Import Command

```bash
terraform import taskmate_task.example 1
```

| Part | Meaning |
|------|---------|
| `taskmate_task` | Resource type |
| `example` | Resource name in .tf file |
| `1` | Task ID to import |

---


## Data Source Implementation

### File: `task_data_source.go`

Data sources provide read-only access to existing resources.

```go
type TaskDataSource struct {
    client *Client
}

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
```

### Schema Definition

```go
func (d *TaskDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "TaskMate task data source",
        
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                MarkdownDescription: "Task identifier",
                Required:            true,  // User must provide ID
            },
            "title": schema.StringAttribute{
                MarkdownDescription: "Task title",
                Computed:            true,  // Fetched from API
            },
            "description": schema.StringAttribute{
                MarkdownDescription: "Task description",
                Computed:            true,
            },
            // ... all other fields are Computed
        },
    }
}
```

### Key Differences from Resource

| Aspect | Resource | Data Source |
|--------|----------|-------------|
| Purpose | Manage lifecycle | Read existing data |
| ID | Computed (created by API) | Required (user provides) |
| Other fields | Required/Optional | Computed |
| Methods | Create, Read, Update, Delete | Read only |
| State | Managed by Terraform | Read-only reference |

### Read Method

```go
func (d *TaskDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data TaskDataSourceModel
    
    // Read configuration (user's data block)
    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    
    if resp.Diagnostics.HasError() {
        return
    }
    
    // Parse ID from config
    var id int
    _, err := fmt.Sscanf(data.ID.ValueString(), "%d", &id)
    if err != nil {
        resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Unable to parse task ID: %s", err))
        return
    }
    
    // Fetch task from API
    task, err := d.client.GetTask(id)
    if err != nil {
        resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read task, got error: %s", err))
        return
    }
    
    // Populate all computed fields
    data.Title = types.StringValue(task.Title)
    data.Description = types.StringValue(task.Description)
    data.DueDate = types.StringValue(task.DueDate)
    data.Priority = types.StringValue(task.Priority)
    data.Status = types.StringValue(task.Status)
    data.CreatedAt = types.StringValue(task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))
    data.UpdatedAt = types.StringValue(task.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"))
    
    // Save to state
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### Data Source Flow

```
User's .tf file:
  data "taskmate_task" "existing" {
    id = "1"
  }
    ↓
User runs: terraform plan or terraform apply
    ↓
Terraform calls DataSource.Read()
    ↓
req.Config contains id = "1"
    ↓
Extract ID from config
    ↓
Call client.GetTask(1)
    ↓
API returns task data
    ↓
Populate all computed fields
    ↓
Save to resp.State
    ↓
User can reference: data.taskmate_task.existing.title
```

### Usage Example

```hcl
# Create a task (managed resource)
resource "taskmate_task" "deployment" {
  title       = "Deploy to Production"
  description = "Deploy v2.0"
  priority    = "high"
}

# Read an existing task (data source)
data "taskmate_task" "existing" {
  id = "42"  # Task created outside Terraform
}

# Reference data source attributes
output "existing_task_title" {
  value = data.taskmate_task.existing.title
}

# Reference resource attributes directly
output "new_task_id" {
  value = taskmate_task.deployment.id
}
```

### When to Use Data Sources

| Scenario | Use |
|----------|-----|
| Task created by Terraform | Reference resource directly: `taskmate_task.name.id` |
| Task created outside Terraform | Use data source: `data.taskmate_task.name.id` |
| Need to fetch external data | Use data source |
| Managing full lifecycle | Use resource |

---

## Request Flow Diagrams

### Complete Create Flow

```
┌─────────────────────────────────────────────────────────────┐
│ User writes .tf file                                        │
│                                                             │
│ resource "taskmate_task" "example" {                        │
│   title       = "Deploy"                                    │
│   description = "Deploy v2.0"                               │
│   priority    = "high"                                      │
│ }                                                           │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ terraform apply
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ Terraform Core                                              │
│ • Parses .tf files                                          │
│ • Builds dependency graph                                   │
│ • Determines operations needed                              │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ gRPC: Create request
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ Provider Binary (main.go)                                   │
│ • Receives gRPC request                                     │
│ • Routes to appropriate resource                            │
└────────────────────────┬────────────────────────────────────┘
                         │
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ TaskResource.Create()                                       │
│ 1. req.Plan.Get() → Extract user config                     │
│    • title = "Deploy"                                       │
│    • description = "Deploy v2.0"                            │
│    • priority = "high"                                      │
│                                                             │
│ 2. Validate data                                            │
│                                                             │
│ 3. Call client.CreateTask()                                 │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ HTTP POST
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ Client.CreateTask()                                         │
│ 1. Build request body:                                      │
│    {                                                        │
│      "title": "Deploy",                                     │
│      "description": "Deploy v2.0",                          │
│      "priority": "high"                                     │
│    }                                                        │
│                                                             │
│ 2. makeRequest("POST", "/tasks", body)                      │
│    • URL: http://localhost:8080/api/v1/tasks               │
│    • Header: X-API-Key: todo-secret-key                     │
│    • Header: Content-Type: application/json                 │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ HTTP Request
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ TaskMate API Server                                         │
│ 1. Authenticate (check X-API-Key)                           │
│ 2. Parse JSON body                                          │
│ 3. Generate ID = 1                                          │
│ 4. Set status = "pending"                                   │
│ 5. Set timestamps                                           │
│ 6. Save to tasks.json                                       │
│ 7. Return 201 Created with JSON:                            │
│    {                                                        │
│      "id": 1,                                               │
│      "title": "Deploy",                                     │
│      "description": "Deploy v2.0",                          │
│      "priority": "high",                                    │
│      "status": "pending",                                   │
│      "created_at": "2024-01-15T10:30:00Z",                  │
│      "updated_at": "2024-01-15T10:30:00Z"                   │
│    }                                                        │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ HTTP Response
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ Client.CreateTask()                                         │
│ 1. Check status code (201 = success)                        │
│ 2. json.Decode() response into Task struct                  │
│ 3. Return &Task{...}                                        │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ Return Task
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ TaskResource.Create()                                       │
│ 4. Map API response to model:                               │
│    data.ID = "1"                                            │
│    data.Title = "Deploy"                                    │
│    data.Description = "Deploy v2.0"                         │
│    data.Priority = "high"                                   │
│    data.Status = "pending"                                  │
│    data.CreatedAt = "2024-01-15T10:30:00Z"                  │
│    data.UpdatedAt = "2024-01-15T10:30:00Z"                  │
│                                                             │
│ 5. resp.State.Set() → Save to Terraform state               │
└────────────────────────┬────────────────────────────────────┘
                         │
                         │ gRPC: Create response
                         ▼
┌─────────────────────────────────────────────────────────────┐
│ Terraform Core                                              │
│ • Receives state data                                       │
│ • Writes to terraform.tfstate:                              │
│   {                                                         │
│     "resources": [{                                         │
│       "type": "taskmate_task",                              │
│       "name": "example",                                    │
│       "instances": [{                                       │
│         "attributes": {                                     │
│           "id": "1",                                        │
│           "title": "Deploy",                                │
│           "status": "pending",                              │
│           ...                                               │
│         }                                                   │
│       }]                                                    │
│     }]                                                      │
│   }                                                         │
│ • Displays success message                                  │
└─────────────────────────────────────────────────────────────┘
```

### Update Flow

```
User modifies .tf file:
  status = "completed"
    ↓
terraform apply
    ↓
Terraform Core
  • Reads current state (status = "pending")
  • Compares with desired state (status = "completed")
  • Detects change
    ↓
Calls TaskResource.Read()
  • Fetches current state from API
  • Confirms status = "pending"
    ↓
Calls TaskResource.Update()
  • req.Plan contains new values
  • Extracts ID = "1" from state
  • Calls client.UpdateTask(1, ..., status="completed")
    ↓
Client.UpdateTask()
  • PUT /api/v1/tasks/1
  • Body: {"title": "Deploy", ..., "status": "completed"}
    ↓
TaskMate API
  • Updates task in memory and tasks.json
  • Returns updated task with new updated_at
    ↓
TaskResource.Update()
  • Maps response to model
  • Saves to resp.State
    ↓
Terraform Core
  • Updates terraform.tfstate
  • Shows: taskmate_task.example: Modifications complete
```

### Read (Refresh) Flow

```
terraform refresh
    ↓
Terraform Core
  • Reads terraform.tfstate
  • For each resource, calls Read()
    ↓
TaskResource.Read()
  • req.State contains ID = "1"
  • Calls client.GetTask(1)
    ↓
Client.GetTask()
  • GET /api/v1/tasks/1
    ↓
TaskMate API
  • Fetches task from memory
  • Returns current state
    ↓
TaskResource.Read()
  • Compares API response with state
  • If different, updates state
  • If task deleted (404), removes from state
    ↓
Terraform Core
  • Updates terraform.tfstate with fresh data
  • Shows drift if detected
```

### Delete Flow

```
terraform destroy
    ↓
Terraform Core
  • Reads terraform.tfstate
  • Determines deletion order (reverse dependencies)
    ↓
Calls TaskResource.Delete()
  • req.State contains ID = "1"
  • Calls client.DeleteTask(1)
    ↓
Client.DeleteTask()
  • DELETE /api/v1/tasks/1
    ↓
TaskMate API
  • Removes task from memory and tasks.json
  • Returns 204 No Content
    ↓
TaskResource.Delete()
  • Returns without error
  • Does NOT call resp.State.Set()
    ↓
Terraform Core
  • Removes resource from terraform.tfstate
  • Shows: taskmate_task.example: Destruction complete
```

### Data Source Flow

```
data "taskmate_task" "existing" {
  id = "42"
}
    ↓
terraform plan
    ↓
Terraform Core
  • Detects data source
  • Calls TaskDataSource.Read()
    ↓
TaskDataSource.Read()
  • req.Config contains id = "42"
  • Calls client.GetTask(42)
    ↓
Client.GetTask()
  • GET /api/v1/tasks/42
    ↓
TaskMate API
  • Returns task data
    ↓
TaskDataSource.Read()
  • Populates all computed fields
  • Saves to resp.State
    ↓
Terraform Core
  • Stores in state (not persisted to .tfstate)
  • Makes available for reference:
    data.taskmate_task.existing.title
```

---


## Parameter Reference

### Context Parameters

Every method receives a `context.Context` as the first parameter:

```go
func (r *TaskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse)
```

| Parameter | Purpose | Usage |
|-----------|---------|-------|
| `ctx` | Request context | Check cancellation: `ctx.Done()` |
| | | Pass to API calls for timeout control |
| | | Required by framework methods |

### Request Objects

#### resource.CreateRequest

```go
type CreateRequest struct {
    Plan tfsdk.Plan  // User's desired state from .tf file
}
```

**Usage**:
```go
var data TaskResourceModel
resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
```

#### resource.ReadRequest

```go
type ReadRequest struct {
    State tfsdk.State  // Current state from terraform.tfstate
}
```

**Usage**:
```go
var data TaskResourceModel
resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
```

#### resource.UpdateRequest

```go
type UpdateRequest struct {
    Plan  tfsdk.Plan   // New desired state
    State tfsdk.State  // Current state
}
```

**Usage**:
```go
var data TaskResourceModel
resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)  // Get new values
```

#### resource.DeleteRequest

```go
type DeleteRequest struct {
    State tfsdk.State  // Current state (contains ID to delete)
}
```

**Usage**:
```go
var data TaskResourceModel
resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
```

#### datasource.ReadRequest

```go
type ReadRequest struct {
    Config tfsdk.Config  // User's data block configuration
}
```

**Usage**:
```go
var data TaskDataSourceModel
resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
```

### Response Objects

#### resource.CreateResponse

```go
type CreateResponse struct {
    State       tfsdk.State
    Diagnostics diag.Diagnostics
}
```

**Usage**:
```go
// Save state
resp.State.Set(ctx, &data)

// Add error
resp.Diagnostics.AddError("Title", "Description")

// Add warning
resp.Diagnostics.AddWarning("Title", "Description")
```

#### resource.ReadResponse

```go
type ReadResponse struct {
    State       tfsdk.State
    Diagnostics diag.Diagnostics
}
```

**Special case**: If resource no longer exists:
```go
resp.State.RemoveResource(ctx)  // Removes from state
```

#### resource.UpdateResponse

```go
type UpdateResponse struct {
    State       tfsdk.State
    Diagnostics diag.Diagnostics
}
```

#### resource.DeleteResponse

```go
type DeleteResponse struct {
    State       tfsdk.State  // Automatically cleared by framework
    Diagnostics diag.Diagnostics
}
```

**Note**: Don't call `resp.State.Set()` in Delete - state is automatically removed.

### Diagnostics

```go
// Add error (stops execution)
resp.Diagnostics.AddError(
    "Summary",
    "Detailed error message",
)

// Add warning (continues execution)
resp.Diagnostics.AddWarning(
    "Summary",
    "Warning message",
)

// Check if errors exist
if resp.Diagnostics.HasError() {
    return
}

// Add multiple diagnostics
resp.Diagnostics.Append(otherDiagnostics...)
```

### Types Package

Terraform uses special types from `github.com/hashicorp/terraform-plugin-framework/types`:

```go
import "github.com/hashicorp/terraform-plugin-framework/types"

// String type
var title types.String
title = types.StringValue("Deploy")      // Set value
str := title.ValueString()               // Get Go string
isNull := title.IsNull()                 // Check if null
isUnknown := title.IsUnknown()           // Check if unknown

// Int64 type
var count types.Int64
count = types.Int64Value(42)
num := count.ValueInt64()

// Bool type
var enabled types.Bool
enabled = types.BoolValue(true)
b := enabled.ValueBool()

// Null values
title = types.StringNull()               // Explicitly null
title = types.StringUnknown()            // Unknown (during plan)
```

### Why Special Types?

Terraform needs to distinguish between:
- **Set**: User provided a value
- **Null**: User explicitly set to null
- **Unknown**: Value not yet known (during plan phase)

Go's native types can't represent all three states.

### Schema Attribute Types

```go
// String attribute
"title": schema.StringAttribute{
    Required: true,
}

// Int64 attribute
"count": schema.Int64Attribute{
    Optional: true,
}

// Bool attribute
"enabled": schema.BoolAttribute{
    Computed: true,
}

// List attribute
"tags": schema.ListAttribute{
    ElementType: types.StringType,
    Optional:    true,
}

// Map attribute
"labels": schema.MapAttribute{
    ElementType: types.StringType,
    Optional:    true,
}

// Object attribute
"config": schema.SingleNestedAttribute{
    Attributes: map[string]schema.Attribute{
        "host": schema.StringAttribute{
            Required: true,
        },
        "port": schema.Int64Attribute{
            Required: true,
        },
    },
    Optional: true,
}
```

### Plan Modifiers

Plan modifiers control how Terraform handles attribute changes during planning.

```go
"id": schema.StringAttribute{
    Computed: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
}
```

#### Common Plan Modifiers

| Modifier | Purpose | Use Case |
|----------|---------|----------|
| `UseStateForUnknown()` | Use state value if unknown | Computed IDs |
| `RequiresReplace()` | Force resource replacement | Immutable fields |
| `RequiresReplaceIfConfigured()` | Replace if user sets value | Optional immutable fields |

**Example - Immutable Field**:
```go
"region": schema.StringAttribute{
    Required: true,
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.RequiresReplace(),
    },
}
```

If user changes `region`, Terraform will destroy and recreate the resource.

### Validators

Validators ensure user input meets requirements.

```go
import "github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"

"priority": schema.StringAttribute{
    Optional: true,
    Validators: []validator.String{
        stringvalidator.OneOf("low", "medium", "high"),
    },
}

"due_date": schema.StringAttribute{
    Optional: true,
    Validators: []validator.String{
        stringvalidator.RegexMatches(
            regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`),
            "must be in YYYY-MM-DD format",
        ),
    },
}
```

#### Common Validators

| Validator | Purpose | Example |
|-----------|---------|---------|
| `OneOf()` | Value must be in list | `OneOf("low", "medium", "high")` |
| `LengthBetween()` | String length range | `LengthBetween(1, 100)` |
| `RegexMatches()` | Match regex pattern | `RegexMatches(regexp, "error msg")` |
| `AtLeastOneOf()` | At least one field set | `AtLeastOneOf(path.Expressions...)` |

---

## Common Patterns

### Error Handling

```go
// API call failed
task, err := r.client.CreateTask(...)
if err != nil {
    resp.Diagnostics.AddError(
        "Client Error",
        fmt.Sprintf("Unable to create task, got error: %s", err),
    )
    return
}

// Validation failed
if data.Priority.ValueString() != "" {
    priority := data.Priority.ValueString()
    if priority != "low" && priority != "medium" && priority != "high" {
        resp.Diagnostics.AddError(
            "Invalid Priority",
            fmt.Sprintf("Priority must be low, medium, or high, got: %s", priority),
        )
        return
    }
}

// Resource not found (in Read)
task, err := r.client.GetTask(id)
if err != nil {
    if strings.Contains(err.Error(), "not found") {
        resp.State.RemoveResource(ctx)
        return
    }
    resp.Diagnostics.AddError("Client Error", err.Error())
    return
}
```

### Default Values

```go
// In Configure method
host := data.Host.ValueString()
if host == "" {
    host = "http://localhost:8080"
}

// In Create method
priority := data.Priority.ValueString()
if priority == "" {
    priority = "medium"
}
```

### Handling Optional Fields

```go
// Check if field is set
if !data.Description.IsNull() {
    description = data.Description.ValueString()
}

// Or use empty string as default
description := data.Description.ValueString()  // Returns "" if null
```

### Time Formatting

```go
// API returns time.Time, convert to string for Terraform
data.CreatedAt = types.StringValue(task.CreatedAt.Format("2006-01-02T15:04:05Z07:00"))

// Parse string to time.Time
t, err := time.Parse("2006-01-02T15:04:05Z07:00", data.CreatedAt.ValueString())
```

### ID Conversion

```go
// String to int
var id int
_, err := fmt.Sscanf(data.ID.ValueString(), "%d", &id)

// Int to string
data.ID = types.StringValue(fmt.Sprintf("%d", task.ID))

// Alternative: strconv
id, err := strconv.Atoi(data.ID.ValueString())
data.ID = types.StringValue(strconv.Itoa(task.ID))
```

---

## Testing Your Provider

### Manual Testing

```bash
# 1. Start TaskMate API
cd taskmate
go run main.go

# 2. Build provider
cd taskmate-provider
go mod tidy
go build -o terraform-provider-taskmate

# 3. Install locally
mkdir -p ~/.terraform.d/plugins/hashitalk/taskmate/1.0.0/darwin_amd64
cp terraform-provider-taskmate ~/.terraform.d/plugins/hashitalk/taskmate/1.0.0/darwin_amd64/

# 4. Test with Terraform
cd examples/resources/taskmate_task
terraform init
terraform plan
terraform apply
terraform destroy
```

### Development Override

For faster iteration, use dev overrides:

**~/.terraformrc**:
```hcl
provider_installation {
  dev_overrides {
    "hashitalk/taskmate" = "/Users/swatityagi/hashitalk/taskmate-provider"
  }
  direct {}
}
```

Then just rebuild:
```bash
go build -o terraform-provider-taskmate
terraform plan  # Uses new binary immediately
```

### Debug Mode

```bash
# Start provider in debug mode
go run main.go --debug

# Output shows:
# TF_REATTACH_PROVIDERS='{"hashitalk/taskmate":{"Protocol":"grpc","ProtocolVersion":6,"Pid":12345,"Test":true,"Addr":{"Network":"unix","String":"/tmp/plugin123"}}}'

# In another terminal, use the environment variable
export TF_REATTACH_PROVIDERS='...'
terraform plan
```

Now you can attach a debugger (delve) to the provider process.

### Logging

```go
import "github.com/hashicorp/terraform-plugin-log/tflog"

func (r *TaskResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    tflog.Debug(ctx, "Creating task", map[string]interface{}{
        "title": data.Title.ValueString(),
    })
    
    task, err := r.client.CreateTask(...)
    
    tflog.Info(ctx, "Task created", map[string]interface{}{
        "id": task.ID,
    })
}
```

Enable logs:
```bash
export TF_LOG=DEBUG
terraform apply
```

---

## Summary

### Key Concepts

1. **Provider**: Entry point, configuration, client initialization
2. **Client**: HTTP communication, API abstraction
3. **Resource**: Full CRUD lifecycle management
4. **Data Source**: Read-only access to existing data
5. **State**: Terraform's record of managed infrastructure
6. **Schema**: Defines attributes and their properties
7. **Types**: Special Terraform types (String, Int64, Bool, etc.)
8. **Diagnostics**: Error and warning reporting

### Request Flow Summary

```
.tf file → Terraform Core → Provider Binary → Resource/DataSource → Client → API
                                                                              ↓
terraform.tfstate ← Terraform Core ← Provider Binary ← Resource/DataSource ← API
```

### Method Call Order

**Create**:
1. Provider.Configure()
2. Resource.Configure()
3. Resource.Create()
4. Resource.Read() (implicit refresh)

**Update**:
1. Resource.Read() (check current state)
2. Resource.Update()
3. Resource.Read() (implicit refresh)

**Delete**:
1. Resource.Read() (check current state)
2. Resource.Delete()

**Data Source**:
1. Provider.Configure()
2. DataSource.Configure()
3. DataSource.Read()

### Best Practices

1. **Always check diagnostics**: Return early if errors exist
2. **Use plan modifiers**: Prevent unnecessary replacements
3. **Handle 404 gracefully**: Remove from state in Read()
4. **Set all computed fields**: Even if unchanged
5. **Use meaningful error messages**: Include context and suggestions
6. **Validate input**: Use validators or manual checks
7. **Log important operations**: Use tflog for debugging
8. **Test thoroughly**: Manual testing + automated tests
9. **Follow naming conventions**: Use snake_case for attributes
10. **Document everything**: MarkdownDescription for all attributes

### Common Pitfalls

1. **Forgetting to set state**: Always call `resp.State.Set()`
2. **Not handling null values**: Check `IsNull()` before `ValueString()`
3. **Wrong attribute flags**: Computed fields can't be Required
4. **Modifying state in Delete**: State is auto-removed
5. **Not checking diagnostics**: Leads to nil pointer panics
6. **Hardcoded values**: Use provider configuration
7. **Poor error messages**: Include actionable information
8. **Not testing edge cases**: Empty strings, null values, API errors

---

## Next Steps

1. **Add Validation**: Implement validators for priority, due_date, status
2. **Add Tests**: Write acceptance tests using Terraform testing framework
3. **Enhance Client**: Add retry logic, better error handling
4. **Add Resources**: Implement more resource types (projects, users, etc.)
5. **Publish Provider**: Submit to Terraform Registry
6. **Add Documentation**: Generate docs using tfplugindocs

### Resources

- [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework)
- [Plugin Development](https://developer.hashicorp.com/terraform/plugin)
- [Provider Design Principles](https://developer.hashicorp.com/terraform/plugin/best-practices/hashicorp-provider-design-principles)
- [Testing Guide](https://developer.hashicorp.com/terraform/plugin/framework/acctests)

---

**Congratulations!** You now understand how Terraform providers work, from the entry point through API communication to state management. Use this knowledge to build robust, production-ready providers.
