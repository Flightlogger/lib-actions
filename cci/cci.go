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

type getPipelinesResult struct {
  Items []struct {
    Id string `json:"id"`
  } `json:"items"`
}

type getWorkflowsResult struct {
  Items []struct {
    Id string `json:"id"`
    Name string `json:"name"`
    Status string `json:"status"`
  } `json:"items"`
}

// Cancels all workflows that match `workflowName` for the latest pipeline of the `ref` reference (git).
// Returns the number of workflows cancelled and possible error. On error, int return value is -1.
// NOTE(jpp): Will currently error if no pipelines are found, or if pipeline has no workflows. This isn't technically a "failure", but I opted to keep it that way for now to identify possible bugs. When more stable, we can remove.
func (client Client) CancelLastPipelineWorkflows(ref string, workflowName string) (int, error) {
  // Get the pipelines of the branch.
  var pipelinesResult getPipelinesResult
  resp, err := client.req.R().
    SetResult(&pipelinesResult).
    SetQueryParam("branch", ref).
    Get(client.projectUrl("pipeline"))
  if err != nil { return -1, errors.New("Failed to get pipeline.") }
  if !resp.IsSuccess() { return -1, errors.New(fmt.Sprintf("Failed to get pipeline with response: %s", readResponse(resp.Body))) }
  if len(pipelinesResult.Items) <= 0 { return 0, errors.New("No pipelines found.") }

  // Get workflows of the first found pipeline (should be the latest)
  item := pipelinesResult.Items[0]
  var workflowsResult getWorkflowsResult
  resp, err = client.req.R().
    SetResult(&workflowsResult).
    Get(client.url(fmt.Sprintf("pipeline/%s/workflow", item.Id)))
  if err != nil { return -1, errors.New("Failed to get workflows.") }
  if !resp.IsSuccess() { return -1, errors.New(fmt.Sprintf("Failed to get workflows with response: %s", readResponse(resp.Body))) }
  if len(workflowsResult.Items) <= 0 { return -1, errors.New("Pipeline has no workflows.") }

  // Cancel all found workflows that are running and match the given name.
  cancellations := 0
  for _, workflow := range workflowsResult.Items {
    if workflow.Name == workflowName && workflow.Status == "running" {
      resp, err = client.req.R().Post(client.url(fmt.Sprintf("workflow/%s/cancel", workflow.Id)))
      if err != nil { return -1, errors.New("Failed to cancel workflow.") }
      if !resp.IsSuccess() { return -1, errors.New(fmt.Sprintf("Failed to cancel workflow with response: %s", readResponse(resp.Body))) }
      cancellations++
    }
  }

  return cancellations, nil
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
