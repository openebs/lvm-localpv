## PersistentVolumeClaim Conformance matrix

Following matrix shows supported PersistentVolumeClaim parameters for lvm-localpv.

| PersistentVolumeClaim Parameters | LVM CSI Driver | E2E Coverage |
| -------------------------------- | -------------- | ------------ |
| [AccessMode](#accessmode) <br> Supported access modes are <li> ReadWriteOnce </li> </br> | Supported | [Yes](https://github.com/openebs/lvm-localpv/tree/master/e2e-tests/experiments/lvm-localpv-provisioner#readme) |
| [Storageclass](#storageclassname) | Supported | [Yes](https://github.com/openebs/lvm-localpv/tree/master/e2e-tests/experiments/lvm-localpv-provisioner#readme) |
| [Capacity Resource](#capacity-resource) | Supported | Yes |
| [VolumeMode](#volumemode) <br> Supported volume modes are <li> Block </li> <li> Filesystem </li> </br> | Supported | Yes <br> Test cases available for Filesystem mode </br>|
| [Selectors](#selectors)   | Supported | Pending |
| [VolumeName](#volumename) | Supported | Pending | 

## PersistentVolumeClaim Parameters

### AccessMode

LVM-LocalPV supports only `ReadWriteOnce` access mode i.e volume can be mounted as read-write by a single node.
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

LVM CSI-Driver supports dynamic provision of volume for the PVCs refered to lvm storageclass.

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

Admin/User can specify the desired capacity for lvm volume. Once the CSI-Driver gets request it will check whether available space in underlying volume group. If it has enough space a success respone will be returned to caller else error will be reported.

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

### VolumeMode

LVM-LocalPV supports two kind of volume modes(Defaults to Filesystem):
- Block  (Block mode can be used in a case where application itself maintains filesystem)
- Filesystem (Application which requires filesystem as a prerequisite)

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


### Selectors

Users can bind any of retained lvm volumes to new PersistentVolumeClaim object via selector field.
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
      storage: 4Gi
```
- Verify bound status of PV
```sh
$ kubectl get pv
NAME                                       CAPACITY   ACCESS MODES   RECLAIM POLICY   STATUS   CLAIM               STORAGECLASS    REASON   AGE
pvc-8376b776-75f9-4786-8311-f8780adfabdb   6Gi        RWO            Retain           Bound    default/csi-lvmpv   openebs-lvmpv   9h
```

### VolumeName

VolumeName can be used to bind PersistentVolumeClaim(PVC) to retained PersistentVolume(PV). When VolumeName is specified K8s will ignore selector field.
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
      storage: 4Gi
```
