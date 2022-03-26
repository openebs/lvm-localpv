# OpenEBS LVM CSI Driver
[![FOSSA Status](https://app.fossa.io/api/projects/git%2Bgithub.com%2Fopenebs%2Flvm-localpv.svg?type=shield)](https://app.fossa.io/projects/git%2Bgithub.com%2Fopenebs%2Flvm-localpv?ref=badge_shield)
[![CII Best Practices](https://bestpractices.coreinfrastructure.org/projects/3523/badge)](https://bestpractices.coreinfrastructure.org/en/projects/4548)
[![Slack](https://img.shields.io/badge/chat!!!-slack-ff1493.svg?style=flat-square)](https://kubernetes.slack.com/messages/openebs)
[![Community Meetings](https://img.shields.io/badge/Community-Meetings-blue)](https://hackmd.io/yJb407JWRyiwLU-XDndOLA?view)
[![Go Report](https://goreportcard.com/badge/github.com/openebs/lvm-localpv)](https://goreportcard.com/report/github.com/openebs/lvm-localpv)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopenebs%2Flvm-localpv.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopenebs%2Flvm-localpv?ref=badge_shield)

<img width="300" align="right" alt="OpenEBS Logo" src="https://raw.githubusercontent.com/cncf/artwork/master/projects/openebs/stacked/color/openebs-stacked-color.png" xmlns="http://www.w3.org/1999/html">

CSI driver for provisioning Local PVs backed by LVM and more.

## Project Status

LVM-LocalPV CSI Driver is declared GA in August 2021 with the release version as 0.8.0.

## Project Tracker

See [roadmap](https://github.com/orgs/openebs/projects/30).

## Usage

### Prerequisites

Before installing LVM driver please make sure your Kubernetes Cluster
must meet the following prerequisites:

1. all the nodes must have lvm2 utils installed and the dm-snapshot kernel module loaded
2. volume group has been setup for provisioning the volume
3. You have access to install RBAC components into kube-system namespace.
   The OpenEBS LVM driver components are installed in kube-system namespace
   to allow them to be flagged as system critical components.

### Supported System

K8S : 1.20+

OS : Ubuntu

LVM version : LVM 2

### Setup

Find the disk which you want to use for the LVM, for testing you can use the loopback device

```
truncate -s 1024G /tmp/disk.img
sudo losetup -f /tmp/disk.img --show
```

Create the Volume group on all the nodes, which will be used by the LVM Driver for provisioning the volumes

```
sudo pvcreate /dev/loop0
sudo vgcreate lvmvg /dev/loop0       ## here lvmvg is the volume group name to be created
```

### Installation

We can install the latest release of OpenEBS LVM driver by running the following command.

```
$ kubectl apply -f https://openebs.github.io/charts/lvm-operator.yaml
```

If you want to fetch a versioned manifest, you can use the manifests for a
specific OpenEBS release version, for example:

```
$ kubectl apply -f https://raw.githubusercontent.com/openebs/charts/gh-pages/versioned/3.0.0/lvm-operator.yaml
```

**NOTE:** For some Kubernetes distributions, the `kubelet` directory must be changed at all relevant places in the YAML powering the operator (both the `openebs-lvm-controller` and `openebs-lvm-node`).

- For `microk8s`, we need to change the kubelet directory to `/var/snap/microk8s/common/var/lib/kubelet/`, we need to replace `/var/lib/kubelet/` with `/var/snap/microk8s/common/var/lib/kubelet/` at all the places in the operator yaml and then we can apply it on microk8s.

- For `k0s`, the default directory (`/var/lib/kubelet`) should be changed to `/var/lib/k0s/kubelet`.

- For `RancherOS`, the default directory (`/var/lib/kubelet`) should be changed to `/opt/rke/var/lib/kubelet`.

Verify that the LVM driver Components are installed and running using below command :

```
$ kubectl get pods -n kube-system -l role=openebs-lvm
```

Depending on number of nodes, you will see one lvm-controller pod and lvm-node daemonset running
on the nodes.

```
NAME                       READY   STATUS    RESTARTS   AGE
openebs-lvm-controller-0   5/5     Running   0          35s
openebs-lvm-node-54slv     2/2     Running   0          35s
openebs-lvm-node-9vg28     2/2     Running   0          35s
openebs-lvm-node-qbv57     2/2     Running   0          35s

```
Once LVM driver is successfully installed, we can provision volumes.

### Deployment


#### 1. Create a Storage class

```
$ cat sc.yaml

apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvmpv
parameters:
  storage: "lvm"
  volgroup: "lvmvg"
provisioner: local.csi.openebs.io
```

Check the doc on [storageclasses](docs/storageclasses.md) to know all the supported parameters for LVM-LocalPV

##### VolumeGroup Availability

If LVM volume group is available on certain nodes only, then make use of topology to tell the list of nodes where we have the volgroup available.
As shown in the below storage class, we can use allowedTopologies to describe volume group availability on nodes.

```
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvmpv
allowVolumeExpansion: true
parameters:
  storage: "lvm"
  volgroup: "lvmvg"
provisioner: local.csi.openebs.io
allowedTopologies:
- matchLabelExpressions:
  - key: kubernetes.io/hostname
    values:
      - lvmpv-node1
      - lvmpv-node2
```

The above storage class tells that volume group "lvmvg" is available on nodes lvmpv-node1 and lvmpv-node2 only. The LVM driver will create volumes on those nodes only.

Please note that the provisioner name for LVM driver is "local.csi.openebs.io", we have to use this while creating the storage class so that the volume provisioning/deprovisioning request can come to LVM driver.

#### 2. Create the PVC

```
$ cat pvc.yaml

kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: csi-lvmpv
spec:
  storageClassName: openebs-lvmpv
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 4Gi
```

Create a PVC using the storage class created for the LVM driver.

#### 3. Deploy the application

Create the deployment yaml using the pvc backed by LVM storage.

```
$ cat fio.yaml

apiVersion: v1
kind: Pod
metadata:
  name: fio
spec:
  restartPolicy: Never
  containers:
  - name: perfrunner
    image: openebs/tests-fio
    command: ["/bin/bash"]
    args: ["-c", "while true ;do sleep 50; done"]
    volumeMounts:
       - mountPath: /datadir
         name: fio-vol
    tty: true
  volumes:
  - name: fio-vol
    persistentVolumeClaim:
      claimName: csi-lvmpv
```

After the deployment of the application, we can go to the node and see that the lvm volume is being used
by the application for reading/writting the data and space is consumed from the LVM. Please note that to check the provisioned volumes on the node, we need to run `pvscan --cache` command to update the lvm cache and then we can use lvdisplay and all other lvm commands on the node.

#### 4. Deprovisioning

for deprovisioning the volume we can delete the application which is using the volume and then we can go ahead and delete the pv, as part of deletion of pv this volume will also be deleted from the volume group and data will be freed.

```
$ kubectl delete -f fio.yaml
pod "fio" deleted
$ kubectl delete -f pvc.yaml
persistentvolumeclaim "csi-lvmpv" deleted
```

Features
---

- [x] Access Modes
    - [x] ReadWriteOnce
    - ~~ReadOnlyMany~~
    - ~~ReadWriteMany~~
- [x] Volume modes
    - [x] `Filesystem` mode
    - [x] [`Block`](docs/raw-block-volume.md) mode
- [x] Supports fsTypes: `ext4`, `btrfs`, `xfs`
- [x] Volume metrics
- [x] Topology
- [x] [Snapshot](docs/snapshot.md)
- [ ] Clone
- [x] [Volume Resize](docs/resize.md)
- [x] [Thin Provision](docs/thin_provision.md)
- [ ] Backup/Restore
- [ ] Ephemeral inline volume

### Limitation
- Resize of volumes with snapshot is not supported


## License
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fopenebs%2Flvm-localpv.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2Fopenebs%2Flvm-localpv?ref=badge_large)

## Repobeats

![Alt](https://repobeats.axiom.co/api/embed/baab8c2a9d1606494ab32714cbf91b65845a6001.svg "Repobeats analytics image")
