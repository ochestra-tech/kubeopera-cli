// Package installer Main installation logic for the KubeForge CLI
package installer

import (
	"fmt"

	"github.com/ochestra-tech/kubeforge-cli/pkg/config"
	"github.com/ochestra-tech/kubeforge-cli/pkg/providers"
	"github.com/ochestra-tech/kubeforge-cli/pkg/ssh"
)

// Installer manages the Kubernetes installation process
type Installer struct {
	Client   *ssh.Client
	Config   *config.Config
	Provider providers.Provider
}

// NewInstaller creates a new installer
func NewInstaller(client *ssh.Client, cfg *config.Config) *Installer {
	provider := providers.NewProvider(client, cfg)

	return &Installer{
		Client:   client,
		Config:   cfg,
		Provider: provider,
	}
}

// InstallPrerequisites installs required dependencies based on cloud provider and distribution
func (i *Installer) InstallPrerequisites() error {
	pm := i.Config.GetPackageManager()

	// Common prerequisites for all distributions
	commonCommands := []string{
		"sudo swapoff -a",
		"sudo sed -i '/swap/d' /etc/fstab",
		"sudo modprobe overlay",
		"sudo modprobe br_netfilter",
		"echo '1' | sudo tee /proc/sys/net/ipv4/ip_forward",
		"echo '1' | sudo tee /proc/sys/net/bridge/bridge-nf-call-iptables",
		"echo '1' | sudo tee /proc/sys/net/bridge/bridge-nf-call-ip6tables",
		"cat <<EOF | sudo tee /etc/modules-load.d/k8s.conf\noverlay\nbr_netfilter\nEOF",
		"cat <<EOF | sudo tee /etc/sysctl.d/k8s.conf\nnet.bridge.bridge-nf-call-iptables = 1\nnet.bridge.bridge-nf-call-ip6tables = 1\nnet.ipv4.ip_forward = 1\nEOF",
		"sudo sysctl --system",
	}

	// Distribution-specific commands
	var distroCommands []string

	if i.Config.IsDebianBased() {
		distroCommands = []string{
			pm["update"],
			pm["install"] + " apt-transport-https ca-certificates curl software-properties-common gnupg lsb-release",
		}
	} else if i.Config.IsRHELBased() {
		distroCommands = []string{
			"sudo setenforce 0 || true",
			"sudo sed -i 's/^SELINUX=enforcing$/SELINUX=permissive/' /etc/selinux/config || true",
			pm["update"],
			pm["install"] + " curl wget socat conntrack ebtables ipset",
		}
	}

	// Cloud provider-specific commands
	var providerCommands []string

	switch i.Config.Provider {
	case config.AWS:
		// AWS-specific optimizations
		providerCommands = []string{
			"sudo hostnamectl set-hostname $(curl -s http://169.254.169.254/latest/meta-data/local-hostname) || true",
		}
	case config.GCP:
		// GCP-specific optimizations
		providerCommands = []string{
			"sudo hostnamectl set-hostname $(curl -s -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/hostname | cut -d. -f1) || true",
		}
	case config.Azure:
		// Azure-specific optimizations
		providerCommands = []string{
			"sudo hostnamectl set-hostname $(curl -s -H Metadata:true 'http://169.254.169.254/metadata/instance/compute/name?api-version=2019-06-01&format=text') || true",
		}
	case config.Oracle:
		// Oracle-specific optimizations
		providerCommands = []string{
			// Oracle Cloud doesn't have a standard metadata service like other providers
			// so we'll just use the hostname command
			"sudo hostnamectl set-hostname $(hostname) || true",
		}
	}

	// Combine all commands
	commands := append(commonCommands, distroCommands...)
	commands = append(commands, providerCommands...)

	return i.Client.RunCommands(commands)
}

