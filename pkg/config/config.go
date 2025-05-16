// Config structs and functions
package config

import (
	"fmt"
)

// CloudProvider represents the type of cloud provider
type CloudProvider string

const (
	AWS    CloudProvider = "aws"
	GCP    CloudProvider = "gcp"
	Azure  CloudProvider = "azure"
	Oracle CloudProvider = "oracle"
)

// Config stores the connection and installation configuration
type Config struct {
	Host         string
	Port         string
	User         string
	PrivateKey   string
	Password     string
	Provider     CloudProvider
	Distribution string
}

// NewConfig creates a new configuration with validation and defaults
func NewConfig(host, port, user, keyPath, password, provider, distribution string) (*Config, error) {
	if host == "" {
		return nil, fmt.Errorf("host IP address is required")
	}

	if keyPath == "" && password == "" {
		return nil, fmt.Errorf("either private key or password is required")
	}

	// Validate cloud provider
	cloudProvider := CloudProvider(provider)
	if !isValidProvider(cloudProvider) {
		return nil, fmt.Errorf("invalid cloud provider '%s': use aws, gcp, azure, or oracle", provider)
	}

	// Set default user based on provider if not specified
	username := user
	if username == "" {
		username = getDefaultUser(cloudProvider, distribution)
	}

	// Set default distribution if not specified
	distro := distribution
	if distro == "" {
		distro = getDefaultDistribution(cloudProvider)
	}

	return &Config{
		Host:         host,
		Port:         port,
		User:         username,
		PrivateKey:   keyPath,
		Password:     password,
		Provider:     cloudProvider,
		Distribution: distro,
	}, nil
}

// isValidProvider checks if the provided cloud provider is valid
func isValidProvider(provider CloudProvider) bool {
	return provider == AWS || provider == GCP || provider == Azure || provider == Oracle
}

// getDefaultUser returns the default SSH user for the given cloud provider and distribution
func getDefaultUser(provider CloudProvider, distribution string) string {
	switch provider {
	case AWS:
		if distribution == "ubuntu" {
			return "ubuntu"
		}
		return "ec2-user"
	case GCP:
		if distribution == "ubuntu" {
			return "ubuntu"
		}
		return "google_user"
	case Azure:
		if distribution == "ubuntu" {
			return "azureuser"
		}
		return "adminuser"
	case Oracle:
		if distribution == "ubuntu" {
			return "ubuntu"
		}
		return "opc"
	default:
		return "root"
	}
}

// getDefaultDistribution returns the default distribution for the given cloud provider
func getDefaultDistribution(provider CloudProvider) string {
	switch provider {
	case AWS:
		return "amazon"
	case GCP:
		return "debian"
	case Azure:
		return "ubuntu"
	case Oracle:
		return "oracle"
	default:
		return "ubuntu"
	}
}

// IsDebianBased returns true if the distribution is Debian-based
func (c *Config) IsDebianBased() bool {
	return c.Distribution == "ubuntu" || c.Distribution == "debian"
}

// IsRHELBased returns true if the distribution is RHEL-based
func (c *Config) IsRHELBased() bool {
	return c.Distribution == "centos" || c.Distribution == "rhel" ||
		c.Distribution == "amazon" || c.Distribution == "oracle"
}

// GetPackageManager returns the appropriate package manager commands for the distribution
func (c *Config) GetPackageManager() map[string]string {
	switch {
	case c.IsDebianBased():
		return map[string]string{
			"update":     "sudo apt-get update",
			"install":    "sudo apt-get install -y",
			"repository": "sudo apt-add-repository",
		}
	case c.IsRHELBased():
		return map[string]string{
			"update":     "sudo yum update -y",
			"install":    "sudo yum install -y",
			"repository": "sudo yum-config-manager --add-repo",
		}
	default:
		// Default to Ubuntu
		return map[string]string{
			"update":     "sudo apt-get update",
			"install":    "sudo apt-get install -y",
			"repository": "sudo apt-add-repository",
		}
	}
}
