#!/usr/bin/env bash

###############################################################################
# OpenEBS LVM LocalPV Upgrade Script
#
# This script upgrades OpenEBS LVM LocalPV to a new version.
#
# Usage:
#   ./upgrade.sh [OPTIONS]
#
# Options:
#   --help                  Show this help message
#   --namespace NAMESPACE   Kubernetes namespace (default: openebs)
#   --release RELEASE       Helm release name (default: openebs-lvmlocalpv)
#   --chart-version VERSION Specify target chart version
#   --offline               Use offline mode with local chart
#   --chart-dir DIR         Path to local chart directory (for offline)
#   --dry-run               Preview upgrade without executing
#   --force                 Force upgrade without confirmation
#   --log-level LEVEL      Set log level (INFO, DEBUG, WARN, ERROR)
#
###############################################################################

set -euo pipefail

# Script metadata
readonly SCRIPT_NAME="upgrade.sh"
readonly SCRIPT_VERSION="1.0.0"
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
readonly CHART_DIR="${ROOT_DIR}/deploy/helm/charts"
readonly CHART_YAML="${CHART_DIR}/Chart.yaml"
readonly INSTALL_SCRIPT_DIR="${SCRIPT_DIR}"

# Default values
readonly DEFAULT_NAMESPACE="openebs"
readonly DEFAULT_RELEASE="openebs-lvmlocalpv"
readonly DEFAULT_CHART="openebs-lvmlocalpv/lvm-localpv"
readonly DEFAULT_TIMEOUT="600s"
readonly DEFAULT_LOG_LEVEL="INFO"

# Global variables
OPENEBS_KUBE_NAMESPACE="${OPENEBS_KUBE_NAMESPACE:-${DEFAULT_NAMESPACE}}"
RELEASE="${OPENEBS_RELEASE:-${DEFAULT_RELEASE}}"
CHART="${DEFAULT_CHART}"
CHART_VERSION="${OPENEBS_CHART_VERSION:-}"
OFFLINE_INSTALL="${OFFLINE_INSTALL:-false}"
OPENEBS_CHART_DIR="${OPENEBS_CHART_DIR:-}"
DRY_RUN="${DRY_RUN:-false}"
FORCE="${FORCE:-false}"
LOG_LEVEL="${LOG_LEVEL:-${DEFAULT_LOG_LEVEL}}"
TIME_OUT_SECOND="${TIME_OUT_SECOND:-${DEFAULT_TIMEOUT}}"

# Log file path
if [[ -w /tmp ]]; then
  UPGRADE_LOG_PATH="/tmp/openebs-lvmlocalpv_upgrade-$(date +'%Y-%m-%d_%H-%M-%S').log"
else
  UPGRADE_LOG_PATH="./openebs-lvmlocalpv_upgrade-$(date +'%Y-%m-%d_%H-%M-%S').log"
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
  
  echo "[${level}][${timestamp}]: ${message}" | tee -a "${UPGRADE_LOG_PATH}"
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
    error "kubectl is required but not installed."
  fi
  
  if ! installed helm; then
    error "helm is required but not installed."
  fi
  
  if ! kubectl cluster-info &>/dev/null; then
    error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
  fi
}

init_log() {
  touch "${UPGRADE_LOG_PATH}" || error "Cannot create log file: ${UPGRADE_LOG_PATH}"
  info "Log file: ${UPGRADE_LOG_PATH}"
  debug "Script version: ${SCRIPT_VERSION}"
}

