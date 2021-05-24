## StorageClass Parameters Conformance matrix

Following matrix shows supported storageclass parameters for lvm-localpv

| StorageClass Parameters | LVM CSI Driver | E2E Coverage |
| ----------------------- | -------------- | ------------ |
|   [Passing Secrets](https://kubernetes-csi.github.io/docs/secrets-and-credentials-storage-class.html#examples) | No Use Case | NA |
|   [fsType](#fstype-optional) <br>(supports ext2, ext3, ext4, xfs & btrfs filesystem) </br> | Supported | [Yes](https://github.com/openebs/lvm-localpv/tree/master/e2e-tests/experiments/functional/lvm-controller-high-availability#readme) <br> (Test coverage exist for ext4 & xfs)</br> |
|   [allowVolumeExpansion](#allowvolumeexpansion-optional) <br>(Supports expansion on ext2, ext3, ext4 & xfs) </br> | Supported | [Yes](https://github.com/openebs/lvm-localpv/tree/master/e2e-tests/experiments/functional/lvm-volume-resize#about-this-experiment) <br> (Test coverage exist for ext4 & xfs) </br> |
| Reclaim Policy <br> (Supports Retain & Delete reclaim policy) </br> | Supported  | Yes <br> (Test coverage exist for Delete reclaim policy) </br> |
| [MountOptions](#mountoptions-optional) | Supported | Pending |
| [VolumeBindingMode](#volumebindingmode-optional) <br> (Supports  Immediate & WaitForFirstConsumer modes) | Supported | Yes |
| [allowedTopologies](#storageclass-with-custom-node-labels) | Supported | [Yes](https://github.com/openebs/lvm-localpv/tree/master/e2e-tests/experiments/functional/lvmpv-custom-topology#readme) |
| [shared](#shared-optional) | Supported | [Yes](https://github.com/openebs/lvm-localpv/tree/master/e2e-tests/experiments/functional/lvmpv-shared-mount#readme) |
| [vgpattern](#vgpattern-must-parameter-if-volgroup-is-not-provided-otherwise-optional)  | Supported | Pending |
| [volgroup](#volgroup-must-parameter-if-vgpattern-is-not-provided-otherwise-optional)   | Supported | [Yes](https://github.com/openebs/lvm-localpv/blob/master/e2e-tests/experiments/lvm-localpv-provisioner/openebs-lvmsc.j2) |
| [thinProvision](#thinprovision-optional-parameter) | Supported | Pending |

## StorageClass Parameters


### FsType (Optional)

Admin can specify filesystem in storageclass. lvm-localpv CSI-Driver will format block device with specified filesystem and mount in application pod. If fsType is not specified defaults to "ext4" filesystem.
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvm
allowVolumeExpansion: true
provisioner: local.csi.openebs.io
parameters:
  storage: "lvm"
  vgpattern: "lvmvg.*"
  fsType: xfs               ## Supported filesystems are ext2, ext3, ext4, xfs & btrfs
```

### AllowVolumeExpansion (Optional)

Users can expand the volumes when the underlying StorageClass has `allowVolumeExpansion` field set to true.
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvm
allowVolumeExpansion: true     ## If set to true then dynamically it allows expansion of volume
provisioner: local.csi.openebs.io
parameters:
  storage: "lvm"
  vgpattern: "lvmvg.*"
```

### MountOptions (Optional)

Volumes that are provisioned via lvm-localpv will use the mount options specified in storageclass during volume mounting time.
**Note**: Mount options are not validated. If mount options are invalid, then volume mount fails.
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvm
provisioner: local.csi.openebs.io
parameters:
  storage: "lvm"
  vgpattern: "lvmvg.*"
  mountOptions:    ## Various mount options of volume can be specified here
    - debug
```

### VolumeBindingMode (Optional)

lvm-localpv supports two type volume binding modes that are `Immediate` & `late binding`.
- Immediate: Indicates that volume binding and dynamic provisioning occurs once the PersistentVolumeClaim is created.
- WaitForFirstConsumer: It is also known as late binding which will delay binding & provisioning of a PersistentVolumeClaim until a pod using the PersistentVolumeClaim is created.
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvm
provisioner: local.csi.openebs.io
parameters:
  storage: "lvm"
  vgpattern: "lvmvg.*"
volumeBindingMode: WaitForFirstConsumer     ## It can also accepts Immediate binding mode
```

### volgroup (*must* parameter if vgpattern is not provided, otherwise optional)

volgroup specifies the name of the volume group on nodes from which the volumes will be created. The *volgroup* is the must argument if `vgpattern` is not provided in the storageclass.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvm
provisioner: local.csi.openebs.io
parameters:
  storage: "lvm"
  volgroup: "lvmvg"       ## volgroup specifies name of lvm volume group
```

### vgpattern (*must* parameter if volgroup is not provided, otherwise optional)

vgpattern specifies the regular expression for the volume groups on node from which the volumes can be created. The *vgpattern* is the must argument if `volgroup` parameter is not provided in the storageclass. Here, in this case the driver will pick the volume groups matching the vgpattern with enough free capacity to accomodate the volume and will use the one which has largest capacity available for provisioning the volume.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvm
provisioner: local.csi.openebs.io
parameters:
  storage: "lvm"
  vgpattern: "lvmvg.*"     ## vgpattern specifies pattern of lvm volume group name
```

if `volgroup` and `vgpattern` both the paramaters are defined in the storageclass then `volgroup` will get higher priority and the driver will use that to provision to the volume.

**Note:** Please note that either volgroup or vgpattern should be present in the storageclass parameters to make the provisioning successful.

### thinProvision (*optional* parameter)

For creating thin-provisioned volume, use thinProvision parameter in storage class. It's allowed values are: "yes" and "no". If we don't use this parameter by default it's value will be "no" and it will work as thick provisioned volumes.

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvm
provisioner: local.csi.openebs.io
parameters:
  storage: "lvm"
  volgroup: "lvmvg"
  thinProvision: "yes"
```
Before creating thin provision volume, make ensure that required thin provisioning kernel module `dm_thin_pool` is loaded on all the nodes.

To verify if the modules are loaded, run:
```
lsmod | grep dm_thin_pool
```

If modules are not loaded, then execute the following command to load the modules:
```
modprobe dm_thin_pool
```

### StorageClass With Custom Node Labels

There can be a use case where we have certain kinds of Volume Groups present on certain nodes only, and we want a particular type of application to use that VG. We can create a storage class with `allowedTopologies` and mention all the nodes there where that vg is present:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
 name: lvm-sc
allowVolumeExpansion: true
parameters:
  volgroup: "lvmvg"
provisioner: local.csi.openebs.io
allowedTopologies:
- matchLabelExpressions:
 - key: openebs.io/nodename
   values:
     - node-1
     - node-2
```

Here we can have volume group of name “lvmvg” created on the nvme disks and want to use this high performing LVM volume group for the applications that need higher IOPS. We can use the above SorageClass to create the PVC and deploy the application using that.

The LVM-LocalPV driver will create the Volume in the volume group “lvmvg” present on the node with fewer of volumes provisioned among the given node list. In the above StorageClass, if there provisioned volumes on node-1 are less, it will create the volume on node-1 only. Alternatively, we can use `volumeBindingMode: WaitForFirstConsumer` to let the k8s select the node where the volume should be provisioned.

The problem with the above StorageClass is that it works fine if the number of nodes is less, but if the number of nodes is huge, it is cumbersome to list all the nodes like this. In that case, what we can do is, we can label all the similar nodes using the same key value and use that label to create the StorageClass.

```
user@k8s-master:~ $ kubectl label node k8s-node-2 openebs.io/lvmvg=nvme
node/k8s-node-2 labeled
user@k8s-master:~ $ kubectl label node k8s-node-1 openebs.io/lvmvg=nvme
node/k8s-node-1 labeled
```

Now, restart the LVM-LocalPV Driver (if already deployed, otherwise please ignore) so that it can pick the new node label as the supported topology. Check [faq](./faq.md#1-how-to-add-custom-topology-key) for more details.

```
$ kubectl delete po -n kube-system -l role=openebs-lvm
```

Now, we can create the StorageClass like this:

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
 name: nvme-lvmsc
allowVolumeExpansion: true
parameters:
 volgroup: "lvmvg"
provisioner: local.csi.openebs.io
allowedTopologies:
- matchLabelExpressions:
 - key: openebs.io/lvmvg
   values:
     - nvme
```

Here, the volumes will be provisioned on the nodes which has label “openebs.io/lvmvg” set as “nvme”.


### Shared (Optional)

lvm-localpv volume mount point can be shared among the multiple pods on the same node. Applications that can share the volume can set value of `shared` parameter to true.
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
 name: nvme-lvmsc
allowVolumeExpansion: true
parameters:
 volgroup: "lvmvg"
 shared: "yes"
provisioner: local.csi.openebs.io
```
