---
title: LVM-LocalPV Node Capacity Management and Monitoring
authors:
  - "@avishnu"
owners:
  - "@kmova"
creation-date: 2021-06-17
last-updated: 2021-06-17
status: In-progress
---

# LVM-LocalPV Node Capacity Management and Monitoring

## Table of Contents

- [LVM-LocalPV Node Capacity_Management_and_Monitoring](#lvm-localpv-node-capacity-management-and-monitoring)
  - [Table of Contents](#table-of-contents)
  - [Summary](#summary)
  - [Motivation](#motivation)
    - [Goals](#goals)
    - [Non-Goals](#non-goals)
  - [Proposal](#proposal)
    - [User Stories](#user-stories)
    - [Implementation Details](#implementation-details)
        - [Controller Expansion](#controller-expansion)
        - [Filesystem Expansion](#filesystem-expansion)
    - [Steps to perform volume expansion](#steps-to-perform-volume-expansion)
    - [High level Sequence Diagram](#high-level-sequence-diagram)
    - [Test Plan](#test-plan)
  - [Graduation Criteria](#graduation-criteria)
  - [Drawbacks](#drawbacks)
  - [Alternatives](#alternatives)

## Summary

This proposal charts out the design details to implement monitoring for doing effective capacity management on nodes having LVM-LocalPV Volumes.

## Motivation
Platform SREs must be able to easily query the capacity details at per node level for checking the utilization and planning purposes.

### Goals
- Platform SREs must be able to query the information at the following granularity:
  - Node level
  - Volume Group level
- Platform SREs need the following information that will help with planning based on the capacity:
  - Total Provisioned Capacity
  - Total Allocated Capacity to the PVCs ( (When thin-provisioned PVs are used - it is possible that total allocated capacity can be greater than total provisioned capacity.)
  - Total Used Capacity by the PVCs
- Platforms SREs need the following performance metrics that will help with planning based on the usage:
  - Read / Write Throughput and IOPS (PV)
  - Read / Write Latency (PV)
  - Outstanding IOs (PV)
  - Status ( online / offline )

This document lists the relevant metrics for the above information and the steps to fetch the same.

### Non-Goals

- The visualization and alerting for the above metrics.

## Proposal

### User Stories

As a platform SRE, I should be able to efficiently manage the provisioning of LVM LocalPV on my cluster nodes.

### Concepts



### Implementation Details

LVM volume expansion is a two-step process(controller expansion & filesystem expansion) and
it gets triggered upon updating PVC capacity.

##### Controller Expansion

- External-resizer is a sidecar in `openebs-lvm-controller` pod which watches for PVC capacity
  updates. Once it receives the event external-resizer will make a `ControllerExpandVolume` gRPC request
  to LVM-CSI controller plugin.
- LVM-CSI controller plugin is another container in the same pod where external-resizer is running.
  Once the controller plugin receives the `ControllerExpandVolume` gRPC request it does the following steps:
  - List active snapshots on resizing volume, if there are any snapshots plugin will return an error as
    a response to the request since LVM does not support online resizing of volume if snapshot(s) exist.
  - If there are no snapshots then the plugin will update the desired capacity on corresponding LVMVolume
    resource and return success response to the request.
- Once external-resizer gets success response it will mark PVC pending for filesystem expansion else
  it will retry until it receives success response.

Note: `ControllerExpandVolume` gRPC request returns an error if volume has snapshots.

##### Filesystem Expansion

- Container Orchestrator(CO)[kubelet] will send `NodeExpandVolume` gRPC request to LVM-CSI node driver
  once it observes the PVC marked for filesystem expansion.
- LVM-CSI node driver receives the `NodeExpandVolume` gRPC request and it will expand the lvm volume
  using lvextend cli and trigger filesystem expansion. If expansion is successful it will return
  success gRPC response.

  Note: Kubelet will send requests only when the volume is published on a node.

### Steps to perform volume expansion

1. Edit PVC and update capacity to desired value. In below example updated
   the PVC capacity from 4Gi to 8Gi.

```sh
kubectl patch pvc csi-lvmpv -p '{"spec":{"resources":{"requests":{"storage":"8Gi"}}}}'
persistentvolumeclaim/csi-lvmpv patched
```

2. User can observe the resize related events by describing PVC

```sh
kubectl describe pvc csi-lvmpv
...
...
Events:
  Type     Reason                      Age   From                                                                                Message
  ----     ------                      ----  ----                                                                                -------
  Normal   Provisioning                11m   local.csi.openebs.io_openebs-lvm-controller-0_b4700a50-b7cd-4de5-bc26-d3dd832ac9eb  External provisioner is provisioning volume for claim "default/csi-lvmpv"
  Normal   ExternalProvisioning        11m   persistentvolume-controller                                                         waiting for a volume to be created, either by external provisioner "local.csi.openebs.io" or manually created by system administrator
  Normal   ProvisioningSucceeded       11m   local.csi.openebs.io_openebs-lvm-controller-0_b4700a50-b7cd-4de5-bc26-d3dd832ac9eb  Successfully provisioned volume pvc-f532e80d-b39b-4801-837b-57a47ae08ea8
  Normal   Resizing                    95s   external-resizer local.csi.openebs.io                                               External resizer is resizing volume pvc-f532e80d-b39b-4801-837b-57a47ae08ea8
  Warning  ExternalExpanding           95s   volume_expand                                                                       Ignoring the PVC: didn't find a plugin capable of expanding the volume; waiting for an external controller to process this PVC.
  Normal   FileSystemResizeRequired    95s   external-resizer local.csi.openebs.io                                               Require file system resize of volume on nod
```

3. Once the filesystem expansion is succeeded then success events will
   be generated and status of PVC also will be updated to latest capacity.
```sh
kubectl describe pvc csi-lvmpv

 Normal   FileSystemResizeSuccessful  21s   kubelet                                                                             MountVolume.NodeExpandVolume succeeded for volume "pvc-f532e80d-b39b-4801-837b-57a47ae08ea8"
```

### High level Sequence Diagram

Below is high level sequence diagram for volume expansion workflow

![Volume Expansion Workflow](./images/resize_sequence_diagram.jpg)

### Test Plan
A test plan will include following test cases:
- Test volume expansion operation on all supported filesystems(ext3, ext4, xfs, btrfs).
- Test volume expansion while previous expansion of the volume is already in progress.
- Restart of node LVM-Node-CSI-driver while filesystem expansion is in progress.
- Test volume expansion while application is not consuming the volume(It should succeed only after application mounts the volume).
- Shutdown the node while filesystem expansion is in progress and after recovering volume expansion should be succeeded.
- Test volume expansion on statefulset application with multiple replicas.
- Test volume expansion of thin provisioned volume.
- Test volume expansion with snapshot(s).
- Test volume expansion for thick provisioned volume group by increasing size greater then vg size.


## Graduation Criteria

All testcases mentioned in [Test Plan](#test-plan) section need to be automated

## Drawbacks
NA

## Alternatives
NA
