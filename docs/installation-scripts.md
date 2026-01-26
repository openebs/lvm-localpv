# Installation Scripts Guide

This document describes how to use the installation scripts provided in the `deploy/scripts/` directory for installing, upgrading, and uninstalling OpenEBS LVM LocalPV.

## Overview

The `deploy/scripts/` directory contains scripts for managing OpenEBS LVM LocalPV:

- **install.sh** - Installation script for deploying OpenEBS LVM LocalPV
- **uninstall.sh** - Uninstallation script for removing OpenEBS LVM LocalPV
- **upgrade.sh** - Upgrade script for updating to a new version
- **download.sh** - Download script for offline media (charts and images)
- **load-images.sh** - Image loading script for offline installations

These scripts provide a simplified way to manage OpenEBS LVM LocalPV compared to using Helm commands directly, with features like:

- Automatic prerequisite checking
- Environment validation
- Installation verification
- Rollback on failure
- Comprehensive logging
- Support for both online and offline installation

## Prerequisites

Before using the installation scripts, ensure you have:

1. **Kubernetes cluster** (version 1.23 or higher)
2. **kubectl** installed and configured to access your cluster
3. **helm** (version 3.x) installed
4. **LVM2** utilities installed on all nodes
5. **dm-snapshot** kernel module loaded on all nodes
6. **Volume Groups (VG)** created on nodes where volumes will be provisioned

> **Note**: The scripts will check for kubectl and helm, but will not automatically install them. You must install these tools manually before running the scripts.

## Installation Script (install.sh)

The installation script automates the deployment of OpenEBS LVM LocalPV using Helm charts.

### Quick Start

```bash
# Set required environment variables
export OPENEBS_CONTROLLER_NODE_NAMES="master01,master02"
export OPENEBS_DATA_NODE_NAMES="node01,node02"
export OPENEBS_STORAGECLASS_NAME="openebs-lvmpv"
export OPENEBS_VG_NAME="lvmvg"

# Run installation
cd deploy/scripts
./install.sh
```

### Required Environment Variables

The following environment variables **must** be set before running the installation script:

| Variable                          | Description                                   | Example                 |
| --------------------------------- | --------------------------------------------- | ----------------------- |
| `OPENEBS_CONTROLLER_NODE_NAMES` | Comma-separated list of controller node names | `"master01,master02"` |
| `OPENEBS_DATA_NODE_NAMES`       | Comma-separated list of data node names       | `"node01,node02"`     |
| `OPENEBS_STORAGECLASS_NAME`     | Name for the StorageClass to create           | `"openebs-lvmpv"`     |
| `OPENEBS_VG_NAME`               | LVM volume group name                         | `"lvmvg"`             |

### Optional Environment Variables

| Variable                        | Description                                 | Default       |
| ------------------------------- | ------------------------------------------- | ------------- |
| `OFFLINE_INSTALL`             | Set to `"true"` for offline installation  | `"false"`   |
| `OPENEBS_KUBE_NAMESPACE`      | Kubernetes namespace                        | `"openebs"` |
| `OPENEBS_CREATE_STORAGECLASS` | Set to `"true"` to create StorageClass    | `"false"`   |
| `OPENEBS_CHART_VERSION`       | Chart version (overrides auto-detection)    | Auto-detected |
| `OPENEBS_CHART_DIR`           | Path to local chart directory (for offline) | -             |
| `OPENEBS_IMAGE_REGISTRY`      | Image registry for offline installation     | -             |
| `OPENEBS_STORAGECLASS_YAML`   | Path to StorageClass YAML template          | -             |

### Command Line Options

```bash
./install.sh [OPTIONS]
```

| Option                      | Description                                             |
| --------------------------- | ------------------------------------------------------- |
| `--help`                  | Show help message and exit                              |
| `--version`               | Show script version and exit                            |
| `--dry-run`               | Preview installation without executing                  |
| `--offline`               | Enable offline installation mode                        |
| `--chart-version VERSION` | Specify chart version (overrides auto-detection)        |
| `--namespace NAMESPACE`   | Kubernetes namespace (default: openebs)                 |
| `--create-storageclass`   | Automatically create StorageClass after installation    |
| `--log-level LEVEL`       | Set log level: DEBUG, INFO, WARN, ERROR (default: INFO) |

