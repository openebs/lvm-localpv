#!/usr/bin/env bash

###############################################################################
# OpenEBS LVM LocalPV Uninstallation Script
#
# This script uninstalls OpenEBS LVM LocalPV from a Kubernetes cluster.
#
# Usage:
#   ./uninstall.sh [OPTIONS]
#
# Options:
#   --help                  Show this help message
#   --namespace NAMESPACE   Kubernetes namespace (default: openebs)
#   --release RELEASE       Helm release name (default: openebs-lvmlocalpv)
#   --keep-crds            Keep CRDs after uninstallation
#   --keep-storageclass    Keep StorageClass resources
#   --delete-pvcs          Delete all PVCs using this driver (dangerous!)
#   --force                Force uninstallation without confirmation
#   --log-level LEVEL      Set log level (INFO, DEBUG, WARN, ERROR)
#
###############################################################################

set -euo pipefail

# Script metadata
readonly SCRIPT_NAME="uninstall.sh"
readonly SCRIPT_VERSION="1.0.0"
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default values
readonly DEFAULT_NAMESPACE="openebs"
readonly DEFAULT_RELEASE="openebs-lvmlocalpv"
readonly DEFAULT_LOG_LEVEL="INFO"

# Global variables
OPENEBS_KUBE_NAMESPACE="${OPENEBS_KUBE_NAMESPACE:-${DEFAULT_NAMESPACE}}"
RELEASE="${OPENEBS_RELEASE:-${DEFAULT_RELEASE}}"
KEEP_CRDS="${KEEP_CRDS:-false}"
KEEP_STORAGECLASS="${KEEP_STORAGECLASS:-false}"
DELETE_PVCS="${DELETE_PVCS:-false}"
FORCE="${FORCE:-false}"
LOG_LEVEL="${LOG_LEVEL:-${DEFAULT_LOG_LEVEL}}"

# Log file path
if [[ -w /tmp ]]; then
  UNINSTALL_LOG_PATH="/tmp/openebs-lvmlocalpv_uninstall-$(date +'%Y-%m-%d_%H-%M-%S').log"
else
  UNINSTALL_LOG_PATH="./openebs-lvmlocalpv_uninstall-$(date +'%Y-%m-%d_%H-%M-%S').log"
fi

###############################################################################
# Utility Functions
###############################################################################

log() {
  local level="$1"
  shift
  local message="$*"
  local timestamp
  timestamp=$(date +'%Y-%m-%dT%H:%M:%S%z')
  
  case "${LOG_LEVEL}" in
    ERROR)
      [[ "${level}" != "ERROR" ]] && return 0
      ;;
    WARN)
      [[ "${level}" == "DEBUG" || "${level}" == "INFO" ]] && return 0
      ;;
    INFO)
      [[ "${level}" == "DEBUG" ]] && return 0
      ;;
    DEBUG)
      ;;
  esac
  
  echo "[${level}][${timestamp}]: ${message}" | tee -a "${UNINSTALL_LOG_PATH}"
}

info() {
  log "INFO" "$@"
}

warn() {
  log "WARN" "$@"
}

error() {
  log "ERROR" "$@"
  exit 1
}

debug() {
  log "DEBUG" "$@"
}

installed() {
  command -v "$1" >/dev/null 2>&1
}

check_prerequisites() {
  if ! installed kubectl; then
    error "kubectl is required but not installed. Please install kubectl first."
  fi
  
  if ! installed helm; then
    error "helm is required but not installed. Please install helm first."
  fi
  
  if ! kubectl cluster-info &>/dev/null; then
    error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
  fi
}

init_log() {
  touch "${UNINSTALL_LOG_PATH}" || error "Cannot create log file: ${UNINSTALL_LOG_PATH}"
  info "Log file: ${UNINSTALL_LOG_PATH}"
  debug "Script version: ${SCRIPT_VERSION}"
}

