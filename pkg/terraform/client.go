package terraform

import (
	"context"
	"os/exec"
)

// Client represents a Terraform client
type Client struct {
	BinaryPath string
	WorkingDir string
}

// NewClient creates a new Terraform client
func NewClient(binaryPath, workingDir string) *Client {
	return &Client{
		BinaryPath: binaryPath,
		WorkingDir: workingDir,
	}
}

// Init runs terraform init
func (c *Client) Init(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.BinaryPath, "init")
	cmd.Dir = c.WorkingDir
	return cmd.Run()
}

// Plan runs terraform plan
func (c *Client) Plan(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.BinaryPath, "plan")
	cmd.Dir = c.WorkingDir
	return cmd.Run()
}

// Apply runs terraform apply
func (c *Client) Apply(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.BinaryPath, "apply", "-auto-approve")
	cmd.Dir = c.WorkingDir
	return cmd.Run()
}

// Destroy runs terraform destroy
func (c *Client) Destroy(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, c.BinaryPath, "destroy", "-auto-approve")
	cmd.Dir = c.WorkingDir
	return cmd.Run()
}