### Installation Examples

#### Example 1: Basic Online Installation

```bash
export OPENEBS_CONTROLLER_NODE_NAMES="master01"
export OPENEBS_DATA_NODE_NAMES="node01,node02"
export OPENEBS_STORAGECLASS_NAME="openebs-lvmpv"
export OPENEBS_VG_NAME="lvmvg"
export OPENEBS_CREATE_STORAGECLASS=true

./install.sh
```

#### Example 2: Offline Installation

```bash
export OPENEBS_CONTROLLER_NODE_NAMES="master01"
export OPENEBS_DATA_NODE_NAMES="node01,node02"
export OPENEBS_STORAGECLASS_NAME="openebs-lvmpv"
export OPENEBS_VG_NAME="lvmvg"
export OFFLINE_INSTALL=true
export OPENEBS_CHART_DIR="./charts"
export OPENEBS_IMAGE_REGISTRY="registry.example.com"

./install.sh --offline
```

#### Example 3: Dry Run (Preview)

```bash
export OPENEBS_CONTROLLER_NODE_NAMES="master01"
export OPENEBS_DATA_NODE_NAMES="node01,node02"
export OPENEBS_STORAGECLASS_NAME="openebs-lvmpv"
export OPENEBS_VG_NAME="lvmvg"

./install.sh --dry-run
```

#### Example 4: Custom Namespace and Version

```bash
export OPENEBS_CONTROLLER_NODE_NAMES="master01"
export OPENEBS_DATA_NODE_NAMES="node01,node02"
export OPENEBS_STORAGECLASS_NAME="openebs-lvmpv"
export OPENEBS_VG_NAME="lvmvg"

./install.sh --namespace my-namespace --chart-version 1.8.0 --create-storageclass
```

### What the Installation Script Does

1. **Prerequisite Checking**: Verifies kubectl, helm, and cluster connectivity
2. **Environment Validation**: Checks that all required environment variables are set
3. **Node Labeling**: Automatically labels controller and data nodes
4. **Helm Repository Setup**: Adds and updates the Helm repository (online mode)
5. **Chart Installation**: Installs the Helm chart with appropriate configuration
6. **Verification**: Verifies that pods, CRDs, and LVMNode resources are created
7. **StorageClass Creation**: Optionally creates a StorageClass (if enabled)

### Installation Verification

After installation, verify the deployment:

```bash
# Check Helm release status
helm status openebs-lvmlocalpv -n openebs

# Check pods
kubectl get pods -n openebs -l role=openebs-lvm

# Check LVMNode resources
kubectl get lvmnode -n openebs

# Check StorageClass (if created)
kubectl get storageclass openebs-lvmpv
```

## Uninstallation Script (uninstall.sh)

The uninstallation script removes OpenEBS LVM LocalPV from your cluster.

### Quick Start

```bash
./uninstall.sh
```

### Command Line Options

```bash
./uninstall.sh [OPTIONS]
```

| Option                    | Description                                             |
| ------------------------- | ------------------------------------------------------- |
| `--help`                | Show help message and exit                              |
| `--namespace NAMESPACE` | Kubernetes namespace (default: openebs)                 |
| `--release RELEASE`     | Helm release name (default: openebs-lvmlocalpv)         |
| `--keep-crds`           | Keep CRDs after uninstallation                          |
| `--keep-storageclass`   | Keep StorageClass resources after uninstallation        |
| `--delete-pvcs`         | Delete all PVCs using this driver (WARNING: data loss!) |
| `--force`               | Force uninstallation without confirmation prompt        |
| `--log-level LEVEL`     | Set log level: DEBUG, INFO, WARN, ERROR                 |

### Uninstallation Examples

#### Example 1: Standard Uninstallation

```bash
./uninstall.sh
```

This will:

- Uninstall the Helm release
- Delete CRDs (unless `--keep-crds` is used)
- Delete StorageClass resources (unless `--keep-storageclass` is used)
- Keep PVCs (data remains on nodes)

#### Example 2: Keep CRDs and StorageClass

```bash
./uninstall.sh --keep-crds --keep-storageclass
```

#### Example 3: Force Uninstall Without Confirmation