show_help() {
  cat <<EOF
OpenEBS LVM LocalPV Uninstallation Script v${SCRIPT_VERSION}

USAGE:
    ${SCRIPT_NAME} [OPTIONS]

OPTIONS:
    --help                  Show this help message and exit
    --namespace NAMESPACE   Kubernetes namespace (default: ${DEFAULT_NAMESPACE})
    --release RELEASE       Helm release name (default: ${DEFAULT_RELEASE})
    --keep-crds            Keep CRDs after uninstallation
    --keep-storageclass    Keep StorageClass resources after uninstallation
    --delete-pvcs          Delete all PVCs using this driver (WARNING: data loss!)
    --force                Force uninstallation without confirmation prompt
    --log-level LEVEL       Set log level: DEBUG, INFO, WARN, ERROR (default: ${DEFAULT_LOG_LEVEL})

ENVIRONMENT VARIABLES:
    OPENEBS_KUBE_NAMESPACE     Kubernetes namespace (default: ${DEFAULT_NAMESPACE})
    OPENEBS_RELEASE            Helm release name (default: ${DEFAULT_RELEASE})
    KEEP_CRDS                  Set to "true" to keep CRDs
    KEEP_STORAGECLASS          Set to "true" to keep StorageClass resources
    DELETE_PVCS                Set to "true" to delete PVCs (dangerous!)
    FORCE                      Set to "true" to skip confirmation

EXAMPLES:
    # Standard uninstallation
    ./uninstall.sh

    # Uninstall from custom namespace
    ./uninstall.sh --namespace my-namespace

    # Uninstall and keep CRDs
    ./uninstall.sh --keep-crds

    # Force uninstall without confirmation
    ./uninstall.sh --force

    # Uninstall and delete all PVCs (dangerous!)
    ./uninstall.sh --delete-pvcs --force

WARNING:
    - Uninstalling will remove the LVM LocalPV driver from your cluster
    - PVCs will become unusable but data will remain on nodes
    - Use --delete-pvcs only if you want to delete PVCs (data loss!)
    - CRDs are kept by default to preserve LVMVolume/LVMSnapshot resources

For more information, visit: https://github.com/openebs/lvm-localpv
EOF
}

parse_args() {
  while [[ $# -gt 0 ]]; do
    case $1 in
      --help|-h)
        show_help
        exit 0
        ;;
      --namespace)
        OPENEBS_KUBE_NAMESPACE="$2"
        shift 2
        ;;
      --release)
        RELEASE="$2"
        shift 2
        ;;
      --keep-crds)
        KEEP_CRDS=true
        shift
        ;;
      --keep-storageclass)
        KEEP_STORAGECLASS=true
        shift
        ;;
      --delete-pvcs)
        DELETE_PVCS=true
        shift
        ;;
      --force)
        FORCE=true
        shift
        ;;
      --log-level)
        LOG_LEVEL="$2"
        shift 2
        ;;
      *)
        error "Unknown option: $1. Use --help for usage information."
        ;;
    esac
  done
}

confirm_uninstall() {
  if [[ "${FORCE}" == "true" ]]; then
    return 0
  fi
  
  warn "This will uninstall OpenEBS LVM LocalPV from your cluster."
  warn "Release: ${RELEASE}"
  warn "Namespace: ${OPENEBS_KUBE_NAMESPACE}"
  
  if [[ "${DELETE_PVCS}" == "true" ]]; then
    warn "WARNING: --delete-pvcs is enabled. This will DELETE all PVCs using this driver!"
    warn "This action cannot be undone and will result in DATA LOSS!"
  fi
  
  echo -n "Are you sure you want to continue? (yes/no): "
  read -r response
  if [[ "${response}" != "yes" ]]; then
    info "Uninstallation cancelled."
    exit 0
  fi
}

check_helm_release() {
  if ! helm status "${RELEASE}" -n "${OPENEBS_KUBE_NAMESPACE}" &>/dev/null; then
    warn "Helm release ${RELEASE} not found in namespace ${OPENEBS_KUBE_NAMESPACE}"
    return 1
  fi
  return 0
}

list_pvcs() {
  info "Checking for PVCs using this driver..."
  local pvc_count
  pvc_count=$(kubectl get pvc --all-namespaces -o json 2>/dev/null | \
    jq -r '.items[] | select(.spec.storageClassName != null) | select(.spec.storageClassName | contains("openebs") or contains("lvm")) | .metadata.namespace + "/" + .metadata.name' 2>/dev/null | wc -l || echo "0")
  
  if [[ "${pvc_count}" -gt 0 ]]; then
    warn "Found ${pvc_count} PVC(s) that may be using this driver"
    if [[ "${DELETE_PVCS}" != "true" ]]; then
      info "PVCs will remain but become unusable. Use --delete-pvcs to delete them."
    fi
  else
    info "No PVCs found using this driver"
  fi
}

delete_pvcs() {
  if [[ "${DELETE_PVCS}" != "true" ]]; then
    return 0
  fi
  
  warn "Deleting PVCs using this driver..."
  local deleted=0
  local pvcs
  pvcs=$(kubectl get pvc --all-namespaces -o json 2>/dev/null | \
    jq -r '.items[] | select(.spec.storageClassName != null) | select(.spec.storageClassName | contains("openebs") or contains("lvm")) | .metadata.namespace + " " + .metadata.name' 2>/dev/null || echo "")
  
  if [[ -z "${pvcs}" ]]; then
    info "No PVCs to delete"
    return 0
  fi
  
  while IFS=' ' read -r namespace name; do
    if kubectl delete pvc "${name}" -n "${namespace}" --ignore-not-found &>/dev/null; then
      info "Deleted PVC: ${namespace}/${name}"
      ((deleted++))
    fi
  done <<< "${pvcs}"
  
  info "Deleted ${deleted} PVC(s)"
}

