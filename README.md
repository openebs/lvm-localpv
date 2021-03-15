# OpenEBS LVM CSI Driver
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/3523/badge)](https://bestpractices.coreinfrastructure.org/en/projects/4548)
[![Slack](https://img.shields.io/badge/chat!!!-slack-ff1493.svg?style=flat-square)](https://openebsslacksignup.herokuapp.com/)
[![Community Meetings](https://img.shields.io/badge/Community-Meetings-blue)](https://hackmd.io/yJb407JWRyiwLU-XDndOLA?view)
[![Go Report](https://goreportcard.com/badge/github.com/openebs/lvm-localpv)](https://goreportcard.com/report/github.com/openebs/lvm-localpv)

<img width="300" align="right" alt="OpenEBS Logo" src="https://raw.githubusercontent.com/cncf/artwork/master/projects/openebs/stacked/color/openebs-stacked-color.png" xmlns="http://www.w3.org/1999/html">

CSI driver for provisioning Local PVs backed by LVM and more.

## Project Status

Currently the LVM CSI Driver is in alpha.

## Usage

### Prerequisites

Before installing LVM driver please make sure your Kubernetes Cluster
must meet the following prerequisites:

1. all the nodes must have lvm2 utils installed
2. volume group has been setup for provisioning the volume
3. You have access to install RBAC components into kube-system namespace.
   The OpenEBS LVM driver components are installed in kube-system namespace
   to allow them to be flagged as system critical components.

### Supported System

K8S : 1.17+

OS : Ubuntu

LVM : 2

### Setup

Find the disk which you want to use for the LVM, for testing you can use the loopback device

```
truncate -s 1024G /tmp/disk.img
sudo losetup -f /tmp/disk.img --show
```

Create the Volume group on all the nodes, which will be used by the LVM Driver for provisioning the volumes

```
sudo pvcreate /dev/loop0
sudo vgcreate lvmvg /dev/loop0
```

### Installation

Deploy the Operator yaml 

```
kubectl apply -f https://raw.githubusercontent.com/openebs/lvm-localpv/master/deploy/lvm-operator.yaml
```

### Deployment

deploy the sample fio application 

```
kubectl apply -f https://raw.githubusercontent.com/openebs/lvm-localpv/master/deploy/sample/fio.yaml
```

Features
---

- [x] Access Modes
    - [x] ReadWriteOnce
    - ~~ReadOnlyMany~~
    - ~~ReadWriteMany~~
- [x] Volume modes
    - [x] `Filesystem` mode
    - [ ] `Block` mode
- [ ] Supports fsTypes: `ext4`, `btrfs`, `xfs`
- [x] Volume metrics
- [x] Topology
- [x] [Snapshot](docs/snapshot.md)
- [ ] Clone
- [x] Volume Resize
- [ ] Backup/Restore
- [ ] Ephemeral inline volume

### Limitation
- Resize of volumes with snapshot is not supported