// InstallContainerRuntime installs and configures containerd based on distribution
func (i *Installer) InstallContainerRuntime() error {
	pm := i.Config.GetPackageManager()

	var commands []string

	if i.Config.IsDebianBased() {
		commands = []string{
			// Install containerd
			"sudo mkdir -p /etc/apt/keyrings",
			"curl -fsSL https://download.docker.com/linux/" + i.Config.Distribution + "/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg",
			"echo \"deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/" + i.Config.Distribution + " $(lsb_release -cs) stable\" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null",
			pm["update"],
			pm["install"] + " containerd.io",
		}
	} else if i.Config.IsRHELBased() {
		commands = []string{
			pm["install"] + " yum-utils device-mapper-persistent-data lvm2",
			pm["repository"] + " https://download.docker.com/linux/centos/docker-ce.repo",
			pm["install"] + " containerd.io",
		}
	}

	// Common configuration for all distributions
	commonCommands := []string{
		"sudo mkdir -p /etc/containerd",
		"sudo containerd config default | sudo tee /etc/containerd/config.toml",
		"sudo sed -i 's/SystemdCgroup = false/SystemdCgroup = true/g' /etc/containerd/config.toml",
		"sudo systemctl restart containerd",
		"sudo systemctl enable containerd",
	}

	commands = append(commands, commonCommands...)

	return i.Client.RunCommands(commands)
}

// InstallKubernetesComponents installs kubeadm, kubelet, and kubectl based on distribution
func (i *Installer) InstallKubernetesComponents() error {
	pm := i.Config.GetPackageManager()

	var commands []string

	if i.Config.IsDebianBased() {
		commands = []string{
			"sudo curl -fsSLo /etc/apt/keyrings/kubernetes-archive-keyring.gpg https://packages.cloud.google.com/apt/doc/apt-key.gpg",
			"echo \"deb [signed-by=/etc/apt/keyrings/kubernetes-archive-keyring.gpg] https://apt.kubernetes.io/ kubernetes-xenial main\" | sudo tee /etc/apt/sources.list.d/kubernetes.list",
			pm["update"],
			pm["install"] + " kubelet kubeadm kubectl",
			"sudo apt-mark hold kubelet kubeadm kubectl",
		}
	} else if i.Config.IsRHELBased() {
		commands = []string{
			"cat <<EOF | sudo tee /etc/yum.repos.d/kubernetes.repo\n[kubernetes]\nname=Kubernetes\nbaseurl=https://packages.cloud.google.com/yum/repos/kubernetes-el7-\\$basearch\nenabled=1\ngpgcheck=1\nrepo_gpgcheck=1\ngpgkey=https://packages.cloud.google.com/yum/doc/yum-key.gpg https://packages.cloud.google.com/yum/doc/rpm-package-key.gpg\nEOF",
			pm["install"] + " kubelet kubeadm kubectl --disableexcludes=kubernetes",
		}
	}

	// Common configuration for all distributions
	commonCommands := []string{
		"sudo systemctl enable --now kubelet",
	}

	commands = append(commands, commonCommands...)

	return i.Client.RunCommands(commands)
}

// InitializeCluster initializes the Kubernetes cluster with cloud provider-specific settings
func (i *Installer) InitializeCluster() error {
	// Get cloud provider-specific configuration
	cloudConfig := i.Provider.GetCloudProviderOptions()

	// Initialize Kubernetes cluster with cloud provider if specified
	initCmd := "sudo kubeadm init --pod-network-cidr=10.244.0.0/16"
	if cloudConfig != "" {
		initCmd += " " + cloudConfig
	}

	fmt.Printf("  Running: %s\n", initCmd)
	_, err := i.Client.RunCommand(initCmd)
	if err != nil {
		return err
	}

	// Configure kubectl for the user
	commands := []string{
		"mkdir -p $HOME/.kube",
		"sudo cp -i /etc/kubernetes/admin.conf $HOME/.kube/config",
		"sudo chown $(id -u):$(id -g) $HOME/.kube/config",
		// Install Flannel CNI
		"kubectl apply -f https://raw.githubusercontent.com/flannel-io/flannel/master/Documentation/kube-flannel.yml",
		// Allow pods to run on the master node (optional, remove for production)
		"kubectl taint nodes --all node-role.kubernetes.io/control-plane-",
	}

	err = i.Client.RunCommands(commands)
	if err != nil {
		return err
	}

	// Extract the join command for other nodes (if needed)
	joinCmd, err := i.Client.RunCommand("sudo kubeadm token create --print-join-command")
	if err != nil {
		fmt.Printf("Warning: Could not create join command: %v\n", err)
	} else {
		fmt.Printf("\nUse the following command to join other nodes to the cluster:\n%s\n", joinCmd)
	}

	return nil
}

// SetupCloudProviderIntegration configures the cloud provider integration
func (i *Installer) SetupCloudProviderIntegration() error {
	return i.Provider.SetupCloudProvider()
}

// DisplayCloudProviderInfo shows cloud provider-specific information
func (i *Installer) DisplayCloudProviderInfo() {
	i.Provider.DisplayInfo()
}