uninstall_helm_release() {
  info "Uninstalling Helm release: ${RELEASE}"
  
  if helm uninstall "${RELEASE}" -n "${OPENEBS_KUBE_NAMESPACE}" --wait --timeout 5m 2>&1 | tee -a "${UNINSTALL_LOG_PATH}"; then
    info "Helm release uninstalled successfully"
  else
    warn "Helm uninstall may have failed or release was already removed"
  fi
}

delete_crds() {
  if [[ "${KEEP_CRDS}" == "true" ]]; then
    info "Keeping CRDs as requested"
    return 0
  fi
  
  info "Deleting CRDs..."
  local crds=(
    "lvmvolumes.local.openebs.io"
    "lvmnodes.local.openebs.io"
    "lvmsnapshots.local.openebs.io"
  )
  
  for crd in "${crds[@]}"; do
    if kubectl get crd "${crd}" &>/dev/null; then
      if kubectl delete crd "${crd}" --ignore-not-found &>/dev/null; then
        info "Deleted CRD: ${crd}"
      else
        warn "Failed to delete CRD: ${crd} (may have dependent resources)"
      fi
    else
      debug "CRD ${crd} not found, skipping"
    fi
  done
}

delete_storageclasses() {
  if [[ "${KEEP_STORAGECLASS}" == "true" ]]; then
    info "Keeping StorageClass resources as requested"
    return 0
  fi
  
  info "Checking for StorageClass resources..."
  local sc_list
  sc_list=$(kubectl get storageclass -o json 2>/dev/null | \
    jq -r '.items[] | select(.provisioner == "local.csi.openebs.io") | .metadata.name' 2>/dev/null || echo "")
  
  if [[ -z "${sc_list}" ]]; then
    info "No StorageClass resources found with provisioner local.csi.openebs.io"
    return 0
  fi
  
  while IFS= read -r sc_name; do
    if [[ -n "${sc_name}" ]]; then
      warn "Deleting StorageClass: ${sc_name}"
      kubectl delete storageclass "${sc_name}" --ignore-not-found &>/dev/null || \
        warn "Failed to delete StorageClass: ${sc_name}"
    fi
  done <<< "${sc_list}"
}

cleanup_namespace() {
  info "Checking namespace: ${OPENEBS_KUBE_NAMESPACE}"
  
  # Check if namespace has any remaining resources
  local resource_count
  resource_count=$(kubectl get all -n "${OPENEBS_KUBE_NAMESPACE}" --no-headers 2>/dev/null | wc -l || echo "0")
  
  if [[ "${resource_count}" -eq 0 ]]; then
    info "Namespace ${OPENEBS_KUBE_NAMESPACE} appears to be empty"
    echo -n "Do you want to delete the namespace? (yes/no): "
    if [[ "${FORCE}" == "true" ]]; then
      response="no"
    else
      read -r response
    fi
    
    if [[ "${response}" == "yes" ]]; then
      if kubectl delete namespace "${OPENEBS_KUBE_NAMESPACE}" --ignore-not-found &>/dev/null; then
        info "Deleted namespace: ${OPENEBS_KUBE_NAMESPACE}"
      fi
    fi
  else
    info "Namespace ${OPENEBS_KUBE_NAMESPACE} still contains resources, not deleting"
  fi
}

verify_uninstall() {
  info "Verifying uninstallation..."
  
  if helm status "${RELEASE}" -n "${OPENEBS_KUBE_NAMESPACE}" &>/dev/null; then
    warn "Helm release ${RELEASE} still exists"
    return 1
  fi
  
  local pod_count
  pod_count=$(kubectl get pods -n "${OPENEBS_KUBE_NAMESPACE}" -l 'app in (openebs-lvm-controller,openebs-lvm-node)' --no-headers 2>/dev/null | wc -l || echo "0")
  
  if [[ "${pod_count}" -gt 0 ]]; then
    warn "Found ${pod_count} pod(s) still running"
    return 1
  fi
  
  info "Uninstallation verified successfully"
  return 0
}

main() {
  parse_args "$@"
  
  init_log
  info "OpenEBS LVM LocalPV Uninstallation Script v${SCRIPT_VERSION}"
  
  check_prerequisites
  confirm_uninstall
  
  if ! check_helm_release; then
    warn "Helm release not found. Proceeding with cleanup of remaining resources..."
  else
    list_pvcs
    delete_pvcs
    uninstall_helm_release
  fi
  
  delete_storageclasses
  delete_crds
  cleanup_namespace
  
  if verify_uninstall; then
    info "Uninstallation completed successfully!"
  else
    warn "Uninstallation completed with warnings. Please check the logs."
  fi
  
  info "Next steps:"
  info "  1. Verify all resources are removed: kubectl get all -n ${OPENEBS_KUBE_NAMESPACE}"
  if [[ "${KEEP_CRDS}" != "true" ]]; then
    info "  2. Verify CRDs are removed: kubectl get crd | grep openebs"
  fi
  if [[ "${KEEP_STORAGECLASS}" != "true" ]]; then
    info "  3. Verify StorageClasses are removed: kubectl get storageclass"
  fi
}

main "$@"
