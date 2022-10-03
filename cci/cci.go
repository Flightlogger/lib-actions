package cci

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"github.com/imroc/req/v3"
)

type Client struct {
  ProjectSlug string

  baseUrl string
  req *req.Client
}

func CreateClient(baseUrl string, apiKey string) (*Client) {
  return &Client {
    req: req.C().SetCommonBasicAuth(apiKey, ""),
    baseUrl: strings.Trim(baseUrl, "/"),
  }
}

type createPipelineResult struct { Number int `json:"number"` }

func (client Client) CreatePipeline(branch string) (int, error) {
  var result createPipelineResult
  resp, err := client.req.R().
    SetResult(&result).
    SetBodyJsonString(fmt.Sprintf(`{"branch": "%s"}`, branch)).
    Post(client.projectUrl("pipeline"))
  if err != nil { return -1, err }

  if !resp.IsSuccess() {
    return -1, errors.New(fmt.Sprintf("Failed to create pipeline with response: %s", readResponse(resp.Body)))
  }
  
  return result.Number, nil
}

type Pipeline struct {
  Id string `json:"id"`
  CreatedAt string `json:"created_at"`
  UpdatedAt string `json:"updated_at"`
  Number int `json:"number"`
  State string `json:"state"`
}
type getPipelinesResult struct { Items []Pipeline `json:"items"` }

// Fetches the pipelines of a given VCS branch.
func (client Client) GetBranchPipelines(branch string) ([]Pipeline, error) {
  var result getPipelinesResult
  resp, err := client.req.R().
    SetResult(&result).
    SetQueryParam("branch", branch).
    Get(client.projectUrl("pipeline"))
  if err != nil { return []Pipeline{}, errors.New("Failed to get pipelines.") }
  if !resp.IsSuccess() { return []Pipeline{}, errors.New(fmt.Sprintf("Failed to get pipeline with response: %s", readResponse(resp.Body))) }
  if len(result.Items) <= 0 { return result.Items, errors.New("No pipelines found.") }

  return result.Items, nil
}

type Workflow struct {
  Id string `json:"id"`
  Name string `json:"name"`
  Status string `json:"status"` // Can be one of: success, canceled, running, failed
}
type getWorkflowsResult struct { Items []Workflow `json:"items"` }

// Fetches the workflows of a given pipeline.
func (client Client) GetPipelineWorkflows(pipelineId string) ([]Workflow, error) {
  var result getWorkflowsResult
  resp, err := client.req.R().
    SetResult(&result).
    Get(client.url(fmt.Sprintf("pipeline/%s/workflow", pipelineId)))
  if err != nil { return []Workflow{}, errors.New("Failed to get workflows.") }
  if !resp.IsSuccess() { return []Workflow{}, errors.New(fmt.Sprintf("Failed to get workflows with response: %s", readResponse(resp.Body))) }
  if len(result.Items) <= 0 { return result.Items, errors.New("Pipeline has no workflows.") }

  return result.Items, nil
}

// Cancels a workflow.
func (client Client) CancelWorkflow(workflowId string) (error) {
  resp, err := client.req.R().Post(client.url(fmt.Sprintf("workflow/%s/cancel", workflowId)))
  if err != nil { return errors.New("Failed to cancel workflow.") }
  if !resp.IsSuccess() { return errors.New(fmt.Sprintf("Failed to cancel workflow with response: %s", readResponse(resp.Body))) }

  return nil
}

func (client Client) url(uri string) string {
  return client.baseUrl + "/" + uri
}

func (client Client) projectUrl(uri string) string {
  return client.url("project/" + client.ProjectSlug) + "/" + uri
}

// Reads entire buffer from io.ReadCloser rc. Returns default if failed for any reason.
func readResponse(rc io.ReadCloser) string {
  bytes, err := io.ReadAll(rc)
  if err != nil { return "" }

  return string(bytes)
}
