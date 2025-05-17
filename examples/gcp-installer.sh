#!/bin/bash

# GCP Kubernetes Installer Example Script
# This script demonstrates how to use k8s-cloud-installer with GCP

# Configuration
HOST="<GCP_VM_IP_ADDRESS>"  # Replace with your GCP VM IP address
KEY_PATH="~/.ssh/gcp-key.pem"  # Replace with your SSH key path
USER="<USERNAME>"  # Replace with your GCP VM username

# Print banner
echo "================================================"
echo "  Kubernetes GCP Installation Example"
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
echo "Starting Kubernetes installation on GCP..."
./kubeforge-cli -host="$HOST" -key="$KEY_PATH" -user="$USER" -provider=gcp

echo
echo "Installation complete!"
echo "To access your cluster:"
echo "1. SSH into your GCP VM: ssh -i $KEY_PATH $USER@$HOST"
echo "2. Run kubectl commands: kubectl get nodes"
echo "3. Add local kubectl config (optional):"
echo "   scp -i $KEY_PATH $USER@$HOST:.kube/config ~/.kube/config-gcp"
echo "   export KUBECONFIG=~/.kube/config-gcp"
echo "================================================"