show_help() {
  cat <<EOF
OpenEBS LVM LocalPV Upgrade Script v${SCRIPT_VERSION}

USAGE:
    ${SCRIPT_NAME} [OPTIONS]

OPTIONS:
    --help                  Show this help message and exit
    --namespace NAMESPACE   Kubernetes namespace (default: ${DEFAULT_NAMESPACE})
    --release RELEASE       Helm release name (default: ${DEFAULT_RELEASE})
    --chart-version VERSION Target chart version to upgrade to
    --offline               Use offline mode with local chart
    --chart-dir DIR         Path to local chart directory (for offline mode)
    --dry-run               Preview upgrade without executing
    --force                 Force upgrade without confirmation
    --log-level LEVEL       Set log level: DEBUG, INFO, WARN, ERROR (default: ${DEFAULT_LOG_LEVEL})

ENVIRONMENT VARIABLES:
    OPENEBS_KUBE_NAMESPACE     Kubernetes namespace (default: ${DEFAULT_NAMESPACE})
    OPENEBS_RELEASE            Helm release name (default: ${DEFAULT_RELEASE})
    OPENEBS_CHART_VERSION      Target chart version
    OFFLINE_INSTALL            Set to "true" for offline upgrade
    OPENEBS_CHART_DIR          Path to local chart directory
    OPENEBS_IMAGE_REGISTRY     Image registry for offline upgrade
    OPENEBS_LOAD_IMAGES        Set to "true" to automatically load images from offline media
    OPENEBS_IMAGES_DIR         Path to images directory (for offline upgrade)
    DRY_RUN                    Set to "true" for dry-run mode
    FORCE                      Set to "true" to skip confirmation

EXAMPLES:
    # Upgrade to latest version
    ./upgrade.sh

    # Upgrade to specific version
    ./upgrade.sh --chart-version 1.8.0

    # Offline upgrade
    export OFFLINE_INSTALL=true
    export OPENEBS_CHART_DIR="./charts"
    ./upgrade.sh --offline

    # Dry run to preview
    ./upgrade.sh --dry-run

    # Force upgrade without confirmation
    ./upgrade.sh --force

NOTES:
    - The upgrade process will perform a rolling update
    - Existing PVCs and data will be preserved
    - It's recommended to backup your configuration before upgrading
    - Check release notes for breaking changes between versions

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
      --chart-version)
        CHART_VERSION="$2"
        shift 2
        ;;
      --offline)
        OFFLINE_INSTALL=true
        shift
        ;;
      --chart-dir)
        OPENEBS_CHART_DIR="$2"
        shift 2
        ;;
      --dry-run)
        DRY_RUN=true
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

# Detect and setup offline media directory
detect_offline_media() {
  if [[ "${OFFLINE_INSTALL}" != "true" ]]; then
    return 0
  fi
  
  # Check for standard offline-media directory structure
  local possible_dirs=(
    "./offline-media"
    "../offline-media"
    "${SCRIPT_DIR}/../offline-media"
    "${ROOT_DIR}/offline-media"
  )
  
  local media_dir=""
  for dir in "${possible_dirs[@]}"; do
    if [[ -d "${dir}" ]] && [[ -d "${dir}/charts" ]]; then
      media_dir="${dir}"
      break
    fi
  done
  
  if [[ -n "${media_dir}" ]]; then
    info "Detected offline media directory: ${media_dir}"
    
    # Auto-detect chart directory if not set
    if [[ -z "${OPENEBS_CHART_DIR:-}" ]]; then
      local chart_dir="${media_dir}/charts"
      
      # Check if chart_dir is a tgz file
      if [[ -f "${chart_dir}" ]] && [[ "${chart_dir}" == *.tgz ]]; then
        local chart_file="${chart_dir}"
        chart_dir=$(dirname "${chart_file}")
      # Look for chart tgz file matching target version
      elif [[ -d "${chart_dir}" ]]; then
        local chart_file
        if [[ -n "${CHART_VERSION}" ]]; then
          chart_file=$(find "${chart_dir}" -maxdepth 1 -name "lvm-localpv-${CHART_VERSION}.tgz" -type f | head -1)
        fi
        
        # If not found, get any chart file
        if [[ -z "${chart_file}" ]]; then
          chart_file=$(find "${chart_dir}" -maxdepth 1 -name "lvm-localpv-*.tgz" -type f | head -1)
        fi
      fi
      
      if [[ -n "${chart_file}" ]] && [[ -f "${chart_file}" ]]; then
        # Extract chart if needed
        local chart_name
        chart_name=$(basename "${chart_file}" .tgz)
        local chart_extracted="${chart_dir}/${chart_name}"
        
        if [[ ! -d "${chart_extracted}" ]] || [[ ! -f "${chart_extracted}/Chart.yaml" ]]; then
          info "Extracting chart: ${chart_file}"
          tar -xzf "${chart_file}" -C "${chart_dir}" || \
            error "Failed to extract chart"
        fi
        
        # Find the actual chart directory containing Chart.yaml
        # The extracted directory name might not match the tgz filename
        # (e.g., lvm-localpv-1.8.0.tgz might extract to lvm-localpv/)
        local actual_chart_dir=""
        
        # First check if the expected directory exists
        if [[ -f "${chart_extracted}/Chart.yaml" ]]; then
          actual_chart_dir="${chart_extracted}"
        # Check if there's a lvm-localpv subdirectory (without version)
        elif [[ -d "${chart_extracted}/lvm-localpv" ]] && [[ -f "${chart_extracted}/lvm-localpv/Chart.yaml" ]]; then
          actual_chart_dir="${chart_extracted}/lvm-localpv"
        # Check if tar extracted to a different directory name (e.g., lvm-localpv instead of lvm-localpv-1.8.0)
        elif [[ -d "${chart_dir}/lvm-localpv" ]] && [[ -f "${chart_dir}/lvm-localpv/Chart.yaml" ]]; then
          actual_chart_dir="${chart_dir}/lvm-localpv"
        # Search for Chart.yaml in the chart_dir (after extraction)
        else
          actual_chart_dir=$(find "${chart_dir}" -maxdepth 2 -name "Chart.yaml" -type f | \
            grep -v "/charts/" | grep -v "/crds/" | head -1 | xargs dirname 2>/dev/null || echo "")
        fi
        
        if [[ -n "${actual_chart_dir}" ]] && [[ -f "${actual_chart_dir}/Chart.yaml" ]]; then
          OPENEBS_CHART_DIR="${actual_chart_dir}"
          info "Using extracted chart directory: ${OPENEBS_CHART_DIR}"
        else
          # Debug: list what was extracted
          debug "Searching for Chart.yaml files in ${chart_dir}:"
          find "${chart_dir}" -name "Chart.yaml" -type f 2>/dev/null | while read -r f; do
            debug "  Found: ${f}"
          done
          debug "Chart directory contents:"
          ls -la "${chart_dir}" 2>/dev/null | head -10 | while read -r line; do
            debug "  ${line}"
          done
          error "Failed to find Chart.yaml in extracted chart: ${chart_file}"
        fi
      elif [[ -d "${chart_dir}" ]] && [[ -f "${chart_dir}/Chart.yaml" ]]; then
        # Chart directory already contains Chart.yaml
        OPENEBS_CHART_DIR="${chart_dir}"
        info "Using chart directory: ${OPENEBS_CHART_DIR}"
      fi
    fi
  fi
}

