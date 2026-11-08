
#  OpenEBS LocalPV Provisioner

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Chart Lint and Test](https://github.com/openebs/lvm-localpv/workflows/Chart%20Lint%20and%20Test/badge.svg)
![Release Charts](https://github.com/openebs/lvm-localpv/workflows/Release%20Charts/badge.svg?branch=develop)

A Helm chart for openebs lvm localpv provisioner. This chart bootstraps OpenEBS LVM LocalPV provisioner deployment on a [Kubernetes](http://kubernetes.io) cluster using the  [Helm](https://helm.sh) package manager.


**Homepage:** <http://www.openebs.io/>

## Get Repo Info

```console
helm repo add openebs-lvmlocalpv https://openebs.github.io/lvm-localpv
helm repo update
```

_See [helm repo](https://helm.sh/docs/helm/helm_repo/) for command documentation._

## Install Chart

Please visit the [link](https://openebs.github.io/lvm-localpv/) for install instructions via helm3.

```console
# Helm
$ helm install [RELEASE_NAME] openebs-lvmlocalpv/lvm-localpv --namespace [NAMESPACE]
```

<details>
  <summary>Click here if you're using MicroK8s.</summary>

  ```console
  microk8s helm3 install [RELEASE_NAME] openebs-lvmlocalpv/lvm-localpv --namespace [NAMESPACE] --set-string lvmNode.kubeletDir="/var/snap/microk8s/common/var/lib/kubelet/"
  ```
</details>


**Note:** If moving from the operator to helm
- Make sure the namespace provided in the helm install command is same as `OPENEBS_NAMESPACE` (by default it is `openebs`) env in the controller deployment.
- Before installing, clean up the stale deployment and daemonset from `kube-system` namespace using the below commands
```sh
kubectl delete deployment openebs-lvm-controller -n kube-system
kubectl delete ds openebs-lvm-node -n kube-system
```


_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Uninstall Chart

```console
# Helm
$ helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
# Helm
$ helm upgrade [RELEASE_NAME] [CHART] --install --namespace [NAMESPACE]
```

## Configuration

The following table lists the configurable parameters of the OpenEBS LVM Localpv chart and their default values.

```console
helm install openebs-lvmlocalpv openebs-lvmlocalpv/lvm-localpv --namespace openebs --create-namespace
```
<details>
  <summary>Click here if you're using MicroK8s.</summary>

  If you are using MicroK8s, it is necessary to add the following flag:

  ```console
  --set-string lvmNode.kubeletDir="/var/snap/microk8s/common/var/lib/kubelet/"
  ```
</details>

| Parameter                                           | Description                                                                      | Default                                 |
|-----------------------------------------------------|----------------------------------------------------------------------------------|-----------------------------------------|
| `global.imageRegistry`                              | Global override for container image registry                                                     | `""`                            |
| `global.imagePullSecrets`                           | Global image pull secrets (merged with local imagePullSecrets and not overridden)                                                  | `[]`                            |
| `global.imagePullPolicy`                            | Global override for image pull policy                                                       | `""`                            |
| `global.analytics.enabled`                          | Global override for analytics                                                    | `""`                            |
| `global.kubeletDir`                                 | Global override for kubelet directory                                                    | `""`                            |
| `crds.csi.volumeSnapshots.enabled`                  | Enable/Disable installation of VolumeSnapshot-related CRDs                       | `true`                                  |
| `crds.csi.volumeSnapshots.keep`                     | Keep the VolumeSnapshot-related CRDs on chart uninstall                          | `true`                                  |
| `crds.lvmLocalPv.enabled`                           | Enable/Disable installation of the LVMVolume, LVMSnapshot and LVMNode CRDs       | `true`                                  |
| `crds.lvmLocalPv.keep`                              | Keep the lvm-localpv CRDs on chart uninstall                                     | `true`                                  |
| `imagePullSecrets`                                  | Provides image pull secret                                                       | `""`                                    |
| `storageCapacity`                                   | Enable storage capacity tracking (`--enable-capacity` of csi-provisioner)        | `true`                                  |
| `lvmPlugin.image.registry`                          | Registry for openebs-lvm-plugin image                                            | `docker.io/`                            |
| `lvmPlugin.image.repository`                        | Image repository for openebs-lvm-plugin                                          | `openebs/lvm-driver`                    |
| `lvmPlugin.image.pullPolicy`                        | Image pull policy for openebs-lvm-plugin                                         | `IfNotPresent`                          |
| `lvmPlugin.image.tag`                               | Image tag for openebs-lvm-plugin, set to the chart version on release            | the chart version                       |
| `lvmPlugin.metricsPort`                             | The TCP port number used for exposing lvm-metrics                                | `9500`                                  |
| `lvmPlugin.allowedTopologies`                       | The comma seperated list of allowed node topologies                              | `kubernetes.io/hostname,`               |
| `lvmPlugin.ioLimits.enabled`                        | Set per-GB IOPS limits on the volumes of the lvmnode daemonset                   | `false`                                 |
| `lvmPlugin.ioLimits.containerRuntime`               | Container runtime of the nodes, used to find the pod cgroup for the IO limits    | `containerd`                            |
| `lvmPlugin.ioLimits.readIopsPerGB`                  | Read IOPS per GB of volume size, by volume group name prefix, e.g. `vg1-prefix:100,vg2-prefix:200`| `""`                                    |
| `lvmPlugin.ioLimits.writeIopsPerGB`                 | Write IOPS per GB of volume size, by volume group name prefix, e.g. `vg1-prefix:100,vg2-prefix:200`| `""`                                    |
| `lvmNode.driverRegistrar.image.registry`            | Registry for csi-node-driver-registrar image                                     | `registry.k8s.io/`                      |
| `lvmNode.driverRegistrar.image.repository`          | Image repository for csi-node-driver-registrar                                   | `sig-storage/csi-node-driver-registrar` |
| `lvmNode.driverRegistrar.image.pullPolicy`          | Image pull policy for csi-node-driver-registrar                                  | `IfNotPresent`                          |
| `lvmNode.driverRegistrar.image.tag`                 | Image tag for csi-node-driver-registrar                                          | `v2.13.0`                               |
| `lvmNode.updateStrategy.type`                       | Update strategy for lvmnode daemonset                                            | `RollingUpdate`                         |
| `lvmNode.updateStrategy.rollingUpdate.maxUnavailable` | Max unavailable lvmnode pods during a rolling update                             | `100%`                                  |
| `lvmNode.kubeletDir`                                | Kubelet mount point for lvmnode daemonset                                        | `"/var/lib/kubelet/"`                   |
| `lvmNode.logLevel`                                  | Log level (`--v`) of the csi-node-driver-registrar container                     | `5`                                     |
| `lvmNode.hostNetwork`                               | Run the lvmnode daemonset pods in the host network                               | `false`                                 |
| `lvmNode.kubeClientRateLimiter.qps`                 | Kubernetes API client QPS of the lvmnode agent, 0 keeps the client default       | `0`                                     |
| `lvmNode.kubeClientRateLimiter.burst`               | Kubernetes API client burst of the lvmnode agent, 0 keeps the client default     | `0`                                     |
| `lvmNode.serviceMonitor.enabled`                    | Create a ServiceMonitor for the lvmnode metrics (needs the `monitoring.coreos.com/v1` API)| `false`                                 |
| `lvmNode.priorityClass.create`                      | Create a PriorityClass named `<release>-<name>` and use it for the lvmnode daemonset pods. When `false` the pods get no priority class| `true`                                  |
| `lvmNode.priorityClass.name`                        | Name suffix of the lvmnode PriorityClass                                         | `lvm-localpv-csi-node-critical`         |
| `lvmNode.defaultFormatOptions`                      | Extra mkfs options per filesystem, used when the StorageClass of a volume does not set `formatOptions`. A StorageClass value replaces the default of its filesystem, the two are not merged. Keys are the formatted filesystem types (ext2, ext3, ext4, xfs, btrfs), values the space separated mkfs options of that filesystem as one string, e.g. `{"ext4": "-m 0", "xfs": "-i nrext64=0"}`. | `{}`                                    |
| `lvmNode.annotations`                               | Annotations for lvmnode daemonset metadata                                       | `""`                                    |
| `lvmNode.podAnnotations`                            | Annotations for lvmnode daemonset's pods metadata                                | `""`                                    |
| `lvmNode.resources`                                 | Resource and request and limit for lvmnode daemonset containers                  | `""`                                    |
| `lvmNode.labels`                                    | Labels for lvmnode daemonset metadata                                            | `""`                                    |
| `lvmNode.podLabels`                                 | Appends labels to the lvmnode daemonset pods                                     | `{"app": "openebs-lvm-node"}`           |
| `lvmNode.nodeSelector`                              | Nodeselector for lvmnode daemonset pods                                          | `""`                                    |
| `lvmNode.tolerations`                               | lvmnode daemonset's pod toleration values                                        | `""`                                    |
| `lvmNode.securityContext`                           | Security context for lvmnode daemonset container                                 | `""`                                    |
| `lvmController.resizer.image.registry`              | Registry for csi-resizer image                                                   | `registry.k8s.io/`                      |
| `lvmController.resizer.image.repository`            | Image repository for csi-resizer                                                 | `sig-storage/csi-resizer`               |
| `lvmController.resizer.image.pullPolicy`            | Image pull policy for csi-resizer                                                | `IfNotPresent`                          |
| `lvmController.resizer.image.tag`                   | Image tag for csi-resizer                                                        | `v2.0.0`                                |
| `lvmController.resizer.extraArgs`                   | Additional arguments to pass to csi-resizer                                      | `[]`                                    |
| `lvmController.snapshotter.enabled`                 | Enable the csi-snapshotter container. Disable if VolumeSnapshot CRDs are absent   | `true`                                  |
| `lvmController.snapshotter.image.registry`          | Registry for csi-snapshotter image                                               | `registry.k8s.io/`                      |
| `lvmController.snapshotter.image.repository`        | Image repository for csi-snapshotter                                             | `sig-storage/csi-snapshotter`           |
| `lvmController.snapshotter.image.pullPolicy`        | Image pull policy for csi-snapshotter                                            | `IfNotPresent`                          |
| `lvmController.snapshotter.image.tag`               | Image tag for csi-snapshotter                                                    | `v8.2.0`                                |
| `lvmController.snapshotter.extraArgs`               | Additional arguments to pass to csi-snapshotter                                  | `[]`                                    |
| `lvmController.snapshotController.enabled`          | Enable the snapshot-controller container. Disable if the cluster already runs one | `true`                                  |
| `lvmController.snapshotController.image.registry`   | Registry for snapshot-controller image                                           | `registry.k8s.io/`                      |
| `lvmController.snapshotController.image.repository` | Image repository for snapshot-controller                                         | `sig-storage/snapshot-controller`       |
| `lvmController.snapshotController.image.pullPolicy` | Image pull policy for snapshot-controller                                        | `IfNotPresent`                          |
| `lvmController.snapshotController.image.tag`        | Image tag for snapshot-controller                                                | `v8.2.0`                                |
| `lvmController.snapshotController.extraArgs`        | Additional arguments to pass to snapshot-controller                              | `[]`                                    |
| `lvmController.provisioner.image.registry`          | Registry for csi-provisioner image                                               | `registry.k8s.io/`                      |
| `lvmController.provisioner.image.repository`        | Image repository for csi-provisioner                                             | `sig-storage/csi-provisioner`           |
| `lvmController.provisioner.image.pullPolicy`        | Image pull policy for csi-provisioner                                            | `IfNotPresent`                          |
| `lvmController.provisioner.image.tag`               | Image tag for csi-provisioner                                                    | `v6.1.0`                                |
| `lvmController.provisioner.extraArgs`               | Additional arguments to pass to csi-provisioner                                  | `[]`                                    |
| `lvmController.updateStrategy.type`                 | Update strategy for lvm localpv controller deployment                            | `RollingUpdate`                         |
| `lvmController.replicas`                            | Replicas of the lvm localpv controller deployment, leader election is enabled above 1| `1`                                     |
| `lvmController.logLevel`                            | Log level (`--v`) of csi-resizer, csi-provisioner and snapshot-controller         | `5`                                     |
| `lvmController.kubeClientRateLimiter.qps`           | Kubernetes API client QPS of the lvm localpv controller, 0 keeps the client default| `0`                                     |
| `lvmController.kubeClientRateLimiter.burst`         | Kubernetes API client burst of the lvm localpv controller, 0 keeps the client default| `0`                                     |
| `lvmController.priorityClass.create`                | Create a PriorityClass named `<release>-<name>` and use it for the controller pods. When `false` the pods get no priority class| `true`                                  |
| `lvmController.priorityClass.name`                  | Name suffix of the lvm localpv controller PriorityClass                          | `lvm-localpv-csi-controller-critical`   |
| `lvmController.annotations`                         | Annotations for lvm localpv controller deployment metadata                       | `""`                                    |
| `lvmController.podAnnotations`                      | Annotations for lvm localpv controller deployment's pods metadata                | `""`                                    |
| `lvmController.resources`                           | Resource and request and limit for lvm localpv controller deployment containers  | `""`                                    |
| `lvmController.labels`                              | Labels for lvm localpv controller deployment metadata                            | `""`                                    |
| `lvmController.podLabels`                           | Appends labels to the lvm localpv controller deployment pods                     | `{"name": "openebs-lvm-controller"}`    |
| `lvmController.nodeSelector`                        | Nodeselector for lvm localpv controller deployment pods                          | `""`                                    |
| `lvmController.tolerations`                         | lvm localpv controller deployment's pod toleration values                        | `""`                                    |
| `lvmController.topologySpreadConstraints`           | lvm localpv controller deployment's pod topologySpreadConstraints values         | `""`                                    |
| `lvmController.securityContext`                     | Security context for lvm localpv controller deployment container                 | `""`                                    |
| `lvmController.podAntiAffinity.enabled`             | PodAntiAffinity for lvm localpv controller deployment                            | `false`                                    |
| `rbac.pspEnabled`                                   | Enable PodSecurityPolicy                                                         | `false`                                 |
| `serviceAccount.lvmNode.create`                     | Create a service account for lvmnode or not                                      | `true`                                  |
| `serviceAccount.lvmNode.name`                       | Name for the lvmnode service account                                             | `openebs-lvm-node-sa`                   |
| `serviceAccount.lvmController.create`               | Create a service account for lvm localpv controller or not                       | `true`                                  |
| `serviceAccount.lvmController.name`                 | Name for the lvm localpv controller service account                              | `openebs-lvm-controller-sa`             |
| `analytics.enabled`                                 | Enable or Disable google analytics for the controller                            | `true`                                  |

Specify each parameter using the `--set key=value[,key=value]` argument to `helm install`.

Alternatively, a YAML file that specifies the values for the parameters can be provided while installing the chart. For example,

```bash
helm install <release-name> -f values.yaml openebs/lvm-localpv
```

> **Tip**: You can use the default [values.yaml](values.yaml)
