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
      - [Metrics Identification](#metrics-identification)
        - [Capacity-based Metrics](#capacity-based-metrics)
        - [Usage-based Metrics](#usage-based-metrics)
      - [Metrics Export](#metrics-export)
        - [Node Exporter](#node-exporter)
        - [Custom Exporter](#custom-exporter)
    - [Sample Dashboards](#sample-dashboards)
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
- Clustered LVM.
- Snapshot space management (currently a non-goal, will be a future goal).

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

#### Metrics Identification
##### Capacity-based Metrics
- Total Provisioned Capacity on a node is the aggregate capacity of all Volume Groups on that node. Run the command to get the total capacity of a VG.
`vgdisplay -v <VG name> | grep 'VG Size'`
- Total Used Capacity on a node is the aggregate used capacity of all Volume Groups on that node. Run the command to get the used capacity of a VG.
`vgdisplay -v <VG name> | grep 'Alloc PE'`
- Total Free Capacity on a node is the aggregate free capacity of all Volume Groups on that node. Run the command to get the free capacity of a VG.
`vgdisplay -v <VG name> | grep 'Free PE'`
- Total Allocated Capacity on a node is the aggregate size of all LVs on that node. Run the command to get the size for all Logical Volumes. `lvs -o lv_full_name,lv_size`
- Total Used Capacity for all PVCs on a node is the aggregate used capacity of all LVs on that node. Run the command to get the used capacity for all LVs. `lvs -o lv_full_name,lv_size,data_percent,snap_percent,metadata_percent`
##### Usage-based Metrics
- Read IOPs
- Write IOPs
- Read Throughput
- Write Throughput
- Read Latency
- Write Latency
- Outstanding IOs
- Status
Since each LV corresponds to a device-mapper volume on the node, the performance statistics like IOPs, Throughput, Latency and Outstanding IOs can be obtained by running the standard `iostat -N` command on the node. The Status of each LV can be obtained from the `lvs -o lv_full_name,lv_active` command output.
#### Metrics Export
##### Node Exporter
Node Exporter is a Prometheus exporter for collecting hardware and OS kernel metrics exposed by *NIX* kernels using pluggable metrics collectors. There are many in-built collectors which are enabled by default in the node exporter. Using collectors 'diskstats' and 'filesystem', the node exporter is able to collect and export all the capacity and performance metrics for LVM Logical Volumes. These metrics can be stored in a  time-series database like Prometheus and visualized in Grafana with promQL queries. Since a thin pool is also an LV, the node exporter is able to collect its usage metrics as well.
##### Custom Exporter
Node exporter is able to fetch all metrics related to Logical Volumes. However, there is currently no in-built support for collecting metrics related to Volume Groups. We need a custom exporter to scrape VG metrics like vg_size, vg_used and vg_free.
### Sample Dashboards

Below are sample Grafana dashboards:

### Test Plan

## Graduation Criteria

All testcases mentioned in [Test Plan](#test-plan) section need to be automated

## Drawbacks
NA

## Alternatives
NA
