# Azure Cloud Provider Documentation

This document provides detailed information on using the k8s-cloud-installer with Microsoft Azure virtual machines.

## Prerequisites

Before installing Kubernetes on an Azure VM, ensure you have:

1. An Azure VM with at least:

   - 2 vCPUs (B2s or larger recommended)
   - 2 GB RAM
   - 30 GB disk space
   - Ubuntu 20.04+, CentOS 7+, or RHEL 7+

2. Appropriate Network Security Group (NSG) rules:

   - SSH (port 22) for installer access
   - Kubernetes API server (port 6443)
   - Kubernetes pod and service network traffic (all internal traffic)
   - Node ports (30000-32767) for NodePort services

3. VM Identity configuration:
   - System-assigned Managed Identity enabled on the VM
   - Contributor role on the resource group
   - Network Contributor role for load balancer configuration

## Installation

```bash
./kubeforge-cli -host=<AZURE_VM_IP_ADDRESS> -key=<PATH_TO_PRIVATE_KEY> -provider=azure
```

### Options

- `-user`: SSH username (default: `azureuser` for Ubuntu, `adminuser` for other distributions)
- `-distro`: Linux distribution (default: `ubuntu`, options: `centos`)
- `-port`: SSH port (default: `22`)

## Azure-Specific Configuration

The installer automatically:

1. Sets the hostname from Azure metadata
2. Configures the Azure cloud provider for Kubernetes
3. Creates the required azure.json configuration file
4. Sets up necessary kernel modules and parameters

## Cloud Provider Integration

For proper Azure cloud provider integration, your VM needs:

1. System-assigned Managed Identity with roles:

   - Contributor role on the resource group
   - Network Contributor role

2. All VMs in the cluster should be in the same:

   - Resource group
   - Virtual network
   - Subnet

3. The azure.json configuration file includes:
   ```json
   {
     "cloud": "AzurePublicCloud",
     "tenantId": "",
     "subscriptionId": "<subscription-id>",
     "resourceGroup": "<resource-group>",
     "location": "<azure-region>",
     "useManagedIdentityExtension": true,
     "useInstanceMetadata": true
   }
   ```

## Load Balancer Configuration

The Azure cloud provider automatically creates and configures load balancers when you create Kubernetes Service resources of type LoadBalancer:

1. Create a service with type LoadBalancer:

   ```yaml
   apiVersion: v1
   kind: Service
   metadata:
     name: my-service
     annotations:
       service.beta.kubernetes.io/azure-load-balancer-internal: "false"
   spec:
     type: LoadBalancer
     ports:
       - port: 80
         targetPort: 8080
     selector:
       app: my-app
   ```

2. For internal load balancers, use the annotation:
   ```yaml
   service.beta.kubernetes.io/azure-load-balancer-internal: "true"
   ```

## Multiple Nodes

To add more nodes to your cluster, run the following on each new Azure VM:

1. Install prerequisites:

   ```bash
   ./kubeforge-cli -host=<WORKER_VM_IP> -key=<PATH_TO_KEY> -provider=azure -no-init=true
   ```

2. Run the join command (printed during master installation):
   ```bash
   sudo kubeadm join <MASTER_IP>:6443 --token <TOKEN> --discovery-token-ca-cert-hash <HASH>
   ```

## Persistent Volumes

For persistent storage with Azure Disks:

1. Create a Kubernetes StorageClass:

   ```yaml
   apiVersion: storage.k8s.io/v1
   kind: StorageClass
   metadata:
     name: managed-premium
   provisioner: kubernetes.io/azure-disk
   parameters:
     storageaccounttype: Premium_LRS
     kind: Managed
   ```

2. Use this StorageClass in your PersistentVolumeClaims:
   ```yaml
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: azure-managed-disk
   spec:
     accessModes:
       - ReadWriteOnce
     storageClassName: managed-premium
     resources:
       requests:
         storage: 5Gi
   ```

## Troubleshooting

Common issues and solutions:

1. **VM not reachable**:

   - Check Network Security Group (NSG) rules
   - Verify the VM is running
   - Ensure SSH key permissions are correct

2. **Cloud provider not working**:

   - Verify Managed Identity is enabled on the VM
   - Check that the VM has the required role assignments
   - Examine logs: `kubectl logs -n kube-system kube-controller-manager-<hostname>`

3. **Load balancer not provisioning**:

   - Check Network Contributor role assignment
   - Verify NSG rules allow the required traffic
   - Check service annotations

4. **Disk provisioning issues**:

   - Verify VM has permissions to create/attach disks
   - Check for quota issues in your Azure subscription
   - Examine if the storage class parameters are correct

5. **Authentication failures**:
   - Ensure the azure.json file has correct information
   - Verify Managed Identity is working with:
     ```
     curl -H Metadata:true "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https://management.azure.com/"
     ```

## References

- [Kubernetes Cloud Provider Azure](https://kubernetes.io/docs/concepts/cluster-administration/cloud-providers/#azure)
- [Azure Managed Identities](https://docs.microsoft.com/en-us/azure/active-directory/managed-identities-azure-resources/overview)
- [azure-cloud-provider GitHub](https://github.com/kubernetes-sigs/cloud-provider-azure)
- [Azure Disk Storage Classes](https://docs.microsoft.com/en-us/azure/aks/azure-disks-dynamic-pv)
