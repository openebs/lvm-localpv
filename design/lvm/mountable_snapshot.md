---
title: Snapshot Mount Support for LVM-LocalPV
authors:
  - "@wowditi"
owners:
  - "@kmova"
creation-date: 2022-03-31
last-updated: 2022-03-31
status: Request for Comment
---

# Snapshot Mount Support for LVM-LocalPV

## Table of Contents

* [Table of Contents](#table-of-contents)
* [Summary](#summary)
* [Motivation](#motivation)
    * [Goals](#goals)
    * [Non-Goals](#non-goals)
* [Proposal](#proposal)
* [Implementation Details](#implementation-details)
* [Test Plan](#test-plan)
* [Graduation Criteria](#graduation-criteria)
* [Drawbacks](#drawbacks)

## Summary

LVM Snapshots are space efficient and quick point in time copies of lvm volumes. It consumes the space only when changes are made to the source logical volume. For testing purposes it can be beneficial to start a new application, for example a database, using such a snapshot as the backing storage volume. This allows for creating an exact replica of the application at the moment the snapshot was taken (assuming the state is entirely depedent on the filesystem), which can used to debug issues that occured in production or test how a new version of the application (or an external application) would affect the data.

## Motivation

### Goals

- user should be able to mount snapshots
- user should be able to mount thick and thin snapshots
- user should be able to mount snapshots as read only and copy on write 

### Non-Goals

- Creating clones from Snapshots
- restore of a snapshot


## Proposal

To mount a k8s lvmpv snapshot we need to create a persistent volume claim that references the snapshot as datasource:

```
kind: PersistentVolumeClaim
apiVersion: v1
metadata:
  name: csi-lvmpvc-snap
spec:
  storageClassName: openebs-lvmpv
  dataSource:
    name: lvmpv-snap
    kind: VolumeSnapshot
    apiGroup: snapshot.storage.k8s.io
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
```

Defining resources is required but rather pointless since we also define the size of the snapshot in the snapshotclass. 
Although we could extend the lvm if the defined storage size is larger then that of the snapshot (with a maximum of the size of the original volume).


```
$ kubectl get pvc csi-lvmpvc-snap -o yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  annotations:
    kubectl.kubernetes.io/last-applied-configuration: |
      {"apiVersion":"v1","kind":"PersistentVolumeClaim","metadata":{"annotations":{},"name":"csi-lvmpvc-snap","namespace":"default"},"spec":{"accessModes":["ReadWriteOnce"],"dataSource":{"apiGroup":"snapshot.storage.k8s.io","kind":"VolumeSnapshot","name":"lvmpv-snap"},"resources":{"requests":{"storage":"10Gi"}},"storageClassName":"openebs-lvmpv"}}
    pv.kubernetes.io/bind-completed: "yes"
    pv.kubernetes.io/bound-by-controller: "yes"
    volume.beta.kubernetes.io/storage-provisioner: local.csi.openebs.io
    volume.kubernetes.io/storage-provisioner: local.csi.openebs.io
  creationTimestamp: "2022-03-31T10:37:06Z"
  finalizers:
  - kubernetes.io/pvc-protection
  managedFields:
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          f:pv.kubernetes.io/bind-completed: {}
          f:pv.kubernetes.io/bound-by-controller: {}
          f:volume.beta.kubernetes.io/storage-provisioner: {}
          f:volume.kubernetes.io/storage-provisioner: {}
      f:spec:
        f:volumeName: {}
    manager: kube-controller-manager
    operation: Update
    time: "2022-03-31T10:37:06Z"
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:status:
        f:accessModes: {}
        f:capacity:
          .: {}
          f:storage: {}
        f:phase: {}
    manager: kube-controller-manager
    operation: Update
    subresource: status
    time: "2022-03-31T10:37:06Z"
  - apiVersion: v1
    fieldsType: FieldsV1
    fieldsV1:
      f:metadata:
        f:annotations:
          .: {}
          f:kubectl.kubernetes.io/last-applied-configuration: {}
      f:spec:
        f:accessModes: {}
        f:dataSource: {}
        f:resources:
          f:requests:
            .: {}
            f:storage: {}
        f:storageClassName: {}
        f:volumeMode: {}
    manager: kubectl
    operation: Update
    time: "2022-03-31T10:37:06Z"
  name: csi-lvmpvc-snap
  namespace: default
  resourceVersion: "10994"
  uid: 0217ba5b-0ba2-4e28-af2a-acb97c55e8b1
spec:
  accessModes:
  - ReadWriteOnce
  dataSource:
    apiGroup: snapshot.storage.k8s.io
    kind: VolumeSnapshot
    name: lvmpv-snap
  resources:
    requests:
      storage: 10Gi
  storageClassName: openebs-lvmpv
  volumeMode: Filesystem
  volumeName: pvc-0217ba5b-0ba2-4e28-af2a-acb97c55e8b1
status:
  accessModes:
  - ReadWriteOnce
  capacity:
    storage: 10Gi
  phase: Bound
```

## Implementation Details

In order to implement this only existing code needs to be changed, there is no need for an entirely new control flow.
The changes below are the primary modules that need to be modified, however, there are some additional utility modules that will need to be modified to support these changes. Most likely the following files will need some additions to support snapshots: [kubernetes.go](../../pkg/builder/volbuilder/kubernetes.go) (in order to implement a `Kubeclient.GetSnapshot` function), [lvm_util.go](../../pkg/lvm/lvm_util.go) (in order to create a function that changes the snapshot write access) and [mount.go](../../pkg/lvm/mount.go) (in order to support mounting/unmounting of snapshots).

- In the [controller.go](../../pkg/driver/controller.go) the code path in the `CreateVolume` function for the `controller` type that occurs when contentSource.GetSnapshot() is not `nil` needs to be implemented. When this path is triggered it should return the correct `volName`, `topology` and `cntx`.  
- In the [agent.go](../../pkg/driver/agent.go) the `NodePublishVolume` function for the `node` type needs to be changed such that it checks whether the `volName` is a snapshot and if so mount the snapshot to the specified location.
  - Finding the snapshot that is to be used can be accomplished by taking and VolumeId and removing the `snapshot-` prefix.
  - This also needs to change the write access to `rw` by using the lvchange command when the PersistentVolumeClaim specified the AccessMode as ReadWrite. Note that we do not need to change it back to read only, since we can limit the permissions of future mounts to read only by using the `MountOptions`. 
    - Alternatively we could make the Snapshots writeable by default
- In the [agent.go](../../pkg/driver/agent.go) the `NodeUnpublishVolume` function for the `node` type needs to be changed such that it unmounts the snapshot.


## Test Plan

- Create the PersistentVolumeClaim for the snapshot and verify that it is created successfully and that it can successfully be bound
- Create the PersistentVolumeClaim for the thick snapshot with a storage size larger than the snapSize but less than the size of the original volume and verify that the storage size has increased
- Create the PersistentVolumeClaim for the thick snapshot with a storage size larger than the size of the original volume and verify that the size has been set to exactly the size of the original volume
- Create the PersistentVolumeClaim for the thin snapshot with storage size larger than that of the snapshot and verify that it has not been resized
- Mount a PersistentVolumeClaim of a VolumeSnapshot to a pod and verify that it is mounted successfully
- Delete the VolumeSnapshot and verify that all PersistentVolumeClaims that reference it have been deleted
- Delete the VolumeSnapshot and verify that, once all the pods that have mounted a PersistentVolumeClaim with as dataSource the VolumeSnapshot have been deleted, the Snapshot is deleted
- Mount the PersistentVolumeClaim to a pod as Read and validate that the filesystem is read only
- Mount the PersistentVolumeClaim to a pod as ReadWrite and validate that the filesystem is writable
- Verify the original volume is working fine after mounting the snapshot and that any changes to the snapshot are not propagated to the original volume and vice versa
- Verify that creating a PersistentVolumeClaim for a non existing VolumeSnapshot fails

## Graduation Criteria

All testcases mentioned in [Test Plan](#test-plan) section need to be automated

## Drawbacks

- As far as I understand it, normally when creating a PersistentVolumeClaim with a snapshot as datasource it would clone the Snapshot into an actual volume, so we'd be diverging from this behaviour.
  - We could use an annotation (e.g. `local.csi.openebs.io/volume-type: mounted-snapshot`) applied to the PVC to specify the behaviour that we are implementing here. In the `CreateVolume` function of the [controller.go](../../pkg/driver/controller.go) we could then check whether this annotation is set and only then mount the snapshot. When the annotation is not set or set to `local.csi.openebs.io/volume-type: clone` the code would create a clone instead (which for now would `return nil, status.Error(codes.Unimplemented, "")`) until someone implements the logic to clone a snapshot. 
