#!/bin/bash

# Oracle Cloud Kubernetes Installer Example Script
# This script demonstrates how to use k8s-cloud-installer with Oracle Cloud

# Configuration
HOST="<ORACLE_VM_IP_ADDRESS>"  # Replace with your Oracle Cloud VM IP address
KEY_PATH="~/.ssh/oracle-key.pem"  # Replace with your SSH key path
USER="opc"  # Default user for Oracle Linux, use "ubuntu" for Ubuntu

# Print banner
echo "================================================"
echo "  Kubernetes Oracle Cloud Installation Example"
echo "================================================"
echo "Host: $HOST"
echo "User: $USER"
echo "Key:  $KEY_PATH"
echo

# Check if kubeforge-cli binary exists
if [ ! -f "./kubeforge-cli" ]; then
    echo "Building kubeforge-cli binary..."
    go build -o kubeforge-cli cmd/installer/main.go
    chmod +x kubeforge-cli
fi

# Run the installer
echo "Starting Kubernetes installation on Oracle Cloud..."
./kubeforge-cli -host="$HOST" -key="$KEY_PATH" -user="$USER" -provider=oracle

echo
echo "Installation complete!"
echo "To access your cluster:"
echo "1. SSH into your Oracle Cloud VM: ssh -i $KEY_PATH $USER@$HOST"
echo "2. Run kubectl commands: kubectl get nodes"
echo "3. Add local kubectl config (optional):"
echo "   scp -i $KEY_PATH $USER@$HOST:.kube/config ~/.kube/config-oracle"
echo "   export KUBECONFIG=~/.kube/config-oracle"
echo "================================================"