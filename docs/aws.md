# AWS Cloud Provider Documentation

This document provides detailed information on using the k8s-cloud-installer with AWS EC2 instances.

## Prerequisites

Before installing Kubernetes on an AWS EC2 instance, ensure you have:

1. An EC2 instance with at least:

   - 2 vCPUs
   - 2 GB RAM
   - 20 GB disk space
   - Ubuntu 20.04+, Amazon Linux 2, or CentOS 7+

2. Appropriate security group settings:

   - SSH (port 22) for installer access
   - Kubernetes API server (port 6443)
   - Kubernetes pod and service network traffic

3. IAM Role with:
   - AmazonEC2FullAccess (for the cloud provider integration)
   - AmazonRoute53FullAccess (if using Route53 for DNS)

## Installation

```bash
./k8s-installer -host=<EC2_IP_ADDRESS> -key=<PATH_TO_PRIVATE_KEY> -provider=aws
```

### Options

- `-user`: SSH username (default: `ec2-user` for Amazon Linux, `ubuntu` for Ubuntu)
- `-distro`: Linux distribution (default: `amazon`, options: `ubuntu`, `centos`)
- `-port`: SSH port (default: `22`)

## AWS-Specific Configuration

The installer automatically:

1. Sets the hostname from EC2 metadata
2. Configures the AWS cloud provider for Kubernetes
3. Sets up necessary kernel modules and parameters

## Cloud Provider Integration

For proper AWS cloud provider integration, your EC2 instance needs:

1. IAM Role with appropriate permissions:

   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": [
           "ec2:DescribeInstances",
           "ec2:DescribeRegions",
           "ec2:CreateTags",
           "ec2:DescribeTags",
           "ec2:DescribeVolumes",
           "ec2:CreateVolume",
           "ec2:DeleteVolume",
           "ec2:AttachVolume",
           "ec2:DetachVolume"
         ],
         "Resource": "*"
       }
     ]
   }
   ```

2. Tag all AWS resources with:

   ```
   KubernetesCluster=<cluster-name>
   ```

3. For ELB integration, tag subnets with:
   ```
   kubernetes.io/cluster/<cluster-name>=shared
   ```

## Multiple Nodes

To add more nodes to your cluster, run the following on each new EC2 instance:

1. Install prerequisites:

   ```bash
   ./k8s-installer -host=<WORKER_EC2_IP> -key=<PATH_TO_KEY> -provider=aws -no-init=true
   ```

2. Run the join command (printed during master installation):
   ```bash
   sudo kubeadm join <MASTER_IP>:6443 --token <TOKEN> --discovery-token-ca-cert-hash <HASH>
   ```

## Troubleshooting

Common issues and solutions:

1. **Instance not reachable**:

   - Check security group settings
   - Verify the instance is running
   - Ensure SSH key permissions are correct

2. **Cloud provider not working**:

   - Verify IAM role is attached to the instance
   - Check that the required tags are applied to resources
   - Examine logs: `kubectl logs -n kube-system kube-controller-manager-<hostname>`

3. **Network issues**:
   - Ensure VPC settings allow pod-to-pod communication
   - Check if security groups allow necessary ports
   - Verify route tables for proper networking

## References

- [Kubernetes Cloud Provider AWS](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/#aws)
- [AWS IAM Role for Kubernetes](https://docs.aws.amazon.com/eks/latest/userguide/create-service-account-iam-policy-and-role.html)
- [kubeadm Configuration](https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm-init/)
