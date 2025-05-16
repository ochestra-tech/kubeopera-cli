package providers

import (
	"fmt"

	"github.com/ochestra-tech/kubeforge-cli/pkg/config"
	"github.com/ochestra-tech/kubeforge-cli/pkg/ssh"
)

// AWSProvider implements the Provider interface for AWS
type AWSProvider struct {
	BaseProvider
}

// NewAWSProvider creates a new AWS provider
func NewAWSProvider(client *ssh.Client, cfg *config.Config) *AWSProvider {
	return &AWSProvider{
		BaseProvider: BaseProvider{
			Client: client,
			Config: cfg,
		},
	}
}

// GetMetadata retrieves AWS-specific metadata
func (p *AWSProvider) GetMetadata() (map[string]string, error) {
	metadata := make(map[string]string)

	// AWS metadata commands
	commands := []struct {
		key string
		cmd string
	}{
		{"instance-id", "curl -s http://169.254.169.254/latest/meta-data/instance-id"},
		{"instance-type", "curl -s http://169.254.169.254/latest/meta-data/instance-type"},
		{"availability-zone", "curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone"},
		{"region", "curl -s http://169.254.169.254/latest/meta-data/placement/availability-zone | sed 's/[a-z]$//'"},
		{"local-hostname", "curl -s http://169.254.169.254/latest/meta-data/local-hostname"},
		{"public-ipv4", "curl -s http://169.254.169.254/latest/meta-data/public-ipv4"},
	}

	// Execute each command and store the result
	for _, cmd := range commands {
		output, err := p.Client.RunCommand(cmd.cmd)
		if err == nil {
			metadata[cmd.key] = output
		} else {
			// Don't fail if one metadata command fails, just log it
			fmt.Printf("Warning: Failed to get AWS metadata '%s': %v\n", cmd.key, err)
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

// SetupCloudProvider configures the AWS cloud provider integration
func (p *AWSProvider) SetupCloudProvider() error {
	// Check if instance has IAM role with EC2 permissions
	checkIamCmd := "curl -s http://169.254.169.254/latest/meta-data/iam/security-credentials/"
	iamRole, err := p.Client.RunCommand(checkIamCmd)
	if err != nil || iamRole == "" {
		fmt.Println("Warning: No IAM role found for this instance. Cloud provider integration may not work correctly.")
		fmt.Println("         Please attach an IAM role with EC2 permissions to this instance.")
	}

	// AWS cloud provider commands
	commands := []string{
		// Create a minimal AWS cloud provider config
		"cat <<EOF | sudo tee /etc/kubernetes/cloud.conf\n[global]\nKubernetesClusterID=kubernetes\nEOF",
		"sudo chmod 600 /etc/kubernetes/cloud.conf",
		// Create secret from the cloud.conf file
		"kubectl -n kube-system create secret generic aws-cloud-provider --from-file=/etc/kubernetes/cloud.conf || true",
	}

	return p.Client.RunCommands(commands)
}

// GetCloudProviderOptions returns AWS cloud provider-specific options for kubeadm
func (p *AWSProvider) GetCloudProviderOptions() string {
	return "--cloud-provider=aws --cloud-config=/etc/kubernetes/cloud.conf"
}

// DisplayInfo shows AWS-specific information
func (p *AWSProvider) DisplayInfo() {
	fmt.Println("\n====== AWS Cloud Provider Information ======")
	fmt.Println("For AWS cloud provider integration:")
	fmt.Println("1. Ensure your EC2 instance has an IAM role with the following permissions:")
	fmt.Println("   - AmazonEC2FullAccess")
	fmt.Println("   - AmazonRoute53FullAccess (if using Route53 for DNS)")
	fmt.Println("2. Tag your AWS resources with the following tags:")
	fmt.Println("   - KubernetesCluster=<your-cluster-name>")
	fmt.Println("3. For load balancers, add the following tags to your subnets:")
	fmt.Println("   - kubernetes.io/cluster/<your-cluster-name>=shared")
	fmt.Println("4. For more information, visit:")
	fmt.Println("   https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/#aws")
	fmt.Println("================================================")
}
