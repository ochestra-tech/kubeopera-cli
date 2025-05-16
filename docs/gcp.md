# GCP Cloud Provider Documentation

This document provides detailed information on using the k8s-cloud-installer with Google Cloud Platform (GCP) virtual machines.

## Prerequisites

Before installing Kubernetes on a GCP VM, ensure you have:

1. A Compute Engine VM with at least:

   - 2 vCPUs
   - 2 GB RAM
   - 20 GB disk space
   - Debian 10/11, Ubuntu 20.04+, or CentOS 7+

2. Appropriate firewall rules:

   - SSH (port 22) for installer access
   - Kubernetes API server (port 6443)
   - Kubernetes pod and service network traffic
   - Node ports (30000-32767) for NodePort services

3. Service account with:
   - Compute Admin role (`roles/compute.admin`)
   - Network Admin role (`roles/compute.networkAdmin`)
   - Required OAuth scopes on the VM:
     - `https://www.googleapis.com/auth/compute`
     - `https://www.googleapis.com/auth/devstorage.read_only`

## Installation

```bash
./k8s-installer -host=<GCP_VM_IP_ADDRESS> -key=<PATH_TO_PRIVATE_KEY> -user=<USERNAME> -provider=gcp
```

### Options

- `-user`: SSH username (often the username you configured when creating the VM)
- `-distro`: Linux distribution (default: `debian`, options: `ubuntu`, `centos`)
- `-port`: SSH port (default: `22`)

## GCP-Specific Configuration

The installer automatically:

1. Sets the hostname from GCP metadata
2. Configures the GCP cloud provider for Kubernetes
3. Sets up necessary kernel modules and parameters
4. Creates the required cloud configuration file

## Cloud Provider Integration

For proper GCP cloud provider integration, your VM needs:

1. Service account with appropriate permissions:

   - Required roles:
     ```
     roles/compute.admin
     roles/compute.networkAdmin
     roles/compute.securityAdmin (if using network policies)
     ```

2. OAuth scopes on the VM:

   ```
   https://www.googleapis.com/auth/compute
   https://www.googleapis.com/auth/devstorage.read_only
   ```

3. Network configuration:
   - Firewall rules allowing health check traffic (TCP:10256)
   - Firewall rules allowing NodePort services (TCP:30000-32767)

## Load Balancer Configuration

The GCP cloud provider automatically creates and configures load balancers when you create Kubernetes Service resources of type LoadBalancer.

To use load balancers:

1. Ensure your service account has the `compute.networkAdmin` role
2. Configure your service with appropriate annotations:
   ```yaml
   apiVersion: v1
   kind: Service
   metadata:
     name: my-service
     annotations:
       cloud.google.com/load-balancer-type: "External" # or "Internal"
   spec:
     type: LoadBalancer
     ports:
       - port: 80
         targetPort: 8080
     selector:
       app: my-app
   ```

## Multiple Nodes

To add more nodes to your cluster, run the following on each new GCP VM:

1. Install prerequisites:

   ```bash
   ./k8s-installer -host=<WORKER_VM_IP> -key=<PATH_TO_KEY> -provider=gcp -user=<USERNAME> -no-init=true
   ```

2. Run the join command (printed during master installation):
   ```bash
   sudo kubeadm join <MASTER_IP>:6443 --token <TOKEN> --discovery-token-ca-cert-hash <HASH>
   ```

## Persistent Volumes

For persistent storage:

1. Create a Kubernetes StorageClass for GCP persistent disks:

   ```yaml
   apiVersion: storage.k8s.io/v1
   kind: StorageClass
   metadata:
     name: standard
   provisioner: kubernetes.io/gce-pd
   parameters:
     type: pd-standard
     replication-type: none
   ```

2. Use this StorageClass in your PersistentVolumeClaims:
   ```yaml
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: my-data
   spec:
     accessModes:
       - ReadWriteOnce
     resources:
       requests:
         storage: 10Gi
     storageClassName: standard
   ```

## Troubleshooting

Common issues and solutions:

1. **VM not reachable**:

   - Check firewall rules
   - Verify the VM is running
   - Ensure SSH key permissions are correct

2. **Cloud provider integration issues**:

   - Verify service account has correct roles
   - Check OAuth scopes on the VM
   - Examine controller manager logs: `kubectl logs -n kube-system kube-controller-manager-<hostname>`

3. **Load balancer not provisioning**:

   - Check service account permissions
   - Verify network firewall rules
   - Check for quota issues in your GCP project

4. **Missing OAuth scopes**:
   - VM OAuth scopes cannot be changed after VM creation
   - Create a new VM with the correct scopes if necessary

## References

- [Kubernetes Cloud Provider GCP](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/#gce)
- [GCP Service Accounts](https://cloud.google.com/compute/docs/access/service-accounts)
- [GCP Firewall Rules](https://cloud.google.com/vpc/docs/firewalls)
- [Kubernetes on GCP Documentation](https://cloud.google.com/kubernetes-engine/docs/concepts/kubernetes-engine-overview)
