---
title: LVM-LocalPV RAID
authors:
  - "@nicholascioli"
owners: []
creation-date: 2023-11-04
last-updated: 2023-11-04
status: Implemented
---

# LVM-LocalPV RAID

## Table of Contents
- [LVM-LocalPV RAID](#lvm-localpv-raid)
  - [Table of Contents](#table-of-contents)
  - [Summary](#summary)
  - [Motivation](#motivation)
    - [Goals](#goals)
    - [Non Goals](#non-goals)
  - [Proposal](#proposal)
    - [User Stories](#user-stories)
    - [Implementation Details](#implementation-details)
    - [Usage details](#usage-details)
    - [Test Plan](#test-plan)
  - [Graduation Criteria](#graduation-criteria)
  - [Drawbacks](#drawbacks)
  - [Alternatives](#alternatives)


## Summary

This proposal charts out the workflow details to support creation of RAID volumes.

## Motivation

### Goals

- Able to provision RAID volumes in a VolumeGroup.
- Able to specify VolumeGroup-specific RAID options for all sub volumes.
- Able to specify extra options for all volumes in a VolumeGroup.

### Non Goals

- Validating combinations of RAID types / options.

## Proposal

### User Stories

- RAIDed volumes provide data redundancy and can mitigate data loss due to individual drive failures.
- Ability to specify extra arguments for VolumeGroups allow for user customizations without needing
  to rework k8s schemas.

### Implementation Details

- User/Admin has to set RAID-sepcific options under storageclass parameters which
  are used when creating volumes in the VolumeGroup.
- During volume provisioning time external-provisioner will read all key-value pairs
  that are specified under referenced storageclass and pass information to CSI
  driver as payload for `CreateVolume` gRPC request.
- After receiving the `CreateVolume` request CSI driver will pick appropriate node based
  on scheduling attributes(like topology information, matching VG name and available capacity)
  and creates LVM volume resource by setting `Spec.RaidType` to a valid type along with other properties.
- Once the LVMVolume resource is created corresponding node LVM volume controller reconcile
  LVM volume resource in the following way:
  - LVM controller will check `Spec.RaidType` field, if the field is set to anything other
    than `linear`, then the controller will perform following operations:
    - Fetch information about existence of matching VolumeGroup.
      - If there is a VolumeGroup with <vg_name> name then controller will create a volume.
        Command used to create thin volume: `lvcreate --type <RAID_TYPE> --raidintegrity <INTEGRITY> --nosync ... <LVCREATEOPTIONS> -y`
    - If volume creation is successfull then controller will LVM volume resource as `Ready`.
- After watching `Ready` status CSI driver will return success response to `CreateVolume` gRPC
  request.

### Usage details

1. User/Admin can configure the following options under the storageclass parameters.

Option | Required | Valid Values | Description
-------|----------|--------------|-------------------
`type` | `true` | `raid0` / `stripe`, `raid` / `raid1` / `mirror`, `raid5`, `raid6`, `raid10` | The RAID type of the volume.
`integrity` | `false` | `true`, `false` | Whether or not to enable DM integrity for the volume. Defaults to `false`.
`mirrors` | depends | [0, ∞) | Mirror count. Certain RAID configurations require this to be set.
`nosync` | `false` | `true`, `false` | Whether or not to disable the initial sync. Defaults to false.
`stripecount` | depends | [0, ∞) | Stripe count. Certain RAID configurations require this to be set.
`stripesize` | `false` | [0, ∞) (but must be a power of 2) | The size of each stripe. If not specified, LVM will choose a sane default.
`lvcreateoptions` | `false` | String, delimited by `;` | Extra options to be passed to LVM when creating volumes.

An example is shown below
```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: openebs-lvm
provisioner: local.csi.openebs.io
parameters:
  storage: "lvm"
  volgroup: "lvmvg"
  raidType: "raid1"
  lvcreateoptions: "--vdo;--readahead auto"
```

### Test Plan
- Provision an application on various RAID configurations, verify volume accessibility from application,
  and verify that `lvs` reports correct RAID information.

## Graduation Criteria

All testcases mentioned in [Test Plan](#test-plan) section need to be automated

## Drawbacks

- Since the RAID options exist at the storageclass level, changes to the storage
  class RAID options is not possible without custom logic per RAID type or manual
  operator interactions.
- Validation of the RAID options depend on the version of LVM2 installed as well as
  the type of RAID used and its options. This is outside of the scope of these changes
  and will cause users to have to debug issues with a finer comb to see why certain
  options do not work together or on their specific machine.

## Alternatives

RAID can be done in either software or hardware, with many off-the-shelf products
including built-in hardware solutions. There are also other software RAID alternatives
that can be used below LVM, such as mdadm.

This unfortunately requires operators to decouple
the SotrageClass from the RAID configuration, but does simplify the amount of code maintained by
this project.
