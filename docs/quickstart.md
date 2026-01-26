## Usage and Deployment

## Prerequisites

> Before installing the LocalPV-LVM driver please make sure your Kubernetes Cluster meets the following prerequisites:
> 1. All the nodes must have LVM2 utils package installed
> 2. All the nodes must have dm-snapshot Kernel Module loaded - (Device Mapper Snapshot)

## Setup

Find the disk which you want to use for the LocalPV-LVM.

> **Note**: For testing you can use a loopback device:
> ```bash
> truncate -s 1024G /tmp/disk.img
> sudo losetup -f /tmp/disk.img --show
> ``

> [!NOTE]
> - LocalPV-LVM will not provision the VG for the user <BR>
> - The required Physical Volumes(PV) and Volume Group(VG) names will need to be created and present beforehand.

Create the Volume group on all the nodes, which will be used by the LVM2 Driver for provisioning the volumes

```bash
sudo pvcreate /dev/loop0
sudo vgcreate lvmvg /dev/loop0       ## here lvmvg is the volume group name to be created
```

## Installation

Install the latest release of OpenEBS LVM2 LocalPV-LVM driver by running the following command. Note: All nodes must be running the same version of LocalPV-LVM, LMV2, device-mapper & dm-snapshot.

> **Tip**: For automated installation with node labeling and StorageClass creation, see [Installation Scripts Guide](./installation-scripts.md).

**NOTE:** Installation using operator YAMLs is not the supported way any longer.  
We can install the latest release of OpenEBS LVM driver by running the following command:

```bash
helm repo add openebs-lvmlocalpv https://openebs.github.io/lvm-localpv
helm repo update
helm install openebs-lvmlocalpv openebs-lvmlocalpv/lvm-localpv --namespace openebs --create-namespace
```

**NOTE:** If you are running a custom Kubelet location, or a Kubernetes distribution that uses a custom Kubelet location, the `kubelet` directory must be changed on the helm values at install-time using the flag option `--set lvm-localpv.lvmNode.kubeletDir=<your-directory-path>` in the `helm install` command.

- `microk8s` now symlinks `/var/lib/kubelet/` to its custom directory, so setting the custom value is likely unnecessary. However, you can still replace `/var/lib/kubelet/` with `/var/snap/microk8s/common/var/lib/kubelet/` if desired.
- For `k0s`, the default directory (`/var/lib/kubelet`) should be changed to `/var/lib/k0s/kubelet`.
- For `RancherOS`, the default directory (`/var/lib/kubelet`) should be changed to `/opt/rke/var/lib/kubelet`.

Verify that the LVM driver Components are installed and running using below command. Depending on number of nodes, you will see one lvm-controller pod and lvm-node daemonset running on the nodes :
```bash
$ kubectl get pods -n openebs -l role=openebs-lvm
NAME                                              READY   STATUS    RESTARTS   AGE
openebs-lvm-localpv-controller-7b6d6b4665-fk78q   5/5     Running   0          11m
openebs-lvm-localpv-node-mcch4                    2/2     Running   0          11m
openebs-lvm-localpv-node-pdt88                    2/2     Running   0          11m
openebs-lvm-localpv-node-r9jn2                    2/2     Running   0          11m
```

Once LVM driver is installed and running we can provision a volume.

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

Check the doc on [storageclasses](docs/storageclasses.md) to know all the supported parameters for LocalPV-LVM

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
by the application for reading/writing the data and space is consumed from the LVM. Please note that to check the provisioned volumes on the node, we need to run `pvscan --cache` command to update the lvm cache and then we can use lvdisplay and all other lvm commands on the node.

#### 4. Deprovisioning

for deprovisioning the volume we can delete the application which is using the volume and then we can go ahead and delete the pv, as part of deletion of pv this volume will also be deleted from the volume group and data will be freed.

```
$ kubectl delete -f fio.yaml
pod "fio" deleted
$ kubectl delete -f pvc.yaml
persistentvolumeclaim "csi-lvmpv" deleted
```