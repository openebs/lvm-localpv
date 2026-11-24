#!/usr/bin/env bash

# Runs Red Hat's preflight container certification checks on the lvm-driver image.
# Needs preflight, skopeo and jq, e.g. from nix-shell nix/preflight.nix.
#
# Usage: preflight.sh IMAGE
#   IMAGE  an image reference, an OCI archive (e.g. the CI lvm-driver.tar),
#          or docker-daemon:<image> for an image in the local docker daemon.
#
# Env:
#   PREFLIGHT_PLATFORMS         platforms to check (default: the docker host's)
#   PREFLIGHT_ALLOWED_FAILURES  checks allowed to fail (default: RunAsNonRoot, the node plugin must run as root)
#   PREFLIGHT_ARTIFACTS         where preflight writes its logs and results (default: $TMPDIR/lvm-preflight)
#   PREFLIGHT_REGISTRY_PORT     port of the throwaway local registry (default: 5000)

set -euo pipefail

PLATFORMS="${PREFLIGHT_PLATFORMS:-linux/$(docker version -f '{{.Server.Arch}}')}"
ALLOWED_FAILURES="${PREFLIGHT_ALLOWED_FAILURES:-RunAsNonRoot}"
ARTIFACTS="${PREFLIGHT_ARTIFACTS:-${TMPDIR:-/tmp}/lvm-preflight}"
REGISTRY_PORT="${PREFLIGHT_REGISTRY_PORT:-5000}"
REGISTRY="lvm-preflight-registry"
IMAGE="${1:?usage: $0 <image reference | OCI archive | docker-daemon:image>}"

# preflight only pulls from registries, so local images go through a throwaway one.
insecure=()
if [ -f "$IMAGE" ] || [[ "$IMAGE" == docker-daemon:* ]]; then
  docker run -d --rm -p "127.0.0.1:$REGISTRY_PORT:5000" --name "$REGISTRY" registry:2 >/dev/null
  trap 'docker stop "$REGISTRY" >/dev/null' EXIT
  ref="localhost:$REGISTRY_PORT/lvm-driver:preflight"
  if [ -f "$IMAGE" ]; then
    # The CI archive has one entry per pushed name (same image), so pick one by its tag, if tagged.
    name=$(tar -xOf "$IMAGE" index.json | jq -r '.manifests[0].annotations["org.opencontainers.image.ref.name"] // empty')
    skopeo copy -q --all --dest-tls-verify=false "oci-archive:$IMAGE${name:+:$name}" "docker://$ref"
  else
    skopeo copy -q --dest-tls-verify=false "$IMAGE" "docker://$ref"
  fi
  IMAGE="$ref"
  insecure=(--insecure)
fi

rm -rf "$ARTIFACTS"
mkdir -p "$ARTIFACTS"
failed=0
for platform in ${PLATFORMS//,/ }; do
  arch="${platform#linux/}"
  mkdir -p "$ARTIFACTS/$arch"
  echo "Running preflight on $IMAGE for $arch"
  PFLT_ARTIFACTS="$ARTIFACTS" PFLT_LOGFILE="$ARTIFACTS/$arch/preflight.log" \
    preflight check container "$IMAGE" --platform "$arch" "${insecure[@]}" >/dev/null
  # preflight exits 0 even when checks fail, so go by the results instead.
  unexpected=$(jq -r --arg allowed "$ALLOWED_FAILURES" \
    '($allowed | split(",")) as $ok | (.results.failed + .results.errors)[].name | select(IN($ok[]) | not)' "$ARTIFACTS/$arch/results.json")
  if [ -n "$unexpected" ]; then
    echo "preflight failed for $arch: ${unexpected//$'\n'/, }"
    grep -m 20 'found disallowed modification' "$ARTIFACTS/$arch/preflight.log" || true
    failed=1
  else
    echo "preflight passed for $arch (allowed failures: $ALLOWED_FAILURES)"
  fi
done
exit $failed
