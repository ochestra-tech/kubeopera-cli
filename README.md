# kubeforge-cli

Installs Kubernetes on a remote VM from any of the major cloud providers

KubeForge is a Go-based command-line tool for automated Kubernetes installation on virtual machines across multiple cloud providers (AWS, GCP, Azure, and Oracle Cloud).

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

## Features

- **Multi-cloud support**: Install Kubernetes on AWS, GCP, Azure, and Oracle Cloud VMs
- **Distribution-aware**: Compatible with Ubuntu, Debian, CentOS, Amazon Linux, and Oracle Linux
- **Cloud provider integration**: Configures cloud-specific settings automatically
- **Secure**: Uses SSH for all operations with key-based or password authentication
- **Flexible**: Customizable for different Linux distributions and installation requirements
- **Single-node cluster**: Sets up a working Kubernetes cluster with a single command

## Installation

### Prerequisites

- Go 1.20 or later
- SSH access to a virtual machine on one of the supported cloud providers

### Building from source

```bash
# Clone the repository
git clone https://github.com/ochestra-tech/k8s-cloud-installer.git
cd k8s-cloud-installer

# Build the binary
go build -o kubeforge-cli cmd/installer/main.go

# Make it executable
chmod +x kubeforge-cli

# Optionally, move to a directory in your PATH
sudo mv kubeforge-cli /usr/local/bin/
```

## Usage

```
kubeforge-cli [flags]
```

### Flags

| Flag        | Description                                                           | Default             | Required                    |
| ----------- | --------------------------------------------------------------------- | ------------------- | --------------------------- |
| `-host`     | Remote host IP address                                                | -                   | Yes                         |
| `-port`     | SSH port                                                              | `22`                | No                          |
| `-user`     | SSH username                                                          | Depends on provider | No                          |
| `-key`      | Path to private key file                                              | -                   | Yes (unless using password) |
| `-password` | SSH password                                                          | -                   | Yes (unless using key)      |
| `-provider` | Cloud provider (`aws`, `gcp`, `azure`, `oracle`)                      | `aws`               | No                          |
| `-distro`   | Linux distribution (`ubuntu`, `debian`, `centos`, `amazon`, `oracle`) | Depends on provider | No                          |

### Examples

#### Install on AWS EC2 Instance

```bash
kubeforge-cli -host=54.123.45.67 -key=~/.ssh/aws-key.pem -provider=aws
```

#### Install on GCP Compute Engine VM

```bash
kubeforge-cli -host=35.123.45.67 -key=~/.ssh/gcp-key.pem -provider=gcp -user=username
```

#### Install on Azure VM

```bash
kubeforge-cli -host=20.123.45.67 -key=~/.ssh/azure-key.pem -provider=azure
```

#### Install on Oracle Cloud VM

```bash
kubeforge-cli -host=129.123.45.67 -key=~/.ssh/oracle-key.pem -provider=oracle
```

#### Specify Linux distribution

```bash
kubeforge-cli -host=54.123.45.67 -key=~/.ssh/my-key.pem -provider=aws -distro=ubuntu
```

#### Using password authentication instead of key

```bash
kubeforge-cli -host=54.123.45.67 -password=securepassword -provider=aws
```

## Detailed Implementation

### Project Structure

```
k8s-cloud-installer/
├── cmd/                 # Command-line applications
│   └── installer/       # Main installer command
│       └── main.go      # Entry point
├── pkg/                 # Library code
│   ├── config/          # Configuration handling
│   ├── ssh/             # SSH client operations
│   ├── providers/       # Cloud provider implementations
│   └── installer/       # Kubernetes installation logic
├── docs/                # Documentation
├── examples/            # Example scripts
└── scripts/             # Utility scripts
```

### Core Components

#### 1. Configuration (`pkg/config`)

The configuration package manages command-line flags and creates a Config struct with all necessary parameters for the installation. It handles validation and sets sensible defaults based on the chosen cloud provider.

Key components:

- Cloud provider detection
- Default username selection
- Default distribution selection
- SSH connection parameters

#### 2. SSH Client (`pkg/ssh`)

This package provides SSH client functionality for executing commands on the remote VM. It supports both key-based and password authentication.

Key functionality:

- Creating SSH connections
- Running commands and retrieving output
- Managing sessions
- Error handling

#### 3. Cloud Providers (`pkg/providers`)

Each cloud provider has specific implementation details for integrating Kubernetes:

**AWS Provider**:

- EC2 instance metadata handling
- IAM role configuration
- AWS cloud provider setup

**GCP Provider**:

- GCE VM metadata handling
- Service account configuration
- GCP cloud provider setup

**Azure Provider**:

- Azure VM metadata handling
- Managed identity configuration
- Azure cloud provider setup

**Oracle Provider**:

- Oracle Cloud VM configuration
- Network setup

#### 4. Installer Components (`pkg/installer`)

The installer package contains the core logic for setting up Kubernetes:

**Prerequisites Installation**:

- Disables swap
- Configures kernel modules
- Sets up networking parameters
- Installs base packages

**Container Runtime Installation**:

- Installs containerd
- Configures containerd with systemd cgroup driver
- Ensures proper startup

**Kubernetes Components Installation**:

- Installs kubeadm, kubelet, kubectl
- Configures repositories
- Prepares for initialization

**Cluster Initialization**:

- Initializes Kubernetes with kubeadm
- Sets up network plugin (Flannel)
- Configures kubectl for the user
- Integrates with cloud provider
- Creates join command for additional nodes

## Implementation Details

### 1. Package Management

The system detects the appropriate package manager (apt for Debian/Ubuntu, yum for RHEL/CentOS/Amazon/Oracle) and uses the correct commands:

