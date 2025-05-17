package providers

import (
	"fmt"
	"strings"

	"github.com/ochestra-tech/kubeforge-cli/pkg/config"
	"github.com/ochestra-tech/kubeforge-cli/pkg/ssh"
)

// AzureProvider implements the Provider interface for Azure
type AzureProvider struct {
	BaseProvider
}

// NewAzureProvider creates a new Azure provider
func NewAzureProvider(client *ssh.Client, cfg *config.Config) *AzureProvider {
	return &AzureProvider{
		BaseProvider: BaseProvider{
			Client: client,
			Config: cfg,
		},
	}
}

// GetMetadata retrieves Azure-specific metadata
func (p *AzureProvider) GetMetadata() (map[string]string, error) {
	metadata := make(map[string]string)

	// Azure metadata commands
	commands := []struct {
		key string
		cmd string
	}{
		{"vm-name", "curl -s -H Metadata:true 'http://169.254.169.254/metadata/instance/compute/name?api-version=2019-06-01&format=text'"},
		{"resource-group", "curl -s -H Metadata:true 'http://169.254.169.254/metadata/instance/compute/resourceGroupName?api-version=2019-06-01&format=text'"},
		{"subscription-id", "curl -s -H Metadata:true 'http://169.254.169.254/metadata/instance/compute/subscriptionId?api-version=2019-06-01&format=text'"},
		{"location", "curl -s -H Metadata:true 'http://169.254.169.254/metadata/instance/compute/location?api-version=2019-06-01&format=text'"},
		{"vm-size", "curl -s -H Metadata:true 'http://169.254.169.254/metadata/instance/compute/vmSize?api-version=2019-06-01&format=text'"},
		{"public-ipv4", "curl -s -H Metadata:true 'http://169.254.169.254/metadata/instance/network/interface/0/ipv4/ipAddress/0/publicIpAddress?api-version=2019-06-01&format=text'"},
	}

	// Execute each command and store the result
	for _, cmd := range commands {
		output, err := p.Client.RunCommand(cmd.cmd)
		if err == nil {
			metadata[cmd.key] = output
		} else {
			// Don't fail if one metadata command fails, just log it
			fmt.Printf("Warning: Failed to get Azure metadata '%s': %v\n", cmd.key, err)
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

// SetupCloudProvider configures the Azure cloud provider integration
func (p *AzureProvider) SetupCloudProvider() error {
	// Get the required metadata for cloud.conf
	metadata, err := p.GetMetadata()
	if err != nil {
		return fmt.Errorf("failed to get Azure metadata: %v", err)
	}

	subscriptionID := strings.TrimSpace(metadata["subscription-id"])
	resourceGroup := strings.TrimSpace(metadata["resource-group"])
	location := strings.TrimSpace(metadata["location"])

	if subscriptionID == "" || resourceGroup == "" || location == "" {
		fmt.Println("Warning: Azure metadata incomplete. Cloud provider integration may not work correctly.")
	}

	// Azure cloud provider commands
	commands := []string{
		// Create the Azure cloud provider config file
		fmt.Sprintf(`cat <<EOF | sudo tee /etc/kubernetes/azure.json
{
  "cloud": "AzurePublicCloud",
  "tenantId": "",
  "subscriptionId": "%s",
  "resourceGroup": "%s",
  "location": "%s",
  "useManagedIdentityExtension": true,
  "useInstanceMetadata": true
}
EOF`, subscriptionID, resourceGroup, location),
		"sudo chmod 600 /etc/kubernetes/azure.json",
		// Create secret from the azure.json file
		"kubectl -n kube-system create secret generic azure-cloud-provider --from-file=/etc/kubernetes/azure.json || true",
	}

	return p.Client.RunCommands(commands)
}

// GetCloudProviderOptions returns Azure cloud provider-specific options for kubeadm
func (p *AzureProvider) GetCloudProviderOptions() string {
	return "--cloud-provider=azure --cloud-config=/etc/kubernetes/azure.json"
}

// DisplayInfo shows Azure-specific information
func (p *AzureProvider) DisplayInfo() {
	fmt.Println("\n====== Azure Cloud Provider Information ======")
	fmt.Println("For Azure cloud provider integration:")
	fmt.Println("1. Ensure your VM has a Managed Identity with:")
	fmt.Println("   - Contributor role on the resource group")
	fmt.Println("   - Network Contributor role (for load balancer configuration)")
	fmt.Println("2. For load balancers, ensure your network is properly configured with:")
	fmt.Println("   - Network security group allowing health probe traffic")
	fmt.Println("   - Firewall rules allowing port 10256 for health checks")
	fmt.Println("3. For multi-node clusters, all VMs should be in the same resource group")
	fmt.Println("4. For more information, visit:")
	fmt.Println("   https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/#azure")
	fmt.Println("================================================")
}