# Load images from offline media
load_offline_images() {
  if [[ "${OFFLINE_INSTALL}" != "true" ]]; then
    return 0
  fi
  
  if [[ "${OPENEBS_LOAD_IMAGES:-}" != "true" ]]; then
    return 0
  fi
  
  # Find images directory
  local images_dir=""
  local possible_dirs=(
    "./offline-media/images"
    "../offline-media/images"
    "${SCRIPT_DIR}/../offline-media/images"
    "${ROOT_DIR}/offline-media/images"
    "${OPENEBS_IMAGES_DIR:-}"
  )
  
  for dir in "${possible_dirs[@]}"; do
    if [[ -d "${dir}" ]] && [[ -n "$(find "${dir}" -name "*.tar" -type f)" ]]; then
      images_dir="${dir}"
      break
    fi
  done
  
  if [[ -z "${images_dir}" ]]; then
    warn "OPENEBS_LOAD_IMAGES is set but no images directory found"
    return 0
  fi
  
  info "Loading images from: ${images_dir}"
  
  # Check for load-images.sh script
  local load_script="${INSTALL_SCRIPT_DIR}/load-images.sh"
  if [[ -f "${load_script}" ]]; then
    info "Using load-images.sh script to load images"
    if [[ "${DRY_RUN}" == "true" ]]; then
      debug "Would run: ${load_script} --images-dir ${images_dir}"
    else
      if bash "${load_script}" --images-dir "${images_dir}" 2>&1 | tee -a "${UPGRADE_LOG_PATH}"; then
        info "Images loaded successfully"
      else
        warn "Failed to load images. You may need to load them manually."
      fi
    fi
  else
    warn "load-images.sh not found, skipping image loading"
  fi
}

check_helm_release() {
  if ! helm status "${RELEASE}" -n "${OPENEBS_KUBE_NAMESPACE}" &>/dev/null; then
    error "Helm release ${RELEASE} not found in namespace ${OPENEBS_KUBE_NAMESPACE}.
Please install it first using install.sh"
  fi
  
  local current_version
  current_version=$(helm list -n "${OPENEBS_KUBE_NAMESPACE}" -o json 2>/dev/null | \
    jq -r ".[] | select(.name == \"${RELEASE}\") | .chart" 2>/dev/null || echo "")
  
  if [[ -n "${current_version}" ]]; then
    info "Current version: ${current_version}"
  fi
}