```go
// Determines the right package manager commands for the distribution
func getPackageManager(distro string) map[string]string {
    switch distro {
    case "ubuntu", "debian":
        return map[string]string{
            "update":     "sudo apt-get update",
            "install":    "sudo apt-get install -y",
            "repository": "sudo apt-add-repository",
        }
    case "centos", "rhel", "amazon", "oracle":
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
```

### 2. SSH Client Implementation

The SSH client handles authentication and secure command execution:

```go
// Creates an SSH client using the provided configuration
func createSSHClient(config *Config) (*ssh.Client, error) {
    var authMethod ssh.AuthMethod

    if config.PrivateKey != "" {
        key, err := ioutil.ReadFile(config.PrivateKey)
        if err != nil {
            return nil, fmt.Errorf("unable to read private key: %v", err)
        }

        signer, err := ssh.ParsePrivateKey(key)
        if err != nil {
            return nil, fmt.Errorf("unable to parse private key: %v", err)
        }
        authMethod = ssh.PublicKeys(signer)
    } else {
        authMethod = ssh.Password(config.Password)
    }

    sshConfig := &ssh.ClientConfig{
        User: config.User,
        Auth: []ssh.AuthMethod{
            authMethod,
        },
        HostKeyCallback: ssh.InsecureIgnoreHostKey(),
        Timeout:         15 * time.Second,
    }

    addr := fmt.Sprintf("%s:%s", config.Host, config.Port)
    return ssh.Dial("tcp", addr, sshConfig)
}
```

### 3. Cloud-Specific Configurations

Each cloud provider requires specific configurations:

**AWS**:

- Configures the AWS cloud provider for Kubernetes
- Uses instance metadata for identification
- Requires specific IAM roles for proper integration

**GCP**:

- Integrates with Google Compute Engine features
- Configures service accounts for authentication
- Uses GCP-specific networking capabilities

**Azure**:

- Creates and configures azure.json for cloud provider
- Configures managed identities for authentication
- Sets up network security rules

**Oracle Cloud**:

- Handles Oracle-specific VM configuration
- Sets up networking for Kubernetes services

### 4. Kubernetes Initialization

The cluster initialization process follows these steps:

1. Prepare the environment (disable swap, load kernel modules)
2. Install and configure containerd as the container runtime
3. Install Kubernetes components (kubeadm, kubelet, kubectl)
4. Initialize the cluster with kubeadm
5. Configure networking with Flannel CNI
6. Set up cloud provider integration
7. Configure kubectl for the user
8. Generate a join command for additional nodes

### 5. Error Handling and Logging

The implementation includes comprehensive error handling and logging:

- All command outputs are captured and logged
- Detailed error messages help with troubleshooting
- The system fails gracefully if any step encounters an error

## Cloud Provider Integration

### AWS Integration

AWS integration uses the EC2 instance metadata service to configure the cloud provider:

```go
// Configures the AWS cloud provider integration
func setupAWSCloudProvider(client *ssh.Client) error {
    commands := []string{
        // Ensure the EC2 instance has the correct IAM role for AWS cloud provider
        "kubectl -n kube-system create secret generic aws-cloud-provider --from-literal=cloud-config=''",
    }

    for _, cmd := range commands {
        fmt.Printf("Running: %s\n", cmd)
        if _, err := runCommand(client, cmd); err != nil {
            return err
        }
    }

    return nil
}
```

### GCP Integration

GCP integration configures the necessary service account permissions:

```go
// Configures the GCP cloud provider integration
func setupGCPCloudProvider(client *ssh.Client) error {
    commands := []string{
        // GCP authentication is usually handled by the VM's service account
        "curl -s -H 'Metadata-Flavor: Google' http://metadata.google.internal/computeMetadata/v1/instance/service-accounts/default/token > /dev/null || " +
            "echo 'Warning: This VM may not have the correct service account for GCP integration'",
    }

    for _, cmd := range commands {
        fmt.Printf("Running: %s\n", cmd)
        if _, err := runCommand(client, cmd); err != nil {
            return err
        }
    }

    return nil
}
```

### Azure Integration

Azure integration creates and configures the necessary azure.json file:

```go
// Configures the Azure cloud provider integration
func setupAzureCloudProvider(client *ssh.Client) error {
    commands := []string{
        // Create a minimal azure.json configuration
        "cat <<EOF | sudo tee /etc/kubernetes/azure.json\n{\n  \"cloud\":\"AzurePublicCloud\",\n  \"useManagedIdentityExtension\":true\n}\nEOF",
        "sudo chmod 600 /etc/kubernetes/azure.json",
        // Create secret from the azure.json file
        "kubectl -n kube-system create secret generic azure-cloud-provider --from-file=/etc/kubernetes/azure.json",
    }

    for _, cmd := range commands {
        fmt.Printf("Running: %s\n", cmd)
        if _, err := runCommand(client, cmd); err != nil {
            return err
        }
    }

    return nil
}
```

## Troubleshooting

If you encounter issues during installation:

1. **SSH Connection Failures**:

   - Verify that the VM is accessible via SSH from your machine
   - Check that the username and key/password are correct
   - Ensure the VM's security group/firewall allows SSH access

2. **Package Installation Errors**:

   - Ensure the VM has internet access
   - Verify that the distribution was correctly specified
   - Check for sufficient disk space

3. **Kubernetes Initialization Failures**:
   - Verify VM has at least 2 CPUs and 2GB memory
   - Ensure all prerequisites are correctly installed
   - Check that container runtime is properly configured

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

See [CONTRIBUTING.md](CONTRIBUTING.md) for detailed guidelines.