```bash
./uninstall.sh --force
```

#### Example 4: Delete PVCs (Dangerous!)

```bash
# WARNING: This will delete all PVCs and result in data loss!
./uninstall.sh --delete-pvcs --force
```

> **Warning**: Using `--delete-pvcs` will permanently delete all PersistentVolumeClaims using this driver. This action cannot be undone and will result in data loss. Only use this option if you are certain you want to delete all PVCs.

### What Gets Removed

By default, the uninstallation script removes:

- Helm release (controller and node components)
- CRDs (LVMVolume, LVMNode, LVMSnapshot)
- StorageClass resources with provisioner `local.csi.openebs.io`

**What is NOT removed by default:**

- PVCs (PersistentVolumeClaims) - data remains on nodes
- Namespace (if empty, you'll be prompted to delete it)
- LVM volumes on nodes (physical data remains)

## Upgrade Script (upgrade.sh)

The upgrade script upgrades OpenEBS LVM LocalPV to a new version.

### Quick Start

```bash
# Upgrade to latest version
./upgrade.sh

# Upgrade to specific version
./upgrade.sh --chart-version 1.8.0
```

### Command Line Options

```bash
./upgrade.sh [OPTIONS]
```

| Option                      | Description                                      |
| --------------------------- | ------------------------------------------------ |
| `--help`                  | Show help message and exit                       |
| `--namespace NAMESPACE`   | Kubernetes namespace (default: openebs)          |
| `--release RELEASE`       | Helm release name (default: openebs-lvmlocalpv)  |
| `--chart-version VERSION` | Target chart version to upgrade to               |
| `--offline`               | Use offline mode with local chart                |
| `--chart-dir DIR`         | Path to local chart directory (for offline mode) |
| `--dry-run`               | Preview upgrade without executing                |
| `--force`                 | Force upgrade without confirmation               |
| `--log-level LEVEL`       | Set log level: DEBUG, INFO, WARN, ERROR          |

### Upgrade Examples

#### Example 1: Upgrade to Latest Version

```bash
./upgrade.sh
```

#### Example 2: Upgrade to Specific Version

```bash
./upgrade.sh --chart-version 1.8.0
```

#### Example 3: Offline Upgrade

```bash
export OFFLINE_INSTALL=true
export OPENEBS_CHART_DIR="./charts"
export OPENEBS_IMAGE_REGISTRY="registry.example.com"

./upgrade.sh --offline
```

#### Example 4: Dry Run (Preview)

```bash
./upgrade.sh --dry-run
```

#### Example 5: Force Upgrade Without Confirmation

```bash
./upgrade.sh --force
```

### What the Upgrade Script Does

1. **Pre-upgrade Checks**: Verifies the current installation exists
2. **Configuration Backup**: Backs up current Helm values and release information
3. **Repository Update**: Updates Helm repository (online mode)
4. **Helm Upgrade**: Performs rolling upgrade of the Helm release
5. **Verification**: Verifies that pods are running and healthy after upgrade

### Upgrade Notes

- The upgrade process performs a **rolling update**, ensuring zero downtime
- Existing PVCs and data are **preserved** during upgrade
- It's recommended to **backup your configuration** before upgrading
- Check **release notes** for breaking changes between versions
- The script automatically backs up your configuration to `/tmp/openebs-lvmlocalpv-backup-*`

## Logging

All scripts generate detailed logs for troubleshooting. Log files are created in:

- **Installation**: `/tmp/openebs-lvmlocalpv_install-YYYY-MM-DD_HH-MM-SS.log`
- **Uninstallation**: `/tmp/openebs-lvmlocalpv_uninstall-YYYY-MM-DD_HH-MM-SS.log`
- **Upgrade**: `/tmp/openebs-lvmlocalpv_upgrade-YYYY-MM-DD_HH-MM-SS.log`

If `/tmp` is not writable, logs are created in the current directory.

### Log Levels

You can control the verbosity of logs using the `--log-level` option:

- **ERROR**: Only error messages
- **WARN**: Warnings and errors
- **INFO**: Informational messages, warnings, and errors (default)
- **DEBUG**: All messages including debug information

Example:

```bash
./install.sh --log-level DEBUG
```

## Troubleshooting

### Installation Issues

#### Issue: "Missing required tools"

**Solution**: Install the missing tools (kubectl, helm, curl, envsubst) before running the script.

```bash
# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

# Install helm
curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

#### Issue: "Cannot connect to Kubernetes cluster"

**Solution**: Verify your kubeconfig is set correctly:

```bash
kubectl cluster-info
```

#### Issue: "Failed to label node"

**Solution**: Ensure you have proper permissions and the node names are correct:

```bash
# Check node names
kubectl get nodes

# Verify permissions
kubectl auth can-i create nodes
```

#### Issue: "Helm installation failed"

**Solution**: Check the logs for detailed error messages:

```bash
# View installation log
tail -f /tmp/openebs-lvmlocalpv_install-*.log

# Check Helm release status
helm status openebs-lvmlocalpv -n openebs

# Check pod logs
kubectl logs -n openebs -l app=openebs-lvm-controller
```

### Uninstallation Issues

#### Issue: "CRD deletion failed"

**Solution**: This usually means there are dependent resources. Check for remaining LVMVolume or LVMSnapshot resources:

```bash
# List remaining resources
kubectl get lvmvolume -A
kubectl get lvmsnapshot -A

# Delete them manually if needed
kubectl delete lvmvolume <name> -n <namespace>
```

#### Issue: "PVCs still exist after uninstall"

**Solution**: This is expected behavior. PVCs are not deleted by default to preserve data. To delete them:

```bash
# List PVCs
kubectl get pvc -A

# Delete manually if needed
kubectl delete pvc <name> -n <namespace>
```

### Upgrade Issues

#### Issue: "Helm release not found"

**Solution**: Ensure the release exists and you're using the correct namespace:

```bash
# List releases
helm list -A

# Check if release exists
helm status openebs-lvmlocalpv -n openebs
```

#### Issue: "Upgrade failed - pods not ready"

**Solution**: Check pod status and logs:

```bash
# Check pod status
kubectl get pods -n openebs

# Check pod logs
kubectl logs -n openebs -l app=openebs-lvm-controller

# Check events
kubectl get events -n openebs --sort-by='.lastTimestamp'
```

## Best Practices

### Before Installation

1. **Review Prerequisites**: Ensure all prerequisites are met
2. **Plan Node Labels**: Decide which nodes will be controllers and which will be data nodes
3. **Prepare Volume Groups**: Create LVM volume groups on all data nodes
4. **Test Connectivity**: Verify kubectl can connect to your cluster
5. **Use Dry Run**: Test the installation with `--dry-run` first

### During Installation

1. **Use Logging**: Enable DEBUG logging for detailed information
2. **Monitor Progress**: Watch the installation logs in real-time
3. **Verify Installation**: After installation, verify all components are running

### After Installation

1. **Verify Pods**: Ensure all pods are running
2. **Check LVMNode**: Verify LVMNode resources are created
3. **Test StorageClass**: Create a test PVC to verify StorageClass works
4. **Monitor Logs**: Check controller and node logs for any issues

### Before Upgrade

1. **Backup Configuration**: The script does this automatically, but keep your own backup
2. **Review Release Notes**: Check for breaking changes
3. **Test in Non-Production**: Test upgrades in a non-production environment first
4. **Use Dry Run**: Preview the upgrade with `--dry-run`

### Before Uninstallation

1. **Backup Data**: Ensure you have backups of important data
2. **List Resources**: Check what will be removed
3. **Plan for Data**: Decide what to do with existing PVCs
4. **Use Dry Run**: If available, preview the uninstallation

## Environment Variables Reference

### Installation Variables

| Variable                          | Required | Description                                   |
| --------------------------------- | -------- | --------------------------------------------- |
| `OPENEBS_CONTROLLER_NODE_NAMES` | Yes      | Comma-separated controller node names         |
| `OPENEBS_DATA_NODE_NAMES`       | Yes      | Comma-separated data node names               |
| `OPENEBS_STORAGECLASS_NAME`     | Yes      | StorageClass name                             |
| `OPENEBS_VG_NAME`               | Yes      | LVM volume group name                         |
| `OFFLINE_INSTALL`               | No       | Set to "true" for offline installation        |
| `OPENEBS_KUBE_NAMESPACE`        | No       | Kubernetes namespace (default: openebs)       |
| `OPENEBS_CREATE_STORAGECLASS`   | No       | Set to "true" to create StorageClass          |
| `OPENEBS_CHART_VERSION`         | No       | Chart version (overrides auto-detection)      |
| `OPENEBS_CHART_DIR`             | No*      | Local chart directory (*required for offline) |
| `OPENEBS_IMAGE_REGISTRY`        | No       | Image registry for offline installation       |
| `OPENEBS_STORAGECLASS_YAML`     | No       | Path to StorageClass YAML template            |

### Resource Limit Variables (Optional)

All resource limit variables have defaults and are optional:

- `OPENEBS_CONTROLLER_RESOURCE_LIMITS_CPU` (default: 500m)
- `OPENEBS_CONTROLLER_RESOURCE_LIMITS_MEMORY` (default: 512Mi)
- `OPENEBS_CONTROLLER_RESOURCE_REQUESTS_CPU` (default: 500m)
- `OPENEBS_CONTROLLER_RESOURCE_REQUESTS_MEMORY` (default: 512Mi)
- `OPENEBS_NODE_RESOURCE_LIMITS_CPU` (default: 500m)
- `OPENEBS_NODE_RESOURCE_LIMITS_MEMORY` (default: 512Mi)
- `OPENEBS_NODE_RESOURCE_REQUESTS_CPU` (default: 500m)
- `OPENEBS_NODE_RESOURCE_REQUESTS_MEMORY` (default: 512Mi)

## Compatibility

### Backward Compatibility

The original `deploy/install.sh` script has been moved to `deploy/scripts/install.sh`. A compatibility script remains at `deploy/install.sh` that automatically redirects to the new location, ensuring existing scripts and documentation continue to work.

### Script Versions

- **install.sh**: v1.0.0
- **uninstall.sh**: v1.0.0
- **upgrade.sh**: v1.0.0
- **download.sh**: v1.0.0
- **load-images.sh**: v1.0.0

## Download Script (download.sh)

The download script downloads all required media (Helm charts and Docker images) for offline installation.

### Quick Start

```bash
# Download latest version
./download.sh

# Download specific version and pack
./download.sh --chart-version 1.8.0 --pack
```

### Command Line Options

```bash
./download.sh [OPTIONS]
```

| Option                  | Description                                 |
| ----------------------- | ------------------------------------------- |
| `--help`              | Show help message and exit                  |
| `--version`           | Show script version and exit                |
| `--chart-version VER` | Chart version to download (default: latest) |
| `--output-dir DIR`    | Output directory (default: ./offline-media) |
| `--pack`              | Pack all media into tar.gz after download   |
| `--log-level LEVEL`   | Set log level: DEBUG, INFO, WARN, ERROR     |

### Download Examples

#### Example 1: Download Latest Version

```bash
./download.sh
```

#### Example 2: Download Specific Version and Pack

```bash
./download.sh --chart-version 1.8.0 --pack
```

This will:

- Download Helm chart version 1.8.0
- Download all required Docker images
- Save images as tar files
- Pack everything into `offline-media-1.8.0.tar.gz`

#### Example 3: Custom Output Directory

```bash
./download.sh --output-dir /path/to/media --pack
```

### Output Structure

After downloading, you'll have the following structure:

```
offline-media/
├── charts/
│   └── lvm-localpv-<version>.tgz
├── images/
│   ├── registry.k8s.io_sig-storage_csi-node-driver-registrar_v2.13.0.tar
│   ├── registry.k8s.io_sig-storage_csi-resizer_v1.11.2.tar
│   ├── registry.k8s.io_sig-storage_csi-snapshotter_v7.0.0.tar
│   ├── registry.k8s.io_sig-storage_snapshot-controller_v7.0.0.tar
│   ├── registry.k8s.io_sig-storage_csi-provisioner_v5.2.0.tar
│   ├── docker.io_openebs_lvm-driver_<version>.tar
│   └── images.list
├── manifests/
│   ├── VERSION
│   └── checksums.txt
└── README.md
```

### What Gets Downloaded

The script downloads:

1. **Helm Chart**: The complete Helm chart package (.tgz file)
2. **Docker Images**: All images required by the chart:
   - CSI Node Driver Registrar
   - CSI Resizer
   - CSI Snapshotter
   - Snapshot Controller
   - CSI Provisioner
   - OpenEBS LVM Driver

## Image Loading Script (load-images.sh)

The image loading script loads Docker images from tar files into your local container registry.

### Quick Start

```bash
# Load images from default directory
./load-images.sh

# Load images from offline media
./load-images.sh --images-dir ./offline-media/images
```

### Command Line Options

```bash
./load-images.sh [OPTIONS]
```

| Option                  | Description                                              |
| ----------------------- | -------------------------------------------------------- |
| `--help`              | Show help message and exit                               |
| `--images-dir DIR`    | Directory containing image tar files (default: ./images) |
| `--registry REGISTRY` | Target registry to tag images (optional)                 |
| `--push`              | Push images to registry after loading and tagging        |
| `--log-level LEVEL`   | Set log level: DEBUG, INFO, WARN, ERROR                  |

### Loading Examples

#### Example 1: Load Images to Local Registry

```bash
./load-images.sh --images-dir ./offline-media/images
```

#### Example 2: Load and Push to Private Registry

```bash
./load-images.sh \
  --images-dir ./offline-media/images \
  --registry registry.example.com \
  --push
```

#### Example 3: Tag for Registry (Without Pushing)

```bash
./load-images.sh \
  --images-dir ./offline-media/images \
  --registry registry.example.com
```

## Offline Installation Workflow

### Step 1: Download Media (On Machine with Internet)

On a machine with internet access:

```bash
cd deploy/scripts

# Download and pack media
./download.sh --chart-version 1.8.0 --pack
```

This creates `offline-media-1.8.0.tar.gz` in the parent directory.

### Step 2: Transfer Media to Offline Environment

Transfer the packed file to your offline environment:

```bash
# Using SCP
scp offline-media-1.8.0.tar.gz user@offline-server:/path/

# Or using other transfer methods (USB, etc.)
```

### Step 3: Extract Media (On Offline Machine)

```bash
# Extract the media
tar -xzf offline-media-1.8.0.tar.gz

# This creates offline-media/ directory with all files
```

### Step 4: Load Images (On Offline Machine)

```bash
cd deploy/scripts

# Option 1: Load to local Docker/Podman registry
./load-images.sh --images-dir ../offline-media/images

# Option 2: Load and push to private registry
./load-images.sh \
  --images-dir ../offline-media/images \
  --registry your-registry.example.com \
  --push
```

### Step 5: Install Using Media (On Offline Machine)

```bash
# Set required environment variables
export OPENEBS_CONTROLLER_NODE_NAMES="master01"
export OPENEBS_DATA_NODE_NAMES="node01,node02"
export OPENEBS_STORAGECLASS_NAME="openebs-lvmpv"
export OPENEBS_VG_NAME="lvmvg"

# Option 1: Automatic detection (if media is in ./offline-media/)
export OFFLINE_INSTALL=true
export OPENEBS_LOAD_IMAGES=true  # Automatically load images
./install.sh --offline

# Option 2: Manual specification
export OFFLINE_INSTALL=true
export OPENEBS_CHART_DIR="./offline-media/charts/lvm-localpv-1.8.0"
export OPENEBS_IMAGE_REGISTRY="your-registry.example.com"  # If using private registry
./install.sh --offline
```

### Automatic Media Detection

The install and upgrade scripts can automatically detect the offline media directory if it's in a standard location:

- `./offline-media/`
- `../offline-media/`
- `deploy/../offline-media/`

If detected, the scripts will:

- Automatically extract and use the chart from `offline-media/charts/`
- Optionally load images if `OPENEBS_LOAD_IMAGES=true` is set

## Additional Resources

- [Quickstart Guide](./quickstart.md) - Basic installation using Helm commands
- [StorageClass Documentation](./storageclasses.md) - StorageClass parameters and configuration
- [Upgrade Guide](./upgrade.md) - General upgrade information
- [FAQ](./faq.md) - Frequently asked questions

## Support

For issues, questions, or contributions:

- **GitHub Issues**: https://github.com/openebs/lvm-localpv/issues
- **Documentation**: https://openebs.io/docs
- **Community**: https://openebs.io/community