get_chart_version() {
  if [[ -f "${CHART_YAML}" ]]; then
    if installed yq; then
      yq e '.version' "${CHART_YAML}" 2>/dev/null || echo ""
    elif installed grep; then
      grep -E '^version:' "${CHART_YAML}" | awk '{print $2}' | tr -d '"' || echo ""
    fi
  fi
}

confirm_upgrade() {
  if [[ "${FORCE}" == "true" ]]; then
    return 0
  fi
  
  warn "This will upgrade OpenEBS LVM LocalPV in your cluster."
  warn "Release: ${RELEASE}"
  warn "Namespace: ${OPENEBS_KUBE_NAMESPACE}"
  
  if [[ -n "${CHART_VERSION}" ]]; then
    warn "Target version: ${CHART_VERSION}"
  else
    warn "Target version: latest available"
  fi
  
  echo -n "Are you sure you want to continue? (yes/no): "
  read -r response
  if [[ "${response}" != "yes" ]]; then
    info "Upgrade cancelled."
    exit 0
  fi
}

backup_configuration() {
  info "Backing up current configuration..."
  
  local backup_dir="/tmp/openebs-lvmlocalpv-backup-$(date +'%Y%m%d-%H%M%S')"
  mkdir -p "${backup_dir}"
  
  # Backup Helm values
  if helm get values "${RELEASE}" -n "${OPENEBS_KUBE_NAMESPACE}" &>/dev/null; then
    helm get values "${RELEASE}" -n "${OPENEBS_KUBE_NAMESPACE}" > "${backup_dir}/values.yaml" || true
    info "Backed up Helm values to: ${backup_dir}/values.yaml"
  fi
  
  # Backup current version
  helm list -n "${OPENEBS_KUBE_NAMESPACE}" -o json > "${backup_dir}/releases.json" 2>/dev/null || true
  
  info "Configuration backed up to: ${backup_dir}"
}

online_upgrade() {
  info "Starting online upgrade..."
  
  # Update helm repo
  info "Updating Helm repository..."
  if [[ "${DRY_RUN}" != "true" ]]; then
    helm repo update openebs-lvmlocalpv 2>/dev/null || \
      error "Failed to update Helm repository. Check network connectivity."
  fi
  
  # Determine target version
  if [[ -z "${CHART_VERSION}" ]]; then
    CHART_VERSION=$(helm search repo openebs-lvmlocalpv/lvm-localpv --versions --output json 2>/dev/null | \
      grep -oP '"version":\s*"\K[^"]+' | head -1 || echo "")
    if [[ -z "${CHART_VERSION}" ]]; then
      warn "Could not determine latest chart version, upgrading to latest"
      CHART_VERSION=""
    else
      info "Upgrading to chart version: ${CHART_VERSION}"
    fi
  fi
  
  # Build upgrade command
  local helm_args=(
    "--namespace" "${OPENEBS_KUBE_NAMESPACE}"
    "--timeout" "${TIME_OUT_SECOND}"
    "--wait"
  )
  
  if [[ -n "${CHART_VERSION}" ]]; then
    helm_args+=("--version" "${CHART_VERSION}")
  fi
  
  info "Upgrading ${RELEASE} (this may take several minutes)..."
  if [[ "${DRY_RUN}" == "true" ]]; then
    debug "Would run: helm upgrade ${RELEASE} ${CHART} ${helm_args[*]}"
    info "Dry-run mode: Skipping actual upgrade"
  else
    if helm upgrade "${RELEASE}" "${CHART}" "${helm_args[@]}" 2>&1 | tee -a "${UPGRADE_LOG_PATH}"; then
      info "Helm upgrade completed successfully"
    else
      error "Helm upgrade failed. Check logs for details."
    fi
  fi
}

