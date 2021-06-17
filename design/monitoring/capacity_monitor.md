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
    - [Concepts](#concepts)
        - [Physical Volume](#physical-volume)
        - [Volume Group](#volume-group)
        - [Logical Volume](#logical-volume)
        - [Thin Logical Volume](#thin-logical-volume)
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

As a platform SRE, I should be able to efficiently manage the capacity-based provisioning of LVM LocalPV volumes on my cluster nodes.

### Concepts

LVM (Logical Volume Management) is a system for managing Logical Volumes and file-systems, in a manner more advanced and flexible than the traditional disk-partitioning method.
Benefits of using LVM:
- Resizing volumes on the fly
- Moving volumes on the fly
- Unlimited volumes
- Snapshots and data protection

Following are the basic concepts (components) that LVM manages:
- Physical Volume
- Volume Group
- Logical Volume

##### Physical Volume
A Physical Volume is a disk or block device, it forms the underlying storage unit for a LVM Logical Volume. In order to use a block device or its partition for LVM, it should be first initialized as a Physical Volume using `pvcreate` command from the LVM2 utils package. This places an LVM label near the start of the device.

##### Volume Group
A Volume Group (VG) is a named collection of physical and logical volumes. Physical Volumes are combined into Volume Groups. This creates a pool of disk space out of which Logical Volumes can be allocated. A Volume Group is divided into fixed-sized chunk called extents, which is the smallest unit of allocatable space. A VG can be created using the `vgcreate` command.

##### Logical Volume
A Logical Volume (LV) is an allocatable storage space of the required capacity from the VG. LVs look like devices to applications and can be mounted as file-systems. An LV is like a partition, but it is named (not numbered like a partition), can span across multiple underlying physical volumes in the VG and need not be contiguous. An LV can be created using the `lvcreate` command.

##### Thin Logical Volume
Logical Volumes can be thinly provisioned, which allows to create an LV, larger than the available physical extents. Using thin provisioning, a storage pool of free space known as a thin pool can be allocated to an arbitrary number of devices as thin LVs when needed by applications. The storage administrator can over-commit (over-provision) the physical storage by allocating LVs from the thin pool. As and when applications write the data and the thin pool fills up gradually, the underlying volume group (VG) can be expanded dynamically (using `vgextend`) by adding Physical Volumes on the fly. Once, VG is expanded, the thin pool can also be expanded (using `lvextend`).

### Implementation Details

This involves two phases - identifying the metrics and making them available for consumption.

#### Phase-1: Metrics Identification
##### Capacity-based Metrics
##### Usage-based Metrics
#### Phase-2: Metrics Export
##### Node-Exporter
##### Custom-Exporter

### Sample Dashboards

Below are sample Grafana dashboards:

![Volume Expansion Workflow](./images/resize_sequence_diagram.jpg)

### Test Plan

## Graduation Criteria

All testcases mentioned in [Test Plan](#test-plan) section need to be automated

## Drawbacks
NA

## Alternatives
NA
