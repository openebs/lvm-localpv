## Parameters

### volgroup (*must* parameter if vgpattern is not provided, otherwise optional)

volgroup specifies the name of the volume group on nodes from which the volumes will be created. The *volgroup* is the must argument if `vgpattern` is not provided in the storageclass.

```
volgroup: "lvmvg"
```

### vgpattern (*must* parameter if volgroup is not provided, otherwise optional)

vgpattern specifies the regular expression for the volume groups on node from which the volumes can be created. The *vgpattern* is the must argument if `volgroup` parameter is not provided in the storageclass. Here, in this case the driver will pick the volume groups matching the vgpattern with enough free capacity to accomodate the volume and will use the one which has largest capacity available for provisioning the volume.

```
vgpattern: "lvmvg.*"
```

if `volgroup` and `vgpattern` both the paramaters are defined in the storageclass then `volgroup` will get higher priority and the driver will use that to provision to the volume.

**Note:** Please note that either volgroup or vgpattern should be present in the storageclass parameters to make the provisioning successful.

### thinProvision (*optional* parameter)

For creating thin-provisioned volume, use thinProvision parameter in storage class. It's allowed values are: "yes" and "no". If we don't use this parameter by default it's value will be "no" and it will work as thick provisioned volumes.

```
parameters:
  volgroup: "lvmvg"
  thinProvision: "yes"
provisioner: local.csi.openebs.io
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
 volumegroup: "lvmvg"
provisioner: local.csi.openebs.io
allowedTopologies:
- matchLabelExpressions:
 - key: openebs.io/lvmvg
   values:
     - nvme
```

Here, the volumes will be provisioned on the nodes which has label “openebs.io/lvmvg” set as “nvme”.
