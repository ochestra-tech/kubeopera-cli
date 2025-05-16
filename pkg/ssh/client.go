// SSH client implementation
package ssh

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/ochestra-tech/kubeforge-cli/pkg/config"
	"golang.org/x/crypto/ssh"
)

// Client represents an SSH client connection
type Client struct {
	config *config.Config
	client *ssh.Client
}

// NewClient creates a new SSH client using the provided configuration
func NewClient(cfg *config.Config) (*Client, error) {
	var authMethod ssh.AuthMethod

	if cfg.PrivateKey != "" {
		key, err := ioutil.ReadFile(cfg.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("unable to read private key: %v", err)
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, fmt.Errorf("unable to parse private key: %v", err)
		}
		authMethod = ssh.PublicKeys(signer)
	} else {
		authMethod = ssh.Password(cfg.Password)
	}

	sshConfig := &ssh.ClientConfig{
		User: cfg.User,
		Auth: []ssh.AuthMethod{
			authMethod,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         15 * time.Second,
	}

	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to host: %v", err)
	}

	return &Client{
		config: cfg,
		client: client,
	}, nil
}

// RunCommand executes a command on the remote host
func (c *Client) RunCommand(command string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	if err != nil {
		// Include stderr in the error message for better debugging
		errMsg := stderr.String()
		if errMsg != "" {
			return "", fmt.Errorf("command failed: %v\nError output: %s", err, errMsg)
		}
		return "", fmt.Errorf("command failed: %v", err)
	}

	return stdout.String(), nil
}

// RunCommandWithOutput executes a command and returns both stdout and stderr
func (c *Client) RunCommandWithOutput(command string) (string, string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	err = session.Run(command)
	return stdout.String(), stderr.String(), err
}

// RunCommands executes multiple commands sequentially
func (c *Client) RunCommands(commands []string) error {
	for _, cmd := range commands {
		fmt.Printf("  Running: %s\n", cmd)
		_, err := c.RunCommand(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

// UploadFile uploads a file to the remote host
func (c *Client) UploadFile(localPath, remotePath string) error {
	// Read local file
	content, err := ioutil.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %v", err)
	}

	// Create session for file transfer
	session, err := c.client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()

	// Create remote file
	cmd := fmt.Sprintf("cat > %s", remotePath)
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %v", err)
	}

	// Start the remote cat command
	if err := session.Start(cmd); err != nil {
		return fmt.Errorf("failed to start command: %v", err)
	}

	// Write file content
	_, err = stdin.Write(content)
	if err != nil {
		return fmt.Errorf("failed to write to stdin: %v", err)
	}

	// Close stdin to indicate EOF
	if err := stdin.Close(); err != nil {
		return fmt.Errorf("failed to close stdin: %v", err)
	}

	// Wait for command to complete
	if err := session.Wait(); err != nil {
		return fmt.Errorf("command failed: %v", err)
	}

	return nil
}

// CheckCommandExists checks if a command exists on the remote host
func (c *Client) CheckCommandExists(command string) bool {
	cmd := fmt.Sprintf("command -v %s", command)
	_, err := c.RunCommand(cmd)
	return err == nil
}

// GetRemoteHostname gets the hostname of the remote host
func (c *Client) GetRemoteHostname() (string, error) {
	output, err := c.RunCommand("hostname")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(output), nil
}

// Close closes the SSH client connection
func (c *Client) Close() error {
	return c.client.Close()
}
