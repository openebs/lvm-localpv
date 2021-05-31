## StorageClass Parameters Conformance matrix

Following matrix shows standard storageclass parameters for lvm-localpv

### Standard StorageClass Parameters

<table>
  <tr>
    <th> Parameter </th>
    <th colspan=2> Values </th>
    <th> Development Status </th>
    <th> E2E Coverage </th>
  </tr>
  
  <tr>
    <td rowspan=2> <a href="#allowvolumeexpansion-optional"> allowVolumeExpansion </a> </td>
    <td> true </td>
    <td></td>
    <td> Supported </td>
    <td rowspan=2> <a href="https://github.com/openebs/lvm-localpv/tree/master/e2e-tests/experiments/functional/lvm-volume-resize#about-this-experiment"> Yes </a> <br><i> (Test coverage exist for ext4 & xfs) </i></br> </td>
  </tr>
  <tr>
    <td> false </td>
    <td></td>
    <td> Supported </td>
  </tr>

  <tr>
    <td> <a href="#mountoptions-optional"> MountOptions </a> </td>
    <td> Options supported by filesystem </td>
    <td></td>
    <td> Supported </td>
    <td> Pending </td>
  </tr>

  <tr>
    <td rowspan=2> <a href="#volumebindingmode-optional"> VolumeBindingMode </a> </td>
    <td> Immediate </td>
    <td></td>
    <td> Supported </td>
    <td rowspan=2> Yes </td>
  </tr>
  <tr>
    <td> WaitForFirstConsumer </td>
    <td></td>
    <td> Supported </td>
  </tr>

  <tr>
    <td rowspan=2> Reclaim Policy </td>
    <td> Retain </td>
    <td></td>
    <td> Supported </td>
    <td rowspan=2> Yes <br> <i> (Test coverage exist for Delete reclaim policy) </i> </br> </td>
  </tr>
  <tr>
    <td> Delete </td>
    <td></td>
    <td> Supported </td>
  </tr>

  <tr>
    <td> <a href="#storageclass-with-custom-node-labels"> allowedTopologies </a> </td>
    <td> - </td>
    <td></td>
    <td> Supported </td>
    <td> <a href="https://github.com/openebs/lvm-localpv/tree/master/e2e-tests/experiments/functional/lvmpv-custom-topology#readme"> Yes </a> </td>
  </tr>

  <tr>
    <td rowspan=6> Parameters </td>
    <td> <a href="https://kubernetes-csi.github.io/docs/secrets-and-credentials-storage-class.html#examples"> Passing Secrets </td>
    <td></td>
    <td> No Use Case </td>
    <td> NA </td>
  </tr>
  <tr>
    <td rowspan=5> <a href="#fstype-optional"> fsType </a> </td>
    <td>ext2</td>
    <td rowspan=5> Supported </td>
    <td rowspan=5> <a href="https://github.com/openebs/lvm-localpv/tree/master/e2e-tests/experiments/functional/lvm-controller-high-availability#readme"> Yes </a> <br> <i> (Test coverage exist for ext4 & xfs) </i></br> </td>
  </tr>
  <tr> <td> ext3 </td> </tr>
  <tr> <td> ext4 </td> </tr>
  <tr> <td> xfs </td> </tr>
  <tr> <td> btrfs </td> </tr>
</table>

### LVM Supported StorageClass Parameters

<table>
  <tr>
    <th> Parameter </th>
    <th colspan=2> Values </th>
    <th> Development Status </th>
    <th> E2E Coverage </th>
  </tr>

  <tr>
    <td rowspan=6> Parameters </td>
    <td> <a href="#shared-optional"> shared </td>
    <td> yes </td>
    <td> Supported </td>
    <td> <a href="https://github.com/openebs/lvm-localpv/tree/master/e2e-tests/experiments/functional/lvmpv-shared-mount#readme"> Yes </a> </td>
  </tr>

  <tr>
    <td> <a href="#vgpattern-must-parameter-if-volgroup-is-not-provided-otherwise-optional"> vgpattern </td>
    <td> Regular expression of volume group name </td>
    <td> Supported </td>
    <td> Pending </td>
  </tr>

  <tr>
    <td> <a href="#volgroup-must-parameter-if-vgpattern-is-not-provided-otherwise-optional"> volgroup </td>
    <td> Name of volume group </td>
    <td> Supported </td>
    <td> <a href="https://github.com/openebs/lvm-localpv/blob/master/e2e-tests/experiments/lvm-localpv-provisioner/openebs-lvmsc.j2"> Yes </a> </td>
  </tr>

  <tr>
    <td> <a href="#thinprovision-optional"> thinProvision </td>
    <td> yes </td>
    <td> Supported </td>
    <td> Pending </td>
  </tr>

</table>


## StorageClass Options

### AllowVolumeExpansion (Optional)

Users can expand the volumes only when `allowVolumeExpansion` field is set to true in storageclass. If a field is unspecified, then volume expansion is not supported.
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

Volumes that are provisioned via LVM-LocalPV will use the mount options specified in storageclass during volume mounting time inside an application. If a field is unspecified, `-o default` option will be added to mount the volume.
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

### Parameters

LVM-LocalPV storageclass supports various parameters for different use cases. Following are the supported parameters

- #### FsType (Optional)

  Admin can specify filesystem in storageclass. LVM-LocalPV CSI-Driver will format block device with specified filesystem and mount in application pod. If fsType is not specified defaults to `ext4` filesystem.
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

- #### Shared (Optional)

  lvm-localpv volume mount point can be shared among the multiple pods on the same node. Applications that can share the volume can set value of `shared` parameter to yes defaults to no.
  ```yaml
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: nvme-lvmsc
  allowVolumeExpansion: true
  provisioner: local.csi.openebs.io
  parameters:
    volgroup: "lvmvg"
    shared: "yes"             ## Parameter that states volume can be shared among multiple pods
  ```

- #### vgpattern (*must* parameter if volgroup is not provided, otherwise optional)

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

- #### volgroup (*must* parameter if vgpattern is not provided, otherwise optional)

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

- #### thinProvision (Optional)

  For creating thin-provisioned volume, use thinProvision parameter in storage class. It's allowed values are: "yes" and "no". If we don't set thinProvision parameter by default it's value will be `no` and it will work as thick provisioned volumes.

  ```yaml
  apiVersion: storage.k8s.io/v1
  kind: StorageClass
  metadata:
    name: openebs-lvm
  provisioner: local.csi.openebs.io
  parameters:
    storage: "lvm"
    volgroup: "lvmvg"
    thinProvision: "yes"      ## Parameter that enables thinprovisioning
  ```
  Before creating thin provision volume, make ensure that required thin provisioning kernel module `dm_thin_pool` is loaded on all the nodes.

  To verify if the modules are loaded, run:
  ```sh
  $ lsmod | grep dm_thin_pool
  ```

  If modules are not loaded, then execute the following command to load the modules:
  ```sh
  $ modprobe dm_thin_pool
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
volumeBindingMode: WaitForFirstConsumer     ## It can also replaced by Immediate volume binding mode depends on use case.
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
