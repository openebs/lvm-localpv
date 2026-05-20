v1.9.0 / 2026-05-21
========================
This release of OpenEBS LVM-LocalPV introduces new features, critical bug fixes and enhancements. It builds on the stability delivered in 1.8.x with a focus on delivering important features and fixing outstanding issues.

## New Features and Enhancements
* Volume Attribute Class
   Added volume attribute class support by @rybas-dv in [#444](https://github.com/openebs/lvm-localpv/pull/444)
   OEP : [#4145](https://github.com/openebs/openebs/pull/4145)
* Add support to update QoS policies and IOPS profiles
    Added volume attribute class support by @rybas-dv in [#444](https://github.com/openebs/lvm-localpv/pull/444)
   OEP : [#4145](https://github.com/openebs/openebs/pull/4145)

## Additional contributions
* chore: yq_ibl fix, fix registry to docker by @Abhinandan-Purkait in [#429](https://github.com/openebs/lvm-localpv/pull/429)
* ci: push lvm-localpv helm chart as OCI by @tiagolobocastro in [#431](https://github.com/openebs/lvm-localpv/pull/431)
* ci: refactor ci to reuse actions across workflows by @rohan2794 in [#440](https://github.com/openebs/lvm-localpv/pull/440)
* ci: add trivy scan by @pchandra19 in [#450](https://github.com/openebs/lvm-localpv/pull/450)
* fix: update unit test code coverage action by @rohan2794 in [#451](https://github.com/openebs/lvm-localpv/pull/451)
* fix: update readme with restore from snapshot feature by @rohan2794 in [#453](https://github.com/openebs/lvm-localpv/pull/453)
* ci: bump trivy action version by @pchandra19 in [#457](https://github.com/openebs/lvm-localpv/pull/457)
* fix(provisioning): use thin pool free space in Controller/GetCapacity if thin provisioning is enabled by @Schmazda in [#459](https://github.com/openebs/lvm-localpv/pull/459)
* ci: update trivy-action version by @pchandra19 in [#465](https://github.com/openebs/lvm-localpv/pull/465)
* fix: remove duplicate config values by @rohan2794 in [#466](https://github.com/openebs/lvm-localpv/pull/466)
* fix: propagate unmount errors in NodeUnpublishVolume (#461) by @kotyara85 in [#469](https://github.com/openebs/lvm-localpv/pull/469)
* fix(controller): fallback to VG capacity if thin pool is missing by @rohan2794 in [#470](https://github.com/openebs/lvm-localpv/pull/470)
* Update README to mark Backup/Restore as working by @todeb in [#472](https://github.com/openebs/lvm-localpv/pull/472)
* feat: add helm global values by @krishnaGajabi in [#460](https://github.com/openebs/lvm-localpv/pull/460)

## New Contributors
* @pchandra19 made their first contribution in [#450](https://github.com/openebs/lvm-localpv/pull/450)
* @rybas-dv made their first contribution in [#444](https://github.com/openebs/lvm-localpv/pull/444)
* @Schmazda made their first contribution in [#459](https://github.com/openebs/lvm-localpv/pull/459)
* @kotyara85 made their first contribution in [#469](https://github.com/openebs/lvm-localpv/pull/469)
* @todeb made their first contribution in [#472](https://github.com/openebs/lvm-localpv/pull/472)

**Full Changelog**: [v1.8.1...v1.9.0](https://github.com/openebs/lvm-localpv/compare/v1.8.1...v1.9.0)


v1.8.1 / 2026-02-04
========================
## Bug Fixes and Improvements
* make delete volume idempotent when node is removed from cluster by @abhilashshetty04 in [#413](https://github.com/openebs/lvm-localpv/pull/413)
* Add podAntiAffinity option to lvm-controller by @SupeReuven in [#423](https://github.com/openebs/lvm-localpv/pull/423)
* fix(lvm): Consistently build volume device paths as /dev/mapper/group--name-volume--name by @pschichtel in [#433](https://github.com/openebs/lvm-localpv/pull/433)
* docs: update helm docs by @abhilashshetty04 in [#433](https://github.com/openebs/lvm-localpv/pull/443)

## New Contributors
* @SupeReuven made their first contribution in [#423](https://github.com/openebs/lvm-localpv/pull/423)
* @pschichtel made their first contribution in [#433](https://github.com/openebs/lvm-localpv/pull/433)

**Full Changelog**: [v1.8.0...v1.8.1](https://github.com/openebs/lvm-localpv/compare/v1.8.0...v1.8.1)

v1.8.0 / 2025-11-18
========================
This release of OpenEBS LVM-LocalPV introduces new features, critical bug fixes, enhancements to scheduling, runtime, CSI spec compliance, storage resource management, as well as several documentation and maintenance updates. It builds on the stability delivered in 1.7.0 with a focus on delivering important features and fixing outstanding issues.

## New Features and Enhancements
* Snapshot restore
   LocalPV-LVM snapshot had limited capabilities. Now we support restoring a snapshot to volume by @rohan2794 in [#417](https://github.com/openebs/lvm-localpv/pull/417), [#419](https://github.com/openebs/lvm-localpv/pull/419)
   OEP : [#4080](https://github.com/openebs/openebs/pull/4080)

* ThinPool space reclamation
   LocalPV-LVM will cleanup the thinpool LV after deleting the last thin volume of the thinpool by @dsharma-dc in [#412](https://github.com/openebs/lvm-localpv/pull/412)

*  Scheduler fixes and enhancements
    Record thinpool statistics in lvmnode CR. Fail fast CreateVolume request if thick PVC size cannot be accommodated by any VG.
    Considers thinpool free space while scheduling thin pvc in SpaceWeighted algorithm by @abhilashshetty04 in [#418](https://github.com/openebs/lvm-localpv/pull/418)
    OEP: [#4083](https://github.com/openebs/openebs/pull/4083)

* Runtime improvements
   Updates Go runtime, k8s modules, golint packages etc by @jochenseeber in [#416](https://github.com/openebs/lvm-localpv/pull/416)

## Continuous Integration and Maintenance
*  Staging CI
    Introduction of the staging CI, which enables creating a staging build for e2e testing before releasing, the artifacts are then copied over to production build hosts. by @Abhinandan-Purkait

## Additional contributions
* chore: add helm docs and fix script by @dsharma-dc in [#394](https://github.com/openebs/lvm-localpv/pull/394)
* docs(provisioning): :memo: Add formatOptions parameter document by @mhkarimi1383 in [#396](https://github.com/openebs/lvm-localpv/pull/396)
* docs(provisioning): :memo: Fix mountOptions indent in docs by @mhkarimi1383 in [#401](https://github.com/openebs/lvm-localpv/pull/401)
* fix(resize): validate lv size before executing lvextend by @dsharma-dc in [#409](https://github.com/openebs/lvm-localpv/pull/409)
* chore: update base alpine image from 3.18.4 to 3.22.1 by @krishnaGajabi in [#414](https://github.com/openebs/lvm-localpv/pull/414)
* feat(restore): add thin snapshot restore ci test by @rohan2794 in [#421](https://github.com/openebs/lvm-localpv/pull/421)
* remove node from selected which are not in nmap by @abhilashshetty04 in [#422](https://github.com/openebs/lvm-localpv/pull/422)
* fix: update snapshot create response with restore size for snapshot by @rohan2794 in [#424](https://github.com/openebs/lvm-localpv/pull/424)
* fix: modify test script according to new log fatal script by @Abhinandan-Purkait in [#425](https://github.com/openebs/lvm-localpv/pull/425)
* chore: yq_ibl fix, fix registry to docker by @Abhinandan-Purkait in [428](https://github.com/openebs/lvm-localpv/pull/428)


## New Contributors
* @krishnaGajabi made their first contribution in [#414](https://github.com/openebs/lvm-localpv/pull/414)
* @jochenseeber made their first contribution in [#416](https://github.com/openebs/lvm-localpv/pull/416)
* @rohan2794 made their first contribution in [#417](https://github.com/openebs/lvm-localpv/pull/417)
* @moss-telavox made their first contribution in [#411](https://github.com/openebs/lvm-localpv/pull/411)

**Full Changelog**: [v1.7.0...v1.8.0](https://github.com/openebs/lvm-localpv/compare/v1.7.0...v1.8.0)

v1.7.0 / 2025-06-03
========================
## Bug Fixes and Improvements
* test(bdd): adding scheduler logic bdd by @abhilashshetty04 in [#323](https://github.com/openebs/lvm-localpv/pull/323)
* chore(deps): update analytics dependency by @niladrih in [#325](https://github.com/openebs/lvm-localpv/pull/325)
* feat(charts): add analytics ID and KEY envs to csi controller by @niladrih in [#326](https://github.com/openebs/lvm-localpv/pull/326)
* small typo by @chandanpasunoori in [#327](https://github.com/openebs/lvm-localpv/pull/327)
* fix readme typo by @runzhliu in [#329](https://github.com/openebs/lvm-localpv/pull/329)
* fix(test/chores): Buildscripts and E2E script shebangs by @mhkarimi1383 in [#351](https://github.com/openebs/lvm-localpv/pull/351)
* fix(chores): Fix gocov-html package by @mhkarimi1383 in [#354](https://github.com/openebs/lvm-localpv/pull/354)
* feat(chores): Add missing tools to nix-shell by @mhkarimi1383 in [#352]https://github.com/openebs/lvm-localpv/pull/352
* build: a number of fixes on Makefile and nix-shell by @tiagolobocastro in [#360](https://github.com/openebs/lvm-localpv/pull/360)
* docs(security): cross-reference security docs by @tiagolobocastro in [#362](https://github.com/openebs/lvm-localpv/pull/362)
* docs: improve contribution guides by @tiagolobocastro in [#361](https://github.com/openebs/lvm-localpv/pull/361)
* build: various makefile fixes by @tiagolobocastro in [#363](https://github.com/openebs/lvm-localpv/pull/363)
* [fix] Fix invalid YAMLs in crds by @nilroy in [#364](https://github.com/openebs/lvm-localpv/pull/364)
* test: add volume provisioning test on cordoned node by @abhilashshetty04 in [#375](https://github.com/openebs/lvm-localpv/pull/375)
* docs: update microk8s instructions by @dsharma-dc in [#378](https://github.com/openebs/lvm-localpv/pull/378)
* correctly indent podLabels on node service by @ecniiv in [#380](https://github.com/openebs/lvm-localpv/pull/380)
* use parseint for capacity parsing to avoid range overflow by @abhilashshetty04 in [#387](https://github.com/openebs/lvm-localpv/pull/387)
* update csi spec version to v1.9.0 by @abhilashshetty04 in [#391](https://github.com/openebs/lvm-localpv/pull/391)
* feat(provisioning): extra format options (`mkfs`) added by @mhkarimi1383 in [#335](https://github.com/openebs/lvm-localpv/pull/335)

## New Contributors
* @chandanpasunoori made their first contribution in [#327](https://github.com/openebs/lvm-localpv/pull/327)
* @runzhliu made their first contribution in [#329](https://github.com/openebs/lvm-localpv/pull/329)
* @tiagolobocastro made their first contribution in [#360](https://github.com/openebs/lvm-localpv/pull/360)
* @nilroy made their first contribution in [#364](https://github.com/openebs/lvm-localpv/pull/364)
* @dependabot made their first contribution in [#370](https://github.com/openebs/lvm-localpv/pull/370)
* @ecniiv made their first contribution in [#380](https://github.com/openebs/lvm-localpv/pull/380)

**Full Changelog**: [v1.6.1...v1.7.0](https://github.com/openebs/lvm-localpv/compare/v1.6.1...v1.7.0)

lvm-localpv-1.6.2 / 2024-09-19
========================
<b>NOTE</b>: This was only a chart release that addressed a bug on prior chart. 
* fix(chart): revert env OPENEBS_NAMESPACE to LVM_NAMESPACE for v1.6.x ([#333](https://github.com/openebs/lvm-localpv/pull/333),[@niladrih](https://github.com/niladrih))

v1.6.1 / 2024-09-16
========================
* chore(deps): update analytics dependency ([#325](https://github.com/openebs/lvm-localpv/pull/325),[@niladrih](https://github.com/niladrih))

v1.6.0 / 2024-07-03
========================
* feat(analytics): add heartbeat pinger ([#318](https://github.com/openebs/lvm-localpv/pull/318),[@niladrih](https://github.com/niladrih))

v0.4.0 / 2021-04-14
========================
* updated storage and apiextension version to v1 ([#40](https://github.com/openebs/lvm-localpv/pull/40),[@shubham14bajpai](https://github.com/shubham14bajpai))
* add support for thin provision lvm volumes ([#30](https://github.com/openebs/lvm-localpv/pull/30),[@prateekpandey14](https://github.com/prateekpandey14))
* upgrade grpc lib dependency to v1.34.2 ([#37](https://github.com/openebs/lvm-localpv/pull/37),[@iyashu](https://github.com/iyashu))
* reload lvmetad cache before querying volume groups ([#38](https://github.com/openebs/lvm-localpv/pull/38),[@iyashu](https://github.com/iyashu))

v0.4.0-RC2 / 2021-04-12
========================

v0.4.0-RC1 / 2021-04-07
========================
* updated storage and apiextension version to v1 ([#40](https://github.com/openebs/lvm-localpv/pull/40),[@shubham14bajpai](https://github.com/shubham14bajpai))
* add support for thin provision lvm volumes ([#30](https://github.com/openebs/lvm-localpv/pull/30),[@prateekpandey14](https://github.com/prateekpandey14))
* upgrade grpc lib dependency to v1.34.2 ([#37](https://github.com/openebs/lvm-localpv/pull/37),[@iyashu](https://github.com/iyashu))
* reload lvmetad cache before querying volume groups ([#38](https://github.com/openebs/lvm-localpv/pull/38),[@iyashu](https://github.com/iyashu))


v0.3.0 / 2021-03-12
========================
* Add e2e-test for lvm volume resize support  ([#32](https://github.com/openebs/lvm-localpv/pull/32),[@w3aman](https://github.com/w3aman))
* Add e2e-test for lvm-localpv driver provisioning ([#29](https://github.com/openebs/lvm-localpv/pull/29),[@w3aman](https://github.com/w3aman))
* add volume group capacity tracking ([#21](https://github.com/openebs/lvm-localpv/pull/21),[@iyashu](https://github.com/iyashu))
* move the bdd test cases to github action ([#27](https://github.com/openebs/lvm-localpv/pull/27),[@pawanpraka1](https://github.com/pawanpraka1))
* set IOPS, BPS limit for Pod accessing a Volume ([#19](https://github.com/openebs/lvm-localpv/pull/19),[@abhranilc](https://github.com/abhranilc))
* adding bdd test cases for LVM Driver ([#26](https://github.com/openebs/lvm-localpv/pull/26),[@pawanpraka1](https://github.com/pawanpraka1))
* Add e2e-test for lvm-localpv ([#24](https://github.com/openebs/lvm-localpv/pull/24),[@w3aman](https://github.com/w3aman))
* enable pod resheduling cause of node insufficient capacity ([#23](https://github.com/openebs/lvm-localpv/pull/23),[@iyashu](https://github.com/iyashu))
* updating go mod to v0.2.0 ([#25](https://github.com/openebs/lvm-localpv/pull/25),[@pawanpraka1](https://github.com/pawanpraka1))


v0.2.0 / 2021-02-12
========================
* add support for create/delete snapshot for LVM localPV ([#12](https://github.com/openebs/lvm-localpv/pull/12),[@akhilerm](https://github.com/akhilerm))
* adding raw block volume support for LVM LocalPV ([#14](https://github.com/openebs/lvm-localpv/pull/14),[@pawanpraka1](https://github.com/pawanpraka1))
* add capacity weighted scheduler and make it default for scheduling volumes ([#20](https://github.com/openebs/lvm-localpv/pull/20),[@akhilerm](https://github.com/akhilerm))
* ensure lvm volume creation & deletion idempotent ([#16](https://github.com/openebs/lvm-localpv/pull/16),[@iyashu](https://github.com/iyashu))


v0.1.0 / 2021-01-13
========================
* adding resize support for lvm volumes  ([#2](https://github.com/openebs/lvm-localpv/pull/2),[@pawanpraka1](https://github.com/pawanpraka1))
* adding multi arch build process for LVM Driver ([#1](https://github.com/openebs/lvm-localpv/pull/1),[@pawanpraka1](https://github.com/pawanpraka1))
