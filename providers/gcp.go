// GCP provider for kubeforge CLI
package providers

import (
	"fmt"
	"strings"

	"github.com/ochestra-tech/kubeforge-cli/pkg/config"
	"github.com/ochestra-tech/kubeforge-cli/pkg/ssh"
)

// GCPProvider implements the Provider interface for GCP
type GCPProvider struct {
	BaseProvider
}

// NewGCPProvider creates a new GCP provider
func NewGCPProvider(client *ssh.Client, cfg *config.Config) *GCPProvider {
	return &GCPProvider{
		BaseProvider: BaseProvider{
			Client: client,
			Config: cfg,
		},
	}
}

// GetMetadata retrieves GCP-specific metadata
func (p *GCPProvider) GetMetadata() (map[string]string, error) {
	metadata := make(map[string]string)

	// GCP metadata commands
	commands := []struct {
		key string
		cmd string
	}{
		{"instance-id", "curl -s -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/id"},
		{"instance-name", "curl -s -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/name"},
		{"zone", "curl -s -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/zone | cut -d/ -f4"},
		{"machine-type", "curl -s -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/machine-type | cut -d/ -f4"},
		{"project-id", "curl -s -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/project/project-id"},
		{"external-ip", "curl -s -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip"},
	}

	// Execute each command and store the result
	for _, cmd := range commands {
		output, err := p.Client.RunCommand(cmd.cmd)
		if err == nil {
			metadata[cmd.key] = output
		} else {
			// Don't fail if one metadata command fails, just log it
			fmt.Printf("Warning: Failed to get GCP metadata '%s': %v\n", cmd.key, err)
		}
	}

	// Add common metadata
	commonCommands := baseMetadataCommands()
	for _, cmd := range commonCommands {
		output, err := p.Client.RunCommand(cmd)
		if err == nil {
			metadata[cmd] = output
		}
	}

	return metadata, nil
}

// SetupCloudProvider configures the GCP cloud provider integration
func (p *GCPProvider) SetupCloudProvider() error {
	// Check if VM has the required service account scopes
	checkScopesCmd := "curl -s -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/scopes"
	scopes, err := p.Client.RunCommand(checkScopesCmd)
	if err != nil {
		fmt.Println("Warning: Unable to verify service account scopes. Cloud provider integration may not work correctly.")
	} else {
		if !strings.Contains(scopes, "https://www.googleapis.com/auth/compute") {
			fmt.Println("Warning: VM service account may not have compute scope. Cloud provider integration may not work correctly.")
			fmt.Println("         Ensure the VM's service account has the compute.networkUser role.")
		}
	}

	// Get project ID
	projectIDCmd := "curl -s -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/project/project-id"
	projectID, err := p.Client.RunCommand(projectIDCmd)
	if err != nil {
		return fmt.Errorf("failed to get GCP project ID: %v", err)
	}

	// GCP cloud provider commands
	commands := []string{
		// Create a minimal GCP cloud provider config
		fmt.Sprintf("cat <<EOF | sudo tee /etc/kubernetes/cloud.conf\n[global]\nproject-id = %s\nnode-tags = k8s-node\nnode-instance-prefix = k8s\nEOF", strings.TrimSpace(projectID)),
		"sudo chmod 600 /etc/kubernetes/cloud.conf",
		// Create secret from the cloud.conf file
		"kubectl -n kube-system create secret generic gcp-cloud-provider --from-file=/etc/kubernetes/cloud.conf || true",
	}

	return p.Client.RunCommands(commands)
}

// GetCloudProviderOptions returns GCP cloud provider-specific options for kubeadm
func (p *GCPProvider) GetCloudProviderOptions() string {
	return "--cloud-provider=gce --cloud-config=/etc/kubernetes/cloud.conf"
}

// DisplayInfo shows GCP-specific information
func (p *GCPProvider) DisplayInfo() {
	fmt.Println("\n====== GCP Cloud Provider Information ======")
	fmt.Println("For GCP cloud provider integration:")
	fmt.Println("1. Ensure your VM instance has the following OAuth scopes:")
	fmt.Println("   - compute-rw")
	fmt.Println("   - storage-ro")
	fmt.Println("2. The service account associated with the VM should have:")
	fmt.Println("   - Compute Admin role")
	fmt.Println("   - Network Admin role")
	fmt.Println("3. For load balancers, ensure your network is properly configured with:")
	fmt.Println("   - Proper firewall rules for health checks (TCP:10256)")
	fmt.Println("4. For more information, visit:")
	fmt.Println("   https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/#gce")
	fmt.Println("===============================================")
}
