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
