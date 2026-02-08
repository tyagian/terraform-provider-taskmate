package provider

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client handles API communication with TaskMate
type Client struct {
	Host   string
	Token  string
	client *http.Client
}

// Task represents a task from the API
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

// NewClient creates a new TaskMate API client
func NewClient(host, token string) *Client {
	return &Client{
		Host:  host,
		Token: token,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// makeRequest is a helper to make HTTP requests
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

	// Use X-API-Token header for authentication
	if c.Token != "" {
		req.Header.Set("X-API-Token", c.Token)
	}
	req.Header.Set("Content-Type", "application/json")

	return c.client.Do(req)
}

// CreateTask creates a new task
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

// GetTask retrieves a task by ID
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

// UpdateTask updates an existing task
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

// DeleteTask deletes a task by ID
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

// ListTasks retrieves all tasks
func (c *Client) ListTasks() ([]*Task, error) {
	resp, err := c.makeRequest("GET", "/tasks", nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error: %s (status: %d)", string(body), resp.StatusCode)
	}

	var tasks []*Task
	if err := json.NewDecoder(resp.Body).Decode(&tasks); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return tasks, nil
}
