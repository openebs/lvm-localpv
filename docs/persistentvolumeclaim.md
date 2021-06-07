## PersistentVolumeClaim Conformance matrix

Following matrix shows supported PersistentVolumeClaim parameters for lvm-localpv.

<table>
  <thead>
    <tr>
      <th> Parameter </th>
      <th> Values </th>
      <th> Development Status </th>
      <th> E2E Coverage Status </th>
    </tr>
  </thead>
  <tbody>
    <tr>
      <td rowspan=3> <a href="#accessmode"> AccessMode </a> </td>
      <td> ReadWriteOnce </td>
      <td> Supported </td>
      <td rowspan=3> <a href="https://github.com/openebs/lvm-localpv/tree/master/e2e-tests/experiments/lvm-localpv-provisioner#readme"> Yes </a> </td>
    </tr>
    <tr>
      <td> <strike> ReadWriteMany </strkie> </td>
      <td> Not Supported </td>
    </tr>
    <tr>
      <td> <strike> ReadOnlyMany </strike> </td>
      <td> Not Supported </td>
    </tr>
    <tr>
      <td> <a href="#storageclassname"> Storageclass </td>
      <td> StorageClassName </td>
      <td> Supported </td>
      <td> <a href="https://github.com/openebs/lvm-localpv/tree/master/e2e-tests/experiments/lvm-localpv-provisioner#readme"> Yes </a> </td>
    </tr>
    <tr>
      <td> <a href="#capacity-resource"> Capacity Resource </a> </td>
      <td> Number along with size unit </td>
      <td> Supported </td>
      <td> <a href="https://github.com/openebs/lvm-localpv/tree/master/e2e-tests/experiments/functional/lvm-volume-resize#readme"> Yes </a> </td>
    </tr>
    <tr>
      <td rowspan=2> <a href="#volumemode-optional"> VolumeMode </a> </td>
      <td> Block </td>
      <td> Supported </td>
      <td rowspan=2> <a href="https://github.com/openebs/lvm-localpv/blob/master/e2e-tests/apps/percona/deployers/run_e2e_test.yml"> Yes </a> <br> <i> Test cases available for Filesystem mode </i> </br> </td>
    </tr>
    <tr>
      <td> Filesystem </td>
      <td> Supported </td>
    </tr>
    <tr>
      <td> <a href="#selectors-optional"> Selectors </a> </td>
      <td> Equality & Set based selections </td>
      <td> Supported </td>
      <td> Pending </td>
    </tr>
    <tr>
      <td> <a href="#volumename-optional"> VolumeName </a> </td>
      <td> Available PV name </td>
      <td> Supported </td>
      <td> Pending </td>
    </tr>
    <tr>
      <td> DataSource </td>
      <td> - </td>
      <td> Not Supported </td>
      <td> Pending </td>
    </tr>
  </tbody>
</table>


## PersistentVolumeClaim Parameters

### AccessMode

LVM-LocalPV supports only `ReadWriteOnce` access mode i.e volume can be mounted as read-write by a single node. AccessMode is a required field, if the field is unspecified then it will lead to a creation error.
```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: csi-lvmpv
spec:
  accessModes:
    - ReadWriteOnce        ## Specify ReadWriteOnce(RWO) access modes
  storageClassName: openebs-lvm
  resources:
    requests:
      storage: 4Gi
```

### StorageClassName

LVM CSI-Driver supports dynamic provision of volume for the PVCs referred to lvm storageclass. StorageClassName is a required field, if field is unspecified then it will lead to a provision errors.

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: csi-lvmpv
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: openebs-lvm    ## It must be OpenEBS LVM storageclass for provisioning LVM volumes
  resources:
    requests:
      storage: 4Gi
```

### Capacity Resource

Admin/User can specify the desired capacity for lvm volume. CSI-Driver will provision a volume if the underlying volume group has requested capacity available else provisioning volume will be errored. StorageClassName is a required field, if field is unspecified then it will lead to a provision errors.

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: csi-lvmpv
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: openebs-lvm
  resources:
    requests:
      storage: 4Gi       ## Specify required storage for an application
```

### VolumeMode (Optional)

LVM-LocalPV supports two kind of volume modes(Defaults to Filesystem):
- Block  (Block mode can be used in a case where application itself maintains filesystem)
- Filesystem (Application which requires filesystem as a prerequisite)
Note: By default K8s will take volume mode as **Filesystem**.

```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: csi-lvmpv
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: openebs-lvm
  volumeMode: Filesystem     ## Specifies in which mode volume should be attached to pod
  resources:
    requests:
      storage: 4Gi
```


### Selectors (Optional)

Users can bind any of retained lvm volumes to new PersistentVolumeClaim object via selector field. If selector and [volumeName](#volumename-optional) fields are unspecified then LVM CSI driver will provision new volume.
Follow below steps to specify selector on PersistentVolumeClaim:

- List the persistentvolumes(PVs) which has status Released.
```sh
$ kubectl get pv -ojsonpath='{range .items[?(@.status.phase=="Released")]}{.metadata.name} {.metadata.labels}{"\n"}'

pvc-8376b776-75f9-4786-8311-f8780adfabdb {"openebs.io/lvm-volume":"reuse"}
```
**Note**: If labels doesn't exist for persistent volume then it is required to add labels to PV
```sh
$ kubectl label pv pvc-8376b776-75f9-4786-8311-f8780adfabdb openebs.io/lvm-volume=reuse
```

- Remove the claimRef on selected persistentvolumes using patch command(This will mark PV as `Available` for binding).
```sh
$ kubectl patch pv pvc-8376b776-75f9-4786-8311-f8780adfabdb -p '{"spec":{"claimRef": null}}'

persistentvolume/pvc-8376b776-75f9-4786-8311-f8780adfabdb patched
```
- Create pvc with the selector
```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: csi-lvmpv
spec:
  storageClassName: openebs-lvmpv
  ## Specify selector matching to available PVs label, K8s will bound to any of available PV matches to specified labels
  selector:
    matchLabels:
      openebs.io/lvm-volume: reuse
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 4Gi   ## Capacity should be less than or equal to available PV capacities
```
- Verify bound status of PV
```sh
$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM               STORAGECLASS    REASON   AGE
pvc-8376b776-75f9-4786-8311-f8780adfabdb   6Gi        RWO            Retain           Bound    default/csi-lvmpv   openebs-lvmpv   9h
```

### VolumeName (Optional)

VolumeName can be used to bind PersistentVolumeClaim(PVC) to retained PersistentVolume(PV). When VolumeName is specified K8s will ignore [selector field](#selectors-optional). If volumeName is unspecified then CSI driver will try to provision new volume.
Note: Before creating PVC make PersistentVolume `Available` by removing claimRef on PersistentVolume.
```yaml
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: csi-lvmpv
spec:
  storageClassName: openebs-lvmpv
  volumeName: pvc-8376b776-75f9-4786-8311-f8780adfabdb   ## Name of LVM volume present in Available state
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 4Gi  ## Capacity should be less than or equal to available PV capacities
```
