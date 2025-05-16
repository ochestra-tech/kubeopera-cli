# Oracle Cloud Provider Documentation

This document provides detailed information on using the k8s-cloud-installer with Oracle Cloud Infrastructure (OCI) virtual machines.

## Prerequisites

Before installing Kubernetes on an Oracle Cloud VM, ensure you have:

1. An OCI Compute instance with at least:
    - 2 OCPUs (equivalent to 2 vCPUs)
    - 4 GB RAM
    - 50 GB disk space
    - Oracle Linux 7/8, Ubuntu 20.04+, or CentOS 7+

2. Appropriate security list rules for your Virtual Cloud Network (VCN):
    - SSH (port 22) for installer access
    - Kubernetes API server (port 6443)
    - Kubernetes pod and service network traffic
    - Node ports (30000-32767) for NodePort services

3. Networking configuration:
    - VCN with Internet connectivity
    - Public subnet for external access (or private subnet with NAT gateway)
    - Security list allowing required traffic

## Installation

```bash
./k8s-installer -host=<ORACLE_VM_IP_ADDRESS> -key=<PATH_TO_PRIVATE_KEY> -provider=oracle
```

### Options

- `-user`: SSH username (default: `opc` for Oracle Linux, `ubuntu` for Ubuntu)
- `-distro`: Linux distribution (default: `oracle`, options: `ubuntu`, `centos`)
- `-port`: SSH port (default: `22`)

## Oracle Cloud-Specific Configuration

The installer automatically:

1. Sets the hostname from the system
2. Configures necessary kernel modules and parameters
3. Installs containerd and Kubernetes components
4. Creates a placeholder configuration for OCI integration

## Cloud Provider Limitations

Oracle Cloud does not have a native Kubernetes cloud provider integration built into Kubernetes. Instead, Oracle provides the **OCI Cloud Controller Manager** as a separate project.

The installer will:
1. Set up a basic Kubernetes cluster without cloud provider integration
2. Provide information about installing the OCI Cloud Controller Manager

## Setting Up OCI Cloud Controller Manager

After the installer completes, you can set up the OCI Cloud Controller Manager:

1. Clone the OCI Cloud Controller Manager repository:
   ```bash
   git clone https://github.com/oracle/oci-cloud-controller-manager.git
   cd oci-cloud-controller-manager
   ```

2. Create a configuration file with your OCI credentials:
   ```bash
   cat > cloud-provider.yaml <<EOF
   auth:
     region: <region>
     tenancy: <tenancy-ocid>
     user: <user-ocid>
     key: |
       -----BEGIN RSA PRIVATE KEY-----
       <private-key>
       -----END RSA PRIVATE KEY-----
     fingerprint: <key-fingerprint>
     compartment: <compartment-ocid>
   
   loadBalancer:
     subnet1: <subnet1-ocid>
     subnet2: <subnet2-ocid>
     securityListManagementMode: All
   EOF
   ```

3. Create a Kubernetes secret from this file:
   ```bash
   kubectl create secret generic oci-cloud-controller-manager \
     -n kube-system \
     --from-file=cloud-provider.yaml
   ```

4. Deploy the OCI Cloud Controller Manager:
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/oracle/oci-cloud-controller-manager/master/manifests/cloud-controller-manager/oci-cloud-controller-manager.yaml
   ```

## Load Balancer Configuration

Once the OCI Cloud Controller Manager is installed, you can create Kubernetes services of type LoadBalancer:

```yaml
apiVersion: v1
kind: Service
metadata:
  name: my-service
  annotations:
    service.beta.kubernetes.io/oci-load-balancer-shape: "10Mbps"
spec:
  type: LoadBalancer
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: my-app
```

Available annotations for OCI load balancers:
- `service.beta.kubernetes.io/oci-load-balancer-shape`: Bandwidth shape (10Mbps, 100Mbps, 400Mbps, or flexible)
- `service.beta.kubernetes.io/oci-load-balancer-subnet1`: Subnet OCID for the load balancer
- `service.beta.kubernetes.io/oci-load-balancer-subnet2`: Second subnet OCID for HA
- `service.beta.kubernetes.io/oci-load-balancer-internal`: Set to "true" for internal load balancer

## Block Volume Provisioning

For persistent storage with OCI Block Volumes:

1. Install the OCI Block Volume Provisioner:
   ```bash
   kubectl apply -f https://raw.githubusercontent.com/oracle/oci-cloud-controller-manager/master/manifests/blockvolume-provisioner/oci-block-volume-provisioner.yaml
   ```

2. Create a Kubernetes StorageClass:
   ```yaml
   apiVersion: storage.k8s.io/v1
   kind: StorageClass
   metadata:
     name: oci-block-storage
   provisioner: oracle.com/oci
   parameters:
     attachmentType: iscsi
     availabilityDomain: <availability-domain>
     compartment: <compartment-ocid>
   ```

3. Use this StorageClass in your PersistentVolumeClaims:
   ```yaml
   apiVersion: v1
   kind: PersistentVolumeClaim
   metadata:
     name: my-volume
   spec:
     accessModes:
       - ReadWriteOnce
     storageClassName: oci-block-storage
     resources:
       requests:
         storage: 50Gi
   ```

## Multiple Nodes

To add more nodes to your cluster, run the following on each new Oracle Cloud VM:

1. Install prerequisites:
   ```bash
   ./k8s-installer -host=<WORKER_VM_IP> -key=<PATH_TO_KEY> -provider=oracle -no-init=true
   ```

2. Run the join command (printed during master installation):
   ```bash
   sudo kubeadm join <MASTER_IP>:6443 --token <TOKEN> --discovery-token-ca-cert-hash <HASH>
   ```

## Troubleshooting

Common issues and solutions:

1. **VM not reachable**:
    - Check security list rules
    - Verify the VM is running
    - Ensure SSH key permissions are correct

2. **Network issues**:
    - Verify VCN configuration allows pod-to-pod communication
    - Check if security lists permit the necessary traffic
    - Ensure route tables have proper routes configured

3. **OCI Cloud Controller Manager issues**:
    - Verify OCI credentials and permissions
    - Check logs with: `kubectl logs -n kube-system -l app=oci-cloud-controller-manager`
    - Ensure all required OCIDs in the configuration are correct

4. **Load balancer not provisioning**:
    - Check subnet configuration and OCIDs
    - Verify security list management mode
    - Review load balancer logs

5. **Block volume issues**:
    - Verify availability domain is correct
    - Check if compartment has enough quota for block volumes
    - Examine logs: `kubectl logs -n kube-system -l app=oci-volume-provisioner`

## References

- [OCI Cloud Controller Manager](https://github.com/oracle/oci-cloud-controller-manager)
- [OCI Block Volume Provisioner](https://github.com/oracle/oci-cloud-controller-manager/blob/master/docs/volume-provisioner.md)
- [OCI Load Balancer Configuration](https://github.com/oracle/oci-cloud-controller-manager/blob/master/docs/load-balancer-annotations.md)
- [Oracle Container Engine for Kubernetes (OKE)](https://docs.oracle.com/en-us/iaas/Content/ContEng/Concepts/contengoverview.htm)