offline_upgrade() {
  info "Starting offline upgrade..."
  
  if [[ -z "${OPENEBS_CHART_DIR}" ]]; then
    error "OPENEBS_CHART_DIR must be set for offline upgrade"
  fi
  
  # Resolve relative path
  local chart_dir="${OPENEBS_CHART_DIR}"
  if [[ ! "${chart_dir}" =~ ^/ ]]; then
    chart_dir="$(cd "$(dirname "${chart_dir}")" 2>/dev/null && pwd)/$(basename "${chart_dir}")" || \
      chart_dir="$(realpath "${OPENEBS_CHART_DIR}" 2>/dev/null || echo "${OPENEBS_CHART_DIR}")"
  fi
  
  # Check if it's a tgz file
  if [[ -f "${chart_dir}" ]] && [[ "${chart_dir}" == *.tgz ]]; then
    local chart_file="${chart_dir}"
    chart_dir=$(dirname "${chart_file}")
    local chart_name
    chart_name=$(basename "${chart_file}" .tgz)
    local chart_extracted="${chart_dir}/${chart_name}"
    
    if [[ ! -d "${chart_extracted}" ]] || [[ ! -f "${chart_extracted}/Chart.yaml" ]]; then
      info "Extracting chart: ${chart_file}"
      tar -xzf "${chart_file}" -C "${chart_dir}" || \
        error "Failed to extract chart"
    fi
    
    # Find the actual chart directory containing Chart.yaml
    # The extracted directory name might not match the tgz filename
    local actual_chart_dir=""
    
    # First check if the expected directory exists
    if [[ -f "${chart_extracted}/Chart.yaml" ]]; then
      actual_chart_dir="${chart_extracted}"
    # Check if there's a lvm-localpv subdirectory (without version)
    elif [[ -d "${chart_extracted}/lvm-localpv" ]] && [[ -f "${chart_extracted}/lvm-localpv/Chart.yaml" ]]; then
      actual_chart_dir="${chart_extracted}/lvm-localpv"
    # Check if tar extracted to a different directory name (e.g., lvm-localpv instead of lvm-localpv-1.8.0)
    elif [[ -d "${chart_dir}/lvm-localpv" ]] && [[ -f "${chart_dir}/lvm-localpv/Chart.yaml" ]]; then
      actual_chart_dir="${chart_dir}/lvm-localpv"
    # Search for Chart.yaml in the chart_dir (after extraction)
    else
      actual_chart_dir=$(find "${chart_dir}" -maxdepth 2 -name "Chart.yaml" -type f | \
        grep -v "/charts/" | grep -v "/crds/" | head -1 | xargs dirname 2>/dev/null || echo "")
    fi
    
    if [[ -n "${actual_chart_dir}" ]] && [[ -f "${actual_chart_dir}/Chart.yaml" ]]; then
      OPENEBS_CHART_DIR="${actual_chart_dir}"
      chart_dir="${actual_chart_dir}"
      info "Using extracted chart directory: ${OPENEBS_CHART_DIR}"
    else
      # Debug: list what was extracted
      debug "Searching for Chart.yaml files in ${chart_dir}:"
      find "${chart_dir}" -name "Chart.yaml" -type f 2>/dev/null | while read -r f; do
        debug "  Found: ${f}"
      done
      debug "Chart directory contents:"
      ls -la "${chart_dir}" 2>/dev/null | head -10 | while read -r line; do
        debug "  ${line}"
      done
      error "Failed to find Chart.yaml in extracted chart: ${chart_file}"
    fi
  # Check if it's a directory containing tgz file (but not Chart.yaml)
  elif [[ -d "${chart_dir}" ]] && [[ ! -f "${chart_dir}/Chart.yaml" ]]; then
    local chart_file
    chart_file=$(find "${chart_dir}" -maxdepth 1 -name "lvm-localpv-*.tgz" -type f | head -1)
    if [[ -n "${chart_file}" ]]; then
      local chart_name
      chart_name=$(basename "${chart_file}" .tgz)
      local chart_extracted="${chart_dir}/${chart_name}"
      
      if [[ ! -d "${chart_extracted}" ]] || [[ ! -f "${chart_extracted}/Chart.yaml" ]]; then
        info "Extracting chart: ${chart_file}"
        tar -xzf "${chart_file}" -C "${chart_dir}" || \
          error "Failed to extract chart"
      fi
      
      # Find the actual chart directory containing Chart.yaml
      # The extracted directory name might not match the tgz filename
      # (e.g., lvm-localpv-1.8.0.tgz might extract to lvm-localpv/)
      local actual_chart_dir=""
      
      # First check if the expected directory exists
      if [[ -f "${chart_extracted}/Chart.yaml" ]]; then
        actual_chart_dir="${chart_extracted}"
      # Check if there's a lvm-localpv subdirectory (without version)
      elif [[ -d "${chart_extracted}/lvm-localpv" ]] && [[ -f "${chart_extracted}/lvm-localpv/Chart.yaml" ]]; then
        actual_chart_dir="${chart_extracted}/lvm-localpv"
      # Check if tar extracted to a different directory name (e.g., lvm-localpv instead of lvm-localpv-1.8.0)
      elif [[ -d "${chart_dir}/lvm-localpv" ]] && [[ -f "${chart_dir}/lvm-localpv/Chart.yaml" ]]; then
        actual_chart_dir="${chart_dir}/lvm-localpv"
      # Search for Chart.yaml in the chart_dir (after extraction)
      else
        actual_chart_dir=$(find "${chart_dir}" -maxdepth 2 -name "Chart.yaml" -type f | \
          grep -v "/charts/" | grep -v "/crds/" | head -1 | xargs dirname 2>/dev/null || echo "")
      fi
      
      if [[ -n "${actual_chart_dir}" ]] && [[ -f "${actual_chart_dir}/Chart.yaml" ]]; then
        OPENEBS_CHART_DIR="${actual_chart_dir}"
        chart_dir="${actual_chart_dir}"
        info "Using extracted chart directory: ${OPENEBS_CHART_DIR}"
      else
        # Debug: list what was extracted
        debug "Searching for Chart.yaml files in ${chart_dir}:"
        find "${chart_dir}" -name "Chart.yaml" -type f 2>/dev/null | while read -r f; do
          debug "  Found: ${f}"
        done
        debug "Chart directory contents:"
        ls -la "${chart_dir}" 2>/dev/null | head -10 | while read -r line; do
          debug "  ${line}"
        done
        error "Failed to extract or find valid chart directory with Chart.yaml in: ${chart_dir}"
      fi
    else
      error "Chart directory does not contain Chart.yaml or chart tgz file: ${OPENEBS_CHART_DIR}"
    fi
  # Verify chart directory contains Chart.yaml
  elif [[ -d "${chart_dir}" ]]; then
    if [[ ! -f "${chart_dir}/Chart.yaml" ]]; then
      error "Chart directory does not contain Chart.yaml: ${OPENEBS_CHART_DIR}"
    fi
  else
    error "Chart directory does not exist: ${OPENEBS_CHART_DIR}"
  fi
  
  # Setup image registry
  local image_registry plugin_image_registry
  if [[ -z "${OPENEBS_IMAGE_REGISTRY:-}" ]]; then
    image_registry="registry.k8s.io/"
    plugin_image_registry=""
    warn "OPENEBS_IMAGE_REGISTRY not set, using default: ${image_registry}"
  else
    image_registry="${OPENEBS_IMAGE_REGISTRY%/}/"
    plugin_image_registry="${OPENEBS_IMAGE_REGISTRY%/}/"
  fi
  
  # Build upgrade command
  local helm_args=(
    "--namespace" "${OPENEBS_KUBE_NAMESPACE}"
    "--set-string" "lvmController.resizer.image.registry=${image_registry}"
    "--set-string" "lvmController.snapshotter.image.registry=${image_registry}"
    "--set-string" "lvmController.snapshotController.image.registry=${image_registry}"
    "--set-string" "lvmController.provisioner.image.registry=${image_registry}"
    "--set-string" "lvmNode.driverRegistrar.image.registry=${image_registry}"
    "--set-string" "lvmPlugin.image.registry=${plugin_image_registry}"
    "--timeout" "${TIME_OUT_SECOND}"
    "--wait"
  )
  
  info "Upgrading ${RELEASE} from local chart: ${OPENEBS_CHART_DIR}"
  if [[ "${DRY_RUN}" == "true" ]]; then
    debug "Would run: helm upgrade ${RELEASE} ${OPENEBS_CHART_DIR} ${helm_args[*]}"
    info "Dry-run mode: Skipping actual upgrade"
  else
    if helm upgrade "${RELEASE}" "${OPENEBS_CHART_DIR}" "${helm_args[@]}" 2>&1 | tee -a "${UPGRADE_LOG_PATH}"; then
      info "Helm upgrade completed successfully"
    else
      error "Helm upgrade failed. Check logs for details."
    fi
  fi
}

