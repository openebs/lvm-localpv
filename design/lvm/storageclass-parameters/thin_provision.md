---
title: LVM-LocalPV Thin Provision
authors:
  - "@pawanpraka1"
owners:
  - "@kmova"
creation-date: 2021-06-16
last-updated: 2021-06-16
status: Implemented
---

# LVM-LocalPV Thin Provisioning

## Table of Contents
- [LVM-LocalPV Thin Provisioning](#lvm-localpv-thin-provisioning)
  - [Table of Contents](#table-of-contents)
  - [Summary](#summary)
  - [Motivation](#motivation)
    - [Goals](#goals)
  - [Proposal](#proposal)
    - [User Stories](#user-stories)
    - [Implementation Details](#implementation-details)
    - [Usage details](#usage-details)
    - [Test Plan](#test-plan)
  - [Graduation Criteria](#graduation-criteria)
  - [Drawbacks](#drawbacks)
  - [Alternatives](#alternatives)


## Summary

This proposal charts out the workflow details to support creation of thin provisioned volumes.

## Motivation

### Goals

- Able to provision thin volume in user specified VolumeGroup.

## Proposal

### User Stories

- Thin provisioned volume will occupy storage only on demand and it will help a lot save cost on storage.

### Implementation Details

- User/Admin has to set `thinProvision` parameter to "yes" under storageclass parameters
  which informs driver to create thin provisioned volume.
- During volume provisioning time external-provisioner will read all key-value pairs
  that are specified under referenced storageclass and pass information to CSI
  driver as payload for `CreateVolume` gRPC request.
- After receiving the `CreateVolume` request CSI driver will pick appropriate node based
  on scheduling attributes(like topology information, matching VG name and available capacity)
  and creates LVM volume resource by setting `Spec.ThinProvision` to yes along with other properties.
- Once the LVMVolume resource is created corresponding node LVM volume controller reconcile
  LVM volume resource in the following way:
  - LVM controller will check `Spec.ThinProvision` field, if the field is set then controller
    will perform following operations:
    - Fetch information about existence of thin pool in matching volume group.
      - If no such pool found then controller will create new pool with
        min(volume_request_size, VG_available_size) size along with thin volume.
        Command used to create thin pool & volume: `lvcreate -L <min_pool_size> -T lvmvg/lvmvg_thinpool  -V <volume_size> -n <volume_name> -y`.
      - If there is a thin pool with <vg_name>_thinpool name then controller will create thin volume.
        Command used to create thin volume: `lvcreate -T lvmvg/lvmvg_thinpool -V <volume_size> -n <volume_name> -y`
    - If thin volume creation is successfull then controller will LVM volume resource as `Ready`.
- After watching `Ready` status CSI driver will return success response to `CreateVolume` gRPC
  request.

### Usage details

1. User/Admin can configure `thinProvision` value to `yes` under storageclass parameter.
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

### Test Plan
- Provision a thin volume with a capacity less than underlying VG size.
- Provision a thin volume with a capacity greater than underlying VG size.
- Provision multiple thin volumes with capacities greater than underlying VG size.
 
## Graduation Criteria

All testcases mentioned in [Test Plan](#test-plan) section need to be automated

## Drawbacks
NA

## Alternatives
NA
