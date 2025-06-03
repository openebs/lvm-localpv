
#  OpenEBS Local PV LVM Provisioner

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
![Chart Lint and Test](https://github.com/openebs/lvm-localpv/workflows/Chart%20Lint%20and%20Test/badge.svg)
![Release Charts](https://github.com/openebs/lvm-localpv/workflows/Release%20Charts/badge.svg?branch=develop)

A Helm chart for openebs localpv lvm provisioner. This chart bootstraps OpenEBS LVM LocalPV provisioner deployment on a [Kubernetes](http://kubernetes.io) cluster using the  [Helm](https://helm.sh) package manager.


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
| `crds.csi.volumeSnapshots.enabled`                  | Enable/Disable installation of VolumeSnapshot-related CRDs                       | `true`                                   |
| `imagePullSecrets`                                  | Provides image pull secret                                                       | `""`                                    |
| `lvmPlugin.image.registry`                          | Registry for openebs-lvm-plugin image                                            | `""`                                    |
| `lvmPlugin.image.repository`                        | Image repository for openebs-lvm-plugin                                          | `openebs/lvm-driver`                    |
| `lvmPlugin.image.pullPolicy`                        | Image pull policy for openebs-lvm-plugin                                         | `IfNotPresent`                          |
| `lvmPlugin.image.tag`                               | Image tag for openebs-lvm-plugin                                                 | `1.7.0`                                 |
| `lvmPlugin.metricsPort`                             | The TCP port number used for exposing lvm-metrics                                | `9500`                                  |
| `lvmPlugin.allowedTopologies`                       | The comma seperated list of allowed node topologies                              | `kubernetes.io/hostname,`               |
| `lvmNode.driverRegistrar.image.registry`            | Registry for csi-node-driver-registrar image                                     | `registry.k8s.io/`                      |
| `lvmNode.driverRegistrar.image.repository`          | Image repository for csi-node-driver-registrar                                   | `sig-storage/csi-node-driver-registrar` |
| `lvmNode.driverRegistrar.image.pullPolicy`          | Image pull policy for csi-node-driver-registrar                                  | `IfNotPresent`                          |
| `lvmNode.driverRegistrar.image.tag`                 | Image tag for csi-node-driver-registrar                                          | `v2.13.0`                                |
| `lvmNode.updateStrategy.type`                       | Update strategy for lvmnode daemonset                                            | `RollingUpdate`                         |
| `lvmNode.kubeletDir`                                | Kubelet mount point for lvmnode daemonset                                        | `"/var/lib/kubelet/"`                   |
| `lvmNode.annotations`                               | Annotations for lvmnode daemonset metadata                                       | `""`                                    |
| `lvmNode.podAnnotations`                            | Annotations for lvmnode daemonset's pods metadata                                | `""`                                    |
| `lvmNode.resources`                                 | Resource and request and limit for lvmnode daemonset containers                  | `""`                                    |
| `lvmNode.labels`                                    | Labels for lvmnode daemonset metadata                                            | `""`                                    |
| `lvmNode.podLabels`                                 | Appends labels to the lvmnode daemonset pods                                     | `""`                                    |
| `lvmNode.nodeSelector`                              | Nodeselector for lvmnode daemonset pods                                          | `""`                                    |
| `lvmNode.tolerations`                               | lvmnode daemonset's pod toleration values                                        | `""`                                    |
| `lvmNode.securityContext`                           | Security context for lvmnode daemonset container                                 | `""`                                    |
| `lvmController.resizer.image.registry`              | Registry for csi-resizer image                                                   | `registry.k8s.io/`                      |
| `lvmController.resizer.image.repository`            | Image repository for csi-resizer                                                 | `sig-storage/csi-resizer`               |
| `lvmController.resizer.image.pullPolicy`            | Image pull policy for csi-resizer                                                | `IfNotPresent`                          |
| `lvmController.resizer.image.tag`                   | Image tag for csi-resizer                                                        | `v1.11.2`                                |
| `lvmController.snapshotter.image.registry`          | Registry for csi-snapshotter image                                               | `registry.k8s.io/`                      |
| `lvmController.snapshotter.image.repository`        | Image repository for csi-snapshotter                                             | `sig-storage/csi-snapshotter`           |
| `lvmController.snapshotter.image.pullPolicy`        | Image pull policy for csi-snapshotter                                            | `IfNotPresent`                          |
| `lvmController.snapshotter.image.tag`               | Image tag for csi-snapshotter                                                    | `v7.0.0`                                |
| `lvmController.snapshotController.image.registry`   | Registry for snapshot-controller image                                           | `registry.k8s.io/`                      |
| `lvmController.snapshotController.image.repository` | Image repository for snapshot-controller                                         | `sig-storage/snapshot-controller`       |
| `lvmController.snapshotController.image.pullPolicy` | Image pull policy for snapshot-controller                                        | `IfNotPresent`                          |
| `lvmController.snapshotController.image.tag`        | Image tag for snapshot-controller                                                | `v7.0.0`                                |
| `lvmController.provisioner.image.registry`          | Registry for csi-provisioner image                                               | `registry.k8s.io/`                      |
| `lvmController.provisioner.image.repository`        | Image repository for csi-provisioner                                             | `sig-storage/csi-provisioner`           |
| `lvmController.provisioner.image.pullPolicy`        | Image pull policy for csi-provisioner                                            | `IfNotPresent`                          |
| `lvmController.provisioner.image.tag`               | Image tag for csi-provisioner                                                    | `v5.2.0`                                |
| `lvmController.updateStrategy.type`                 | Update strategy for lvm localpv controller deployment                            | `RollingUpdate`                         |
| `lvmController.annotations`                         | Annotations for lvm localpv controller deployment metadata                       | `""`                                    |
| `lvmController.podAnnotations`                      | Annotations for lvm localpv controller deployment's pods metadata                | `""`                                    |
| `lvmController.resources`                           | Resource and request and limit for lvm localpv controller deployment containers  | `""`                                    |
| `lvmController.labels`                              | Labels for lvm localpv controller deployment metadata                            | `""`                                    |
| `lvmController.podLabels`                           | Appends labels to the lvm localpv controller deployment pods                     | `""`                                    |
| `lvmController.nodeSelector`                        | Nodeselector for lvm localpv controller deployment pods                          | `""`                                    |
| `lvmController.tolerations`                         | lvm localpv controller deployment's pod toleration values                        | `""`                                    |
| `lvmController.topologySpreadConstraints`           | lvm localpv controller deployment's pod topologySpreadConstraints values         | `""`                                    |
| `lvmController.securityContext`                     | Security context for lvm localpv controller deployment container                 | `""`                                    |
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
