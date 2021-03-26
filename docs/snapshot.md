## Snapshot

The LVM driver supports creating snapshots of the LVM volumes. Certain settings are applied by the LVM driver which modifies the default behaviour of LVM snapshot:

- Snapshots created by LVM driver are ReadOnly by default as opposed to the ReadWrite snapshots created by default by `lvcreate` command
- The size of snapshot will be set to the size of the origin volume

### Steps to create a snapshot
1. A SnapshotClass needs to be created. A sample SnapshotClass can be found [here](https://github.com/openebs/lvm-localpv/blob/master/deploy/sample/lvmsnapclass.yaml).
```yaml
kind: VolumeSnapshotClass
apiVersion: snapshot.storage.k8s.io/v1
metadata:
  name: lvmpv-snapclass
  annotations:
    snapshot.storage.kubernetes.io/is-default-class: "true"
driver: lvm.csi.openebs.io
deletionPolicy: Delete
```

Apply the SnapshotClass YAML:
```bash
$ kubectl apply -f snapshotclass.yaml
volumesnapshotclass.snapshot.storage.k8s.io/lvmpv-snapclass created
```

2. Find a PVC for which snapshot has to be created
```bash
$ kubectl get pvc
NAME         STATUS   VOLUME                                     CAPACITY   ACCESS MODES   STORAGECLASS    AGE
csi-lvmpvc   Bound    pvc-c7f42430-f2bb-4459-9182-f76b8896c532   4Gi        RWO            openebs-lvmsc   53s
```

3. Create the snapshot using the created SnapshotClass for the selected PVC
```yaml
apiVersion: snapshot.storage.k8s.io/v1
kind: VolumeSnapshot
metadata:
  name: lvm-localpv-snap
spec:
  volumeSnapshotClassName: lvmpv-snapclass
  source:
    persistentVolumeClaimName: csi-lvmpvc
```

Apply the Snapshot YAML
```bash
$ kubectl apply -f lvmsnapshot.yaml
volumesnapshot.snapshot.storage.k8s.io/lvm-localpv-snap created
```

4. Please note that you have to create the snapshot in the same namespace where the PVC is created. Check the created snapshot resource, make sure readyToUsefield is true, before using this snapshot for any purpose. 
```bash
$ kubectl get volumesnapshot
NAME               READYTOUSE   SOURCEPVC    SOURCESNAPSHOTCONTENT   RESTORESIZE   SNAPSHOTCLASS     SNAPSHOTCONTENT                                    CREATIONTIME   AGE
lvm-localpv-snap   true         csi-lvmpvc                           0             lvmpv-snapclass   snapcontent-f771db56-1cef-43d1-ac88-d0e789d4b718   15s            15s
```

5. Check the OpenEBS resource for the created snapshot and make sure the status is `Ready`
```bash
$ kubectl get lvmsnapshot -n openebs
NAME                                            AGE
snapshot-f771db56-1cef-43d1-ac88-d0e789d4b718   3m12s
```

```bash
$ kubectl get lvmsnapshot -n openebs -o yaml
apiVersion: local.openebs.io/v1alpha1
kind: LVMSnapshot
metadata:
  creationTimestamp: "2021-03-15T08:36:21Z"
  finalizers:
  - lvm.openebs.io/finalizer
  generation: 2
  labels:
    kubernetes.io/nodename: worker-ak1
    openebs.io/persistent-volume: pvc-7d27935e-c72a-4f6b-8314-96ee600e01e8
  name: snapshot-f771db56-1cef-43d1-ac88-d0e789d4b718
  namespace: openebs
  resourceVersion: "95576717"
  selfLink: /apis/local.openebs.io/v1alpha1/namespaces/openebs/lvmsnapshots/snapshot-f771db56-1cef-43d1-ac88-d0e789d4b718
  uid: 96f3f2e4-93aa-4d25-9611-169099ce40a8
spec:
  capacity: "4294967296"
  ownerNodeID: worker-ak1
  shared: "no"
  volGroup: lvmvg
status:
  state: Ready
```

To confirm that snapshot has been created, ssh into the node and check for lvm volumes
```bash
$ lvs
  LV                                       VG    Attr       LSize Pool Origin                                   Data%  Meta%  Move Log Cpy%Sync Convert
  f771db56-1cef-43d1-ac88-d0e789d4b718     lvmvg sri-a-s--- 4.00g      pvc-7d27935e-c72a-4f6b-8314-96ee600e01e8 0.00                                   
  pvc-7d27935e-c72a-4f6b-8314-96ee600e01e8 lvmvg owi-aos--- 4.00g                                                                                      
```

### Limitations

Resize is not supported for volumes that have a snapshot. This is not an LVM limitation, but is intentionally done from the LVM driver, since LVM does not automatically resize the snapshots when origin volume is resized.