verify_upgrade() {
  info "Verifying upgrade..."
  
  if [[ "${DRY_RUN}" == "true" ]]; then
    info "Dry-run mode: Skipping verification"
    return 0
  fi
  
  # Check Helm release status
  local status
  status=$(helm status "${RELEASE}" -n "${OPENEBS_KUBE_NAMESPACE}" 2>/dev/null | grep ^STATUS: | awk '{print $2}' || echo "")
  if [[ "${status}" != "deployed" ]]; then
    error "Helm release status is '${status}', expected 'deployed'"
  fi
  info "Helm release status: ${status}"
  
  # Check pod status
  info "Checking pod status..."
  local controller_ready controller_desired
  controller_ready=$(kubectl get deployment -n "${OPENEBS_KUBE_NAMESPACE}" -l app=openebs-lvm-controller --no-headers 2>/dev/null | \
    awk '{print $4}' | cut -d'/' -f1 || echo "0")
  controller_desired=$(kubectl get deployment -n "${OPENEBS_KUBE_NAMESPACE}" -l app=openebs-lvm-controller --no-headers 2>/dev/null | \
    awk '{print $4}' | cut -d'/' -f2 || echo "0")
  
  if [[ "${controller_desired}" -gt 0 ]]; then
    info "Controller: ${controller_ready}/${controller_desired} replicas ready"
    if [[ "${controller_ready}" -lt "${controller_desired}" ]]; then
      warn "Controller pods are still rolling out"
    fi
  fi
  
  # Check daemonset
  local node_ready node_desired
  node_ready=$(kubectl get daemonset -n "${OPENEBS_KUBE_NAMESPACE}" -l app=openebs-lvm-node --no-headers 2>/dev/null | \
    awk '{print $4}' | cut -d'/' -f1 || echo "0")
  node_desired=$(kubectl get daemonset -n "${OPENEBS_KUBE_NAMESPACE}" -l app=openebs-lvm-node --no-headers 2>/dev/null | \
    awk '{print $4}' | cut -d'/' -f2 || echo "0")
  
  if [[ "${node_desired}" -gt 0 ]]; then
    info "Node daemonset: ${node_ready}/${node_desired} pods ready"
  fi
  
  # Show new version
  local new_version
  new_version=$(helm list -n "${OPENEBS_KUBE_NAMESPACE}" -o json 2>/dev/null | \
    jq -r ".[] | select(.name == \"${RELEASE}\") | .chart" 2>/dev/null || echo "")
  if [[ -n "${new_version}" ]]; then
    info "Upgraded to version: ${new_version}"
  fi
  
  info "Upgrade verification completed"
}

