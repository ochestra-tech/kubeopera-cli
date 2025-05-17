#!/bin/bash

# AWS Kubernetes Installer Example Script
# This script demonstrates how to use k8s-cloud-installer with AWS

# Configuration
HOST="<EC2_IP_ADDRESS>"  # Replace with your EC2 instance IP address
KEY_PATH="~/.ssh/aws-key.pem"  # Replace with your SSH key path
USER="ec2-user"  # Default user for Amazon Linux, use "ubuntu" for Ubuntu AMIs

# Print banner
echo "================================================"
echo "  Kubernetes AWS Installation Example"
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
echo "Starting Kubernetes installation on AWS..."
./kubeforge-cli -host="$HOST" -key="$KEY_PATH" -user="$USER" -provider=aws

echo
echo "Installation complete!"
echo "To access your cluster:"
echo "1. SSH into your EC2 instance: ssh -i $KEY_PATH $USER@$HOST"
echo "2. Run kubectl commands: kubectl get nodes"
echo "3. Add local kubectl config (optional):"
echo "   scp -i $KEY_PATH $USER@$HOST:.kube/config ~/.kube/config-aws"
echo "   export KUBECONFIG=~/.kube/config-aws"
echo "================================================"