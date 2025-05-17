// Oracle provider implementation for the kubeforge CLI
import (
	"fmt"

	"github.com/yourusername/k8s-cloud-installer/pkg/config"
	"github.com/yourusername/k8s-cloud-installer/pkg/ssh"
)

// OracleProvider implements the Provider interface for Oracle Cloud
type OracleProvider struct {
	BaseProvider
}

// NewOracleProvider creates a new Oracle provider
func NewOracleProvider(client *ssh.Client, cfg *config.Config) *OracleProvider {
	return &OracleProvider{
		BaseProvider: BaseProvider{
			Client: client,
			Config: cfg,
		},
	}
}

// GetMetadata retrieves Oracle Cloud-specific metadata
func (p *OracleProvider) GetMetadata() (map[string]string, error) {
	metadata := make(map[string]string)

	// Oracle Cloud doesn't have a standard metadata service like other providers
	// So we use basic system commands to gather information
	commands := []struct {
		key string
		cmd string
	}{
		{"hostname", "hostname"},
		{"ip-addr", "ip addr show | grep 'inet ' | grep -v '127.0.0.1' | awk '{print $2}' | cut -d/ -f1"},
		{"os-version", "cat /etc/os-release | grep PRETTY_NAME | cut -d= -f2 | tr -d '\"'"},
	}

	// Execute each command and store the result
	for _, cmd := range commands {
		output, err := p.Client.RunCommand(cmd.cmd)
		if err == nil {
			metadata[cmd.key] = output
		} else {
			// Don't fail if one metadata command fails, just log it
			fmt.Printf("Warning: Failed to get Oracle Cloud metadata '%s': %v\n", cmd.key, err)
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

// SetupCloudProvider configures the Oracle Cloud provider integration
func (p *OracleProvider) SetupCloudProvider() error {
	// Oracle Cloud doesn't have a native Kubernetes cloud provider
	// So we just display information about the Oracle Cloud Controller Manager
	fmt.Println("Oracle Cloud doesn't have a native Kubernetes cloud provider integration.")
	fmt.Println("For load balancer and volume provisioning support, please install the Oracle Cloud Controller Manager separately.")
	fmt.Println("See: https://github.com/oracle/oci-cloud-controller-manager")

	// We'll just create a placeholder config file
	commands := []string{
		// Create a placeholder cloud provider config
		"cat <<EOF | sudo tee /etc/kubernetes/oci.conf\n# Oracle Cloud configuration\n# See https://github.com/oracle/oci-cloud-controller-manager for more information\nEOF",
		"sudo chmod 600 /etc/kubernetes/oci.conf",
	}

	return p.Client.RunCommands(commands)
}

// GetCloudProviderOptions returns Oracle Cloud provider-specific options for kubeadm
// Oracle Cloud doesn't have a native cloud provider, so this is empty
func (p *OracleProvider) GetCloudProviderOptions() string {
	return ""
}

// DisplayInfo shows Oracle Cloud-specific information
func (p *OracleProvider) DisplayInfo() {
	fmt.Println("\n====== Oracle Cloud Information ======")
	fmt.Println("For Oracle Cloud integration:")
	fmt.Println("1. Oracle Cloud doesn't have a native Kubernetes cloud provider.")
	fmt.Println("2. For load balancer and volume provisioning support:")
	fmt.Println("   - Install the Oracle Cloud Controller Manager from: https://github.com/oracle/oci-cloud-controller-manager")
	fmt.Println("   - Follow the instructions to create a configuration file with the required OCI credentials")
	fmt.Println("3. To set up cloud storage:")
	fmt.Println("   - Install the Oracle Cloud Storage Provisioner: https://github.com/oracle/oci-cloud-controller-manager/blob/master/docs/volume-provisioner.md")
	fmt.Println("4. For networking, ensure your security lists allow:")
	fmt.Println("   - Pod-to-Pod communication")
	fmt.Println("   - NodePort services (30000-32767)")
	fmt.Println("   - Control plane communication (6443, 10250-10252)")
	fmt.Println("===========================================")
}
