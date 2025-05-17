package main

import (
	"flag"
	"fmt"
	"github.com/ochestra-tech/kubeforge-cli/pkg/config"
	"github.com/ochestra-tech/kubeforge-cli/pkg/installer"
	"github.com/ochestra-tech/kubeforge-cli/pkg/ssh"
	"log"
)

func main() {
	// Parse command line arguments
	host := flag.String("host", "", "Remote host IP address")
	port := flag.String("port", "22", "SSH port")
	user := flag.String("user", "", "SSH username")
	keyPath := flag.String("key", "", "Path to private key file")
	password := flag.String("password", "", "SSH password (if not using key)")
	provider := flag.String("provider", "aws", "Cloud provider: aws, gcp, azure, oracle")
	distribution := flag.String("distro", "", "Linux distribution: ubuntu, centos, amazon, oracle")

	flag.Parse()

	// Create configuration from flags
	cfg, err := config.NewConfig(*host, *port, *user, *keyPath, *password, *provider, *distribution)
	if err != nil {
		log.Fatalf("Failed to create configuration: %v", err)
	}

	// Display banner
	fmt.Println("==================================================")
	fmt.Println("  Kubernetes Cloud Installer")
	fmt.Println("  Cloud Provider:", cfg.Provider)
	fmt.Println("  Linux Distribution:", cfg.Distribution)
	fmt.Println("==================================================")

	// Create SSH client
	sshClient, err := ssh.NewClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create SSH client: %v", err)
	}
	defer sshClient.Close()

	fmt.Println("Connected to remote host successfully")

	// Create installer
	k8sInstaller := installer.NewInstaller(sshClient, cfg)

	// Run installation steps
	steps := []struct {
		name string
		fn   func() error
	}{
		{"Installing prerequisites", k8sInstaller.InstallPrerequisites},
		{"Installing container runtime", k8sInstaller.InstallContainerRuntime},
		{"Installing Kubernetes components", k8sInstaller.InstallKubernetesComponents},
		{"Initializing Kubernetes cluster", k8sInstaller.InitializeCluster},
		{"Configuring cloud provider integration", k8sInstaller.SetupCloudProviderIntegration},
	}

	for _, step := range steps {
		fmt.Printf("\n[*] %s...\n", step.name)
		if err := step.fn(); err != nil {
			log.Fatalf("Failed to %s: %v", step.name, err)
		}
		fmt.Printf("[✓] %s completed successfully\n", step.name)
	}

	// Display cloud provider-specific information
	k8sInstaller.DisplayCloudProviderInfo()

	fmt.Println("\n[✓] Kubernetes installation completed successfully!")
	fmt.Println("\nTo access your Kubernetes cluster:")
	fmt.Println("  1. SSH into your VM:    ssh -i", *keyPath, *user+"@"+*host)
	fmt.Println("  2. Check nodes status:  kubectl get nodes")
	fmt.Println("  3. Deploy an application example: kubectl create deployment nginx --image=nginx")
	fmt.Println("  4. Expose the deployment: kubectl expose deployment nginx --port=80 --type=NodePort")
	fmt.Println("\nThank you for using Kubernetes Cloud Installer!")
}
