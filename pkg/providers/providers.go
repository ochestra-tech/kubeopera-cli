package providers

import (
	"github.com/ochestra-tech/kubeforge-cli/pkg/config"
	"github.com/ochestra-tech/kubeforge-cli/pkg/ssh"
)

// Provider defines the interface for cloud provider-specific operations
type Provider interface {
	// GetMetadata retrieves cloud provider-specific metadata
	GetMetadata() (map[string]string, error)

	// SetupCloudProvider configures the Kubernetes cloud provider integration
	SetupCloudProvider() error

	// GetCloudProviderOptions returns cloud provider-specific options for kubeadm
	GetCloudProviderOptions() string

	// DisplayInfo shows cloud provider-specific information
	DisplayInfo()
}

// NewProvider creates a new cloud provider based on the configuration
func NewProvider(client *ssh.Client, cfg *config.Config) Provider {
	switch cfg.Provider {
	case config.AWS:
		return NewAWSProvider(client, cfg)
	case config.GCP:
		return NewGCPProvider(client, cfg)
	case config.Azure:
		return NewAzureProvider(client, cfg)
	case config.Oracle:
		return NewOracleProvider(client, cfg)
	default:
		// Default to AWS
		return NewAWSProvider(client, cfg)
	}
}

// BaseProvider implements common functionality for all providers
type BaseProvider struct {
	Client *ssh.Client
	Config *config.Config
}

// baseMetadataCommands returns common commands for all providers
func baseMetadataCommands() []string {
	return []string{
		"hostname",
		"uname -a",
		"lscpu | grep '^CPU(s):'",
		"free -m | grep '^Mem:'",
	}
}
