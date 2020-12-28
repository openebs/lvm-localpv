---
name: Bug report
about: Tell us about a problem you are experiencing
labels: Bug

---

**What steps did you take and what happened:**
[A clear and concise description of what the bug is, and what commands you ran.]


**What did you expect to happen:**


**The output of the following commands will help us better understand what's going on**:
(Pasting long output into a [GitHub gist](https://gist.github.com) or other [Pastebin](https://pastebin.com/) is fine.)

* `kubectl logs -f openebs-lvm-controller-0 -n kube-system -c openebs-lvm-plugin`
* `kubectl logs -f openebs-lvm-node-[xxxx] -n kube-system -c openebs-lvm-plugin`
* `kubectl get pods -n kube-system`
* `kubectl get lvmvol -A -o yaml`

**Anything else you would like to add:**
[Miscellaneous information that will assist in solving the issue.]


**Environment:**
- LVM Driver version
- Kubernetes version (use `kubectl version`):
- Kubernetes installer & version:
- Cloud provider or hardware configuration:
- OS (e.g. from `/etc/os-release`):
