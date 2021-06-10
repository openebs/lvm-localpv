# OpenEBS LVM Local PV Helm Repository

<img width="300" align="right" alt="OpenEBS Logo" src="https://raw.githubusercontent.com/cncf/artwork/master/projects/openebs/stacked/color/openebs-stacked-color.png" xmlns="http://www.w3.org/1999/html">

[Helm3](https://helm.sh) must be installed to use the charts.
Please refer to Helm's [documentation](https://helm.sh/docs/) to get started.

Once Helm is set up properly, add the repo as follows:

```bash
helm repo add openebs-lvmlocalpv https://openebs.github.io/lvm-localpv
```

You can then run `helm search repo openebs-lvmlocalpv` to see the charts.

#### Update OpenEBS LVM LocalPV Repo

Once OpenEBS LVM Localpv repository has been successfully fetched into the local system, it has to be updated to get the latest version. The LVM LocalPV charts repo can be updated using the following command.

```bash
helm repo update
```

#### Install using Helm 3

- Assign openebs namespace to the current context:
```bash
kubectl config set-context <current_context_name> --namespace=openebs
```

- If namespace is not created, run the following command
```bash
helm install <your-relase-name> openebs-lvmlocalpv/lvm-localpv --create-namespace
```
- Else, if namespace is already created, run the following command
```bash
helm install <your-relase-name> openebs-lvmlocalpv/lvm-localpv
```

_See [configuration](#configuration) below._

_See [helm install](https://helm.sh/docs/helm/helm_install/) for command documentation._

## Uninstall Chart

```console
helm uninstall [RELEASE_NAME]
```

This removes all the Kubernetes components associated with the chart and deletes the release.

_See [helm uninstall](https://helm.sh/docs/helm/helm_uninstall/) for command documentation._

## Upgrading Chart

```console
helm upgrade [RELEASE_NAME] [CHART] --install
```

For more details on installing OpenEBS LVM Local PV please see the [chart readme](https://github.com/openebs/lvm-localpv/blob/master/deploy/helm/charts/README.md).