main() {
  parse_args "$@"
  
  init_log
  info "OpenEBS LVM LocalPV Upgrade Script v${SCRIPT_VERSION}"
  
  if [[ "${DRY_RUN}" == "true" ]]; then
    info "DRY-RUN MODE: No changes will be made"
  fi
  
  check_prerequisites
  check_helm_release
  
  # Detect offline media if in offline mode
  if [[ "${OFFLINE_INSTALL}" == "true" ]]; then
    detect_offline_media
    
    if [[ -z "${OPENEBS_CHART_DIR:-}" ]]; then
      error "OPENEBS_CHART_DIR must be set for offline upgrade.
Or place charts in ./offline-media/charts/ directory."
    fi
  fi
  
  confirm_upgrade
  
  if [[ "${DRY_RUN}" != "true" ]]; then
    backup_configuration
  fi
  
  # Load images if in offline mode and requested
  if [[ "${OFFLINE_INSTALL}" == "true" ]]; then
    load_offline_images
  fi
  
  if [[ "${OFFLINE_INSTALL}" == "true" ]]; then
    offline_upgrade
  else
    online_upgrade
  fi
  
  verify_upgrade
  
  info "Upgrade completed successfully!"
  info "Next steps:"
  info "  1. Verify pods are running: kubectl get pods -n ${OPENEBS_KUBE_NAMESPACE}"
  info "  2. Check for any issues: kubectl logs -n ${OPENEBS_KUBE_NAMESPACE} -l app=openebs-lvm-controller"
  info "  3. Test volume operations to ensure everything works correctly"
}

main "$@"
