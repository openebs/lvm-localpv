#!/usr/bin/env bash

###############################################################################
# OpenEBS LVM LocalPV Installation Script
#
# This script installs OpenEBS LVM LocalPV using Helm charts.
# It supports both online and offline installation modes.
#
# Usage:
#   ./install.sh [OPTIONS]
#
# Options:
#   --help                  Show this help message
#   --version              Show script version
#   --dry-run              Preview installation without executing
#   --offline              Enable offline installation mode
#   --chart-version VERSION Specify chart version (overrides auto-detection)
#   --namespace NAMESPACE  Kubernetes namespace (default: openebs)
#   --create-storageclass  Automatically create StorageClass
#   --log-level LEVEL      Set log level (INFO, DEBUG, WARN, ERROR)
#
# Environment Variables:
#   See help message for complete list of supported environment variables.
#
###############################################################################

set -euo pipefail

# Script metadata
readonly SCRIPT_NAME="install.sh"
readonly SCRIPT_VERSION="1.0.0"
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
readonly CHART_DIR="${ROOT_DIR}/deploy/helm/charts"
readonly CHART_YAML="${CHART_DIR}/Chart.yaml"

# Default values
readonly DEFAULT_CHART="openebs-lvmlocalpv/lvm-localpv"
readonly DEFAULT_RELEASE="openebs-lvmlocalpv"
readonly DEFAULT_TIMEOUT="600s"
readonly DEFAULT_NAMESPACE="openebs"
readonly DEFAULT_LOG_LEVEL="INFO"

# Global variables
CHART="${DEFAULT_CHART}"
RELEASE="${DEFAULT_RELEASE}"
TIME_OUT_SECOND="${DEFAULT_TIMEOUT}"
OFFLINE_INSTALL="${OFFLINE_INSTALL:-false}"
DRY_RUN="${DRY_RUN:-false}"
OPENEBS_KUBE_NAMESPACE="${OPENEBS_KUBE_NAMESPACE:-${DEFAULT_NAMESPACE}}"
OPENEBS_CREATE_STORAGECLASS="${OPENEBS_CREATE_STORAGECLASS:-false}"
OPENEBS_STORAGECLASS_YAML="${OPENEBS_STORAGECLASS_YAML:-}"
LOG_LEVEL="${LOG_LEVEL:-${DEFAULT_LOG_LEVEL}}"

# Resource limits (with defaults)
OPENEBS_CONTROLLER_RESOURCE_LIMITS_CPU="${OPENEBS_CONTROLLER_RESOURCE_LIMITS_CPU:-500m}"
OPENEBS_CONTROLLER_RESOURCE_LIMITS_MEMORY="${OPENEBS_CONTROLLER_RESOURCE_LIMITS_MEMORY:-512Mi}"
OPENEBS_CONTROLLER_RESOURCE_REQUESTS_CPU="${OPENEBS_CONTROLLER_RESOURCE_REQUESTS_CPU:-500m}"
OPENEBS_CONTROLLER_RESOURCE_REQUESTS_MEMORY="${OPENEBS_CONTROLLER_RESOURCE_REQUESTS_MEMORY:-512Mi}"
OPENEBS_NODE_RESOURCE_LIMITS_CPU="${OPENEBS_NODE_RESOURCE_LIMITS_CPU:-500m}"
OPENEBS_NODE_RESOURCE_LIMITS_MEMORY="${OPENEBS_NODE_RESOURCE_LIMITS_MEMORY:-512Mi}"
OPENEBS_NODE_RESOURCE_REQUESTS_CPU="${OPENEBS_NODE_RESOURCE_REQUESTS_CPU:-500m}"
OPENEBS_NODE_RESOURCE_REQUESTS_MEMORY="${OPENEBS_NODE_RESOURCE_REQUESTS_MEMORY:-512Mi}"

# Log file path (use /tmp or current directory if /tmp is not writable)
if [[ -w /tmp ]]; then
  INSTALL_LOG_PATH="/tmp/openebs-lvmlocalpv_install-$(date +'%Y-%m-%d_%H-%M-%S').log"
else
  INSTALL_LOG_PATH="./openebs-lvmlocalpv_install-$(date +'%Y-%m-%d_%H-%M-%S').log"
fi

# Chart version (will be auto-detected or set via env/arg)
CHART_VERSION="${OPENEBS_CHART_VERSION:-}"

# Track resources created for rollback
ROLLBACK_RESOURCES=()

###############################################################################
# Utility Functions
###############################################################################

# Detect system architecture
detect_architecture() {
  local arch
  arch=$(uname -m)
  case "${arch}" in
    x86_64)
      echo "amd64"
      ;;
    aarch64|arm64)
      echo "arm64"
      ;;
    *)
      echo "${arch}"
      ;;
  esac
}

# Detect operating system
detect_os() {
  uname -s | tr '[:upper:]' '[:lower:]'
}

# Logging functions with levels
log() {
  local level="$1"
  shift
  local message="$*"
  local timestamp
  timestamp=$(date +'%Y-%m-%dT%H:%M:%S%z')
  
  # Check if message should be logged based on log level
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
  
  echo "[${level}][${timestamp}]: ${message}" | tee -a "${INSTALL_LOG_PATH}"
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

# Check if command is installed
installed() {
  command -v "$1" >/dev/null 2>&1
}

# Check prerequisites with helpful error messages
check_prerequisites() {
  local missing_tools=()
  
  if ! installed kubectl; then
    missing_tools+=("kubectl")
  fi
  
  if ! installed helm; then
    missing_tools+=("helm")
  fi
  
  if ! installed curl; then
    missing_tools+=("curl")
  fi
  
  if ! installed envsubst; then
    missing_tools+=("envsubst")
  fi
  
  if [[ ${#missing_tools[@]} -gt 0 ]]; then
    error "Missing required tools: ${missing_tools[*]}

Please install the missing tools:
  - kubectl: https://kubernetes.io/docs/tasks/tools/install-kubectl/
  - helm: https://helm.sh/docs/intro/install/
  - curl: Usually pre-installed on Linux/macOS
  - envsubst: Part of gettext package (install via: apt-get install gettext-base or brew install gettext)"
  fi
  
  # Verify kubectl can connect to cluster
  if ! kubectl cluster-info &>/dev/null; then
    error "Cannot connect to Kubernetes cluster. Please check your kubeconfig."
  fi
  
  # Check Kubernetes version (minimum 1.23)
  local k8s_version
  k8s_version=$(kubectl version --client --short 2>/dev/null | grep -oP 'v\d+\.\d+' | head -1 || echo "")
  if [[ -n "${k8s_version}" ]]; then
    local major minor
    IFS='.' read -r major minor <<< "${k8s_version#v}"
    if [[ ${major} -lt 1 ]] || [[ ${major} -eq 1 && ${minor} -lt 23 ]]; then
      warn "Kubernetes version ${k8s_version} detected. Minimum required version is 1.23+"
    fi
  fi
}

# Get chart version from Chart.yaml
get_chart_version() {
  if [[ -f "${CHART_YAML}" ]]; then
    if installed yq; then
      yq e '.version' "${CHART_YAML}" 2>/dev/null || echo ""
    elif installed grep; then
      grep -E '^version:' "${CHART_YAML}" | awk '{print $2}' | tr -d '"' || echo ""
    fi
  fi
}

# Initialize log file
init_log() {
  touch "${INSTALL_LOG_PATH}" || error "Cannot create log file: ${INSTALL_LOG_PATH}"
  info "Log file: ${INSTALL_LOG_PATH}"
  debug "Script version: ${SCRIPT_VERSION}"
  debug "Chart directory: ${CHART_DIR}"
  debug "Architecture: $(detect_architecture)"
  debug "OS: $(detect_os)"
}

# Show help message
show_help() {
  cat <<EOF
OpenEBS LVM LocalPV Installation Script v${SCRIPT_VERSION}

USAGE:
    ${SCRIPT_NAME} [OPTIONS]

OPTIONS:
    --help                      Show this help message and exit
    --version                   Show script version and exit
    --dry-run                   Preview installation without executing
    --offline                   Enable offline installation mode
    --chart-version VERSION     Specify chart version (overrides auto-detection)
    --namespace NAMESPACE       Kubernetes namespace (default: ${DEFAULT_NAMESPACE})
    --create-storageclass       Automatically create StorageClass after installation
    --log-level LEVEL           Set log level: DEBUG, INFO, WARN, ERROR (default: ${DEFAULT_LOG_LEVEL})

ENVIRONMENT VARIABLES:
    Required:
        OPENEBS_CONTROLLER_NODE_NAMES    Comma-separated list of controller node names
        OPENEBS_DATA_NODE_NAMES          Comma-separated list of data node names
        OPENEBS_STORAGECLASS_NAME        Name for the StorageClass to create
        OPENEBS_VG_NAME                  LVM volume group name

    Optional:
        OFFLINE_INSTALL                  Set to "true" for offline installation
        OPENEBS_CHART_DIR                Path to local chart directory (for offline install)
        OPENEBS_IMAGE_REGISTRY           Image registry for offline installation
        OPENEBS_LOAD_IMAGES              Set to "true" to automatically load images from offline media
        OPENEBS_IMAGES_DIR               Path to images directory (for offline install)
        OPENEBS_KUBE_NAMESPACE           Kubernetes namespace (default: ${DEFAULT_NAMESPACE})
        OPENEBS_CREATE_STORAGECLASS      Set to "true" to create StorageClass
        OPENEBS_STORAGECLASS_YAML        Path to StorageClass YAML template
        OPENEBS_CHART_VERSION            Chart version (overrides auto-detection)
        
        Resource limits (all optional, with defaults):
        OPENEBS_CONTROLLER_RESOURCE_LIMITS_CPU
        OPENEBS_CONTROLLER_RESOURCE_LIMITS_MEMORY
        OPENEBS_CONTROLLER_RESOURCE_REQUESTS_CPU
        OPENEBS_CONTROLLER_RESOURCE_REQUESTS_MEMORY
        OPENEBS_NODE_RESOURCE_LIMITS_CPU
        OPENEBS_NODE_RESOURCE_LIMITS_MEMORY
        OPENEBS_NODE_RESOURCE_REQUESTS_CPU
        OPENEBS_NODE_RESOURCE_REQUESTS_MEMORY

EXAMPLES:
    # Online installation
    export OPENEBS_CONTROLLER_NODE_NAMES="master01,master02"
    export OPENEBS_DATA_NODE_NAMES="node01,node02"
    export OPENEBS_STORAGECLASS_NAME="openebs-lvmpv"
    export OPENEBS_VG_NAME="lvmvg"
    ./install.sh

    # Offline installation
    export OFFLINE_INSTALL=true
    export OPENEBS_CHART_DIR="./charts"
    export OPENEBS_IMAGE_REGISTRY="registry.example.com"
    ./install.sh --offline

    # Dry run to preview
    ./install.sh --dry-run

    # With custom namespace and StorageClass creation
    ./install.sh --namespace my-namespace --create-storageclass

For more information, visit: https://github.com/openebs/lvm-localpv
EOF
}

# Show version
show_version() {
  echo "${SCRIPT_NAME} v${SCRIPT_VERSION}"
  if [[ -f "${CHART_YAML}" ]]; then
    local chart_version
    chart_version=$(get_chart_version)
    if [[ -n "${chart_version}" ]]; then
      echo "Chart version: ${chart_version}"
    fi
  fi
}

# Parse command line arguments
parse_args() {
  while [[ $# -gt 0 ]]; do
    case $1 in
      --help|-h)
        show_help
        exit 0
        ;;
      --version|-v)
        show_version
        exit 0
        ;;
      --dry-run)
        DRY_RUN=true
        shift
        ;;
      --offline)
        OFFLINE_INSTALL=true
        shift
        ;;
      --chart-version)
        CHART_VERSION="$2"
        shift 2
        ;;
      --namespace)
        OPENEBS_KUBE_NAMESPACE="$2"
        shift 2
        ;;
      --create-storageclass)
        OPENEBS_CREATE_STORAGECLASS=true
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
    
    # Auto-detect or process chart directory
    if [[ -z "${OPENEBS_CHART_DIR:-}" ]]; then
      # Look for chart tgz file
      local chart_file
      chart_file=$(find "${media_dir}/charts" -name "lvm-localpv-*.tgz" -type f | head -1)
      if [[ -n "${chart_file}" ]]; then
        # Extract chart if needed
        local chart_name
        chart_name=$(basename "${chart_file}" .tgz)
        local chart_extracted="${media_dir}/charts/${chart_name}"
        
        if [[ ! -d "${chart_extracted}" ]]; then
          info "Extracting chart: ${chart_file}"
          tar -xzf "${chart_file}" -C "${media_dir}/charts" || \
            error "Failed to extract chart"
        fi
        
        # Check what was actually extracted
        # Helm charts can have different structures:
        # 1. lvm-localpv-1.8.0/lvm-localpv/Chart.yaml (version dir contains chart dir)
        # 2. lvm-localpv/Chart.yaml (direct chart dir at same level)
        local actual_chart_dir=""
        
        # First, check if Chart.yaml is directly in the extracted version directory
        if [[ -f "${chart_extracted}/Chart.yaml" ]]; then
          actual_chart_dir="${chart_extracted}"
        # Check if there's a lvm-localpv subdirectory inside the version directory
        elif [[ -d "${chart_extracted}/lvm-localpv" ]] && [[ -f "${chart_extracted}/lvm-localpv/Chart.yaml" ]]; then
          actual_chart_dir="${chart_extracted}/lvm-localpv"
        # Check if there's a lvm-localpv directory at the same level (extracted alongside)
        elif [[ -d "${media_dir}/charts/lvm-localpv" ]] && [[ -f "${media_dir}/charts/lvm-localpv/Chart.yaml" ]]; then
          actual_chart_dir="${media_dir}/charts/lvm-localpv"
        # Search for Chart.yaml but exclude crds and charts subdirs
        else
          actual_chart_dir=$(find "${media_dir}/charts" -maxdepth 3 -name "Chart.yaml" -type f | \
            grep -v "/charts/" | grep -v "/crds/" | grep -v "crds" | head -1 | xargs dirname 2>/dev/null || echo "")
        fi
        
        if [[ -n "${actual_chart_dir}" ]] && [[ -f "${actual_chart_dir}/Chart.yaml" ]]; then
          OPENEBS_CHART_DIR="${actual_chart_dir}"
          info "Using chart directory: ${OPENEBS_CHART_DIR}"
        else
          warn "Could not find Chart.yaml in extracted directory: ${chart_extracted}"
          debug "Searching for Chart.yaml files:"
          find "${media_dir}/charts" -name "Chart.yaml" -type f 2>/dev/null | while read -r f; do
            debug "  Found: ${f}"
          done
          debug "Directory contents of ${media_dir}/charts:"
          ls -la "${media_dir}/charts" 2>/dev/null | head -10 | while read -r line; do
            debug "  ${line}"
          done
        fi
      fi
    else
      # OPENEBS_CHART_DIR is set, but check if it points to tgz file or directory with tgz
      local chart_path="${OPENEBS_CHART_DIR}"
      
      # Resolve relative path
      if [[ ! "${chart_path}" =~ ^/ ]]; then
        chart_path="$(cd "$(dirname "${chart_path}")" && pwd)/$(basename "${chart_path}")"
      fi
      
      # Check if it's a tgz file
      if [[ -f "${chart_path}" ]] && [[ "${chart_path}" =~ \.tgz$ ]]; then
        local chart_dir
        chart_dir=$(dirname "${chart_path}")
        local chart_name
        chart_name=$(basename "${chart_path}" .tgz)
        local chart_extracted="${chart_dir}/${chart_name}"
        
        if [[ ! -d "${chart_extracted}" ]]; then
          info "Extracting chart: ${chart_path}"
          mkdir -p "${chart_extracted}"
          tar -xzf "${chart_path}" -C "${chart_dir}" || \
            error "Failed to extract chart"
          chart_extracted=$(find "${chart_dir}" -type d -name "lvm-localpv-*" | head -1)
        fi
        
          # Check if Chart.yaml exists in extracted directory
          # Helm charts typically extract to: chart-name/chart-name/Chart.yaml
          local actual_chart_dir="${chart_extracted}"
          if [[ ! -f "${chart_extracted}/Chart.yaml" ]]; then
            # Try chart-name/chart-name structure first
            if [[ -d "${chart_extracted}/lvm-localpv" ]] && [[ -f "${chart_extracted}/lvm-localpv/Chart.yaml" ]]; then
              actual_chart_dir="${chart_extracted}/lvm-localpv"
            else
              # Find Chart.yaml but prefer the one closest to extracted directory (not in crds or charts subdirs)
              actual_chart_dir=$(find "${chart_extracted}" -maxdepth 2 -name "Chart.yaml" -type f | \
                grep -v "/charts/" | grep -v "/crds/" | head -1 | xargs dirname 2>/dev/null || echo "")
            fi
          fi
          
          if [[ -n "${actual_chart_dir}" ]] && [[ -f "${actual_chart_dir}/Chart.yaml" ]]; then
            OPENEBS_CHART_DIR="${actual_chart_dir}"
            info "Using extracted chart directory: ${OPENEBS_CHART_DIR}"
          else
            warn "Could not find Chart.yaml in extracted directory: ${chart_extracted}"
          fi
      # Check if it's a directory containing tgz file
      elif [[ -d "${chart_path}" ]]; then
        local chart_file
        chart_file=$(find "${chart_path}" -name "lvm-localpv-*.tgz" -type f | head -1)
        if [[ -n "${chart_file}" ]]; then
          local chart_name
          chart_name=$(basename "${chart_file}" .tgz)
          local chart_extracted="${chart_path}/${chart_name}"
          
          if [[ ! -d "${chart_extracted}" ]] || [[ ! -f "${chart_extracted}/Chart.yaml" ]]; then
            info "Extracting chart: ${chart_file}"
            mkdir -p "${chart_extracted}"
            tar -xzf "${chart_file}" -C "${chart_path}" || \
              error "Failed to extract chart"
            chart_extracted=$(find "${chart_path}" -type d -name "lvm-localpv-*" | head -1)
          fi
          
          # Check if Chart.yaml exists in extracted directory
          # Helm charts typically extract to: chart-name/chart-name/Chart.yaml
          local actual_chart_dir="${chart_extracted}"
          if [[ ! -f "${chart_extracted}/Chart.yaml" ]]; then
            # Try chart-name/chart-name structure first
            if [[ -d "${chart_extracted}/lvm-localpv" ]] && [[ -f "${chart_extracted}/lvm-localpv/Chart.yaml" ]]; then
              actual_chart_dir="${chart_extracted}/lvm-localpv"
            else
              # Find Chart.yaml but prefer the one closest to extracted directory (not in crds or charts subdirs)
              actual_chart_dir=$(find "${chart_extracted}" -maxdepth 2 -name "Chart.yaml" -type f | \
                grep -v "/charts/" | grep -v "/crds/" | head -1 | xargs dirname 2>/dev/null || echo "")
            fi
          fi
          
          if [[ -n "${actual_chart_dir}" ]] && [[ -f "${actual_chart_dir}/Chart.yaml" ]]; then
            OPENEBS_CHART_DIR="${actual_chart_dir}"
            info "Using extracted chart directory: ${OPENEBS_CHART_DIR}"
          else
            warn "Could not find Chart.yaml in extracted directory: ${chart_extracted}"
          fi
        fi
      fi
    fi
    
    # Auto-detect images directory
    if [[ -d "${media_dir}/images" ]] && [[ -z "${OPENEBS_LOAD_IMAGES:-}" ]]; then
      local image_count
      image_count=$(find "${media_dir}/images" -name "*.tar" -type f | wc -l)
      if [[ ${image_count} -gt 0 ]]; then
        info "Found ${image_count} image file(s) in ${media_dir}/images"
        info "To load images automatically, set OPENEBS_LOAD_IMAGES=true"
        info "Or run manually: ./deploy/scripts/load-images.sh --images-dir ${media_dir}/images"
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
  local load_script="${SCRIPT_DIR}/load-images.sh"
  if [[ -f "${load_script}" ]]; then
    info "Using load-images.sh script to load images"
    if [[ "${DRY_RUN}" == "true" ]]; then
      debug "Would run: ${load_script} --images-dir ${images_dir}"
    else
      if bash "${load_script}" --images-dir "${images_dir}" 2>&1 | tee -a "${INSTALL_LOG_PATH}"; then
        info "Images loaded successfully"
      else
        warn "Failed to load images. You may need to load them manually."
      fi
    fi
  else
    # Fallback: use docker/podman directly
    warn "load-images.sh not found, using direct docker/podman commands"
    local container_tool=""
    if installed docker; then
      container_tool="docker"
    elif installed podman; then
      container_tool="podman"
    else
      warn "No container tool found, skipping image loading"
      return 0
    fi
    
    info "Loading images using ${container_tool}..."
    local loaded=0
    for img_file in "${images_dir}"/*.tar; do
      if [[ -f "${img_file}" ]]; then
        if [[ "${DRY_RUN}" != "true" ]]; then
          ${container_tool} load -i "${img_file}" &>/dev/null || warn "Failed to load: ${img_file}"
        fi
        ((loaded++))
      fi
    done
    info "Loaded ${loaded} image(s)"
  fi
}

# Validate required environment variables
validate_environment() {
  local missing_vars=()
  
  [[ -z "${OPENEBS_STORAGECLASS_NAME:-}" ]] && missing_vars+=("OPENEBS_STORAGECLASS_NAME")
  [[ -z "${OPENEBS_VG_NAME:-}" ]] && missing_vars+=("OPENEBS_VG_NAME")
  [[ -z "${OPENEBS_CONTROLLER_NODE_NAMES:-}" ]] && missing_vars+=("OPENEBS_CONTROLLER_NODE_NAMES")
  [[ -z "${OPENEBS_DATA_NODE_NAMES:-}" ]] && missing_vars+=("OPENEBS_DATA_NODE_NAMES")
  
  if [[ ${#missing_vars[@]} -gt 0 ]]; then
    error "Missing required environment variables: ${missing_vars[*]}

Please set the required environment variables before running the script.
See --help for more information."
  fi
  
  if [[ "${OFFLINE_INSTALL}" == "true" ]]; then
    # Try to detect media directory first
    detect_offline_media
    
    if [[ -z "${OPENEBS_CHART_DIR:-}" ]]; then
      error "OPENEBS_CHART_DIR must be set for offline installation.
Or place charts in ./offline-media/charts/ directory."
    fi
  fi
}

# Label nodes
label_nodes() {
  local node_type="$1"
  local label_key="$2"
  local label_value="$3"
  local node_names="$4"
  
  IFS="," read -r -a node_array <<< "${node_names}"
  for node in "${node_array[@]}"; do
    node=$(echo "${node}" | xargs) # trim whitespace
    if [[ -z "${node}" ]]; then
      continue
    fi
    
    info "Labeling node: ${node} with ${label_key}=${label_value}"
    if [[ "${DRY_RUN}" == "true" ]]; then
      debug "Would run: kubectl label node ${node} '${label_key}=${label_value}' --overwrite"
    else
      if kubectl label node "${node}" "${label_key}=${label_value}" --overwrite &>/dev/null; then
        info "Successfully labeled node: ${node}"
      else
        error "Failed to label node: ${node}. Please check node name and permissions."
      fi
    fi
  done
}

# Common helm install arguments
get_common_helm_args() {
  local args=(
    "--namespace" "${OPENEBS_KUBE_NAMESPACE}"
    "--create-namespace"
    "--set" "lvmController.nodeSelector.openebs\.io/control-plane=enable"
    "--set-string" "lvmController.resources.limits.cpu=${OPENEBS_CONTROLLER_RESOURCE_LIMITS_CPU}"
    "--set-string" "lvmController.resources.limits.memory=${OPENEBS_CONTROLLER_RESOURCE_LIMITS_MEMORY}"
    "--set-string" "lvmController.resources.requests.cpu=${OPENEBS_CONTROLLER_RESOURCE_REQUESTS_CPU}"
    "--set-string" "lvmController.resources.requests.memory=${OPENEBS_CONTROLLER_RESOURCE_REQUESTS_MEMORY}"
    "--set" "lvmNode.nodeSelector.openebs\.io/node=enable"
    "--set-string" "lvmNode.resources.limits.cpu=${OPENEBS_NODE_RESOURCE_LIMITS_CPU}"
    "--set-string" "lvmNode.resources.limits.memory=${OPENEBS_NODE_RESOURCE_LIMITS_MEMORY}"
    "--set-string" "lvmNode.resources.requests.cpu=${OPENEBS_NODE_RESOURCE_REQUESTS_CPU}"
    "--set-string" "lvmNode.resources.requests.memory=${OPENEBS_NODE_RESOURCE_REQUESTS_MEMORY}"
    "--set" "lvmPlugin.allowedTopologies=kubernetes\.io/hostname\,openebs\.io/node"
    "--set" "analytics.enabled=false"
    "--timeout" "${TIME_OUT_SECOND}"
    "--wait"
  )
  
  echo "${args[@]}"
}

# Online installation
online_install_lvmlocalpv() {
  info "Starting online installation..."
  
  # Check if already installed
  if helm status "${RELEASE}" -n "${OPENEBS_KUBE_NAMESPACE}" &>/dev/null; then
    error "${RELEASE} is already installed in namespace ${OPENEBS_KUBE_NAMESPACE}. 
Please uninstall it first using: helm uninstall ${RELEASE} -n ${OPENEBS_KUBE_NAMESPACE}"
  fi
  
  # Add and update helm repo
  info "Adding Helm repository: openebs-lvmlocalpv"
  if [[ "${DRY_RUN}" == "true" ]]; then
    debug "Would run: helm repo add openebs-lvmlocalpv https://openebs.github.io/lvm-localpv"
    debug "Would run: helm repo update openebs-lvmlocalpv"
  else
    helm repo add openebs-lvmlocalpv https://openebs.github.io/lvm-localpv &>/dev/null || \
      error "Failed to add Helm repository. Check network connectivity."
    info "Updating Helm repository"
    helm repo update openebs-lvmlocalpv 2>/dev/null || \
      error "Failed to update Helm repository."
  fi
  
  # Determine chart version
  if [[ -z "${CHART_VERSION}" ]]; then
    CHART_VERSION=$(helm search repo openebs-lvmlocalpv/lvm-localpv --versions --output json 2>/dev/null | \
      grep -oP '"version":\s*"\K[^"]+' | head -1 || echo "")
    if [[ -z "${CHART_VERSION}" ]]; then
      warn "Could not determine latest chart version, proceeding without version pinning"
    else
      info "Using chart version: ${CHART_VERSION}"
    fi
  fi
  
  # Build helm install command
  local helm_args=($(get_common_helm_args))
  if [[ -n "${CHART_VERSION}" ]]; then
    helm_args+=("--version" "${CHART_VERSION}")
  fi
  
  info "Installing ${RELEASE} (this may take several minutes)..."
  if [[ "${DRY_RUN}" == "true" ]]; then
    debug "Would run: helm install ${RELEASE} ${CHART} ${helm_args[*]}"
    info "Dry-run mode: Skipping actual installation"
  else
    if helm install "${RELEASE}" "${CHART}" "${helm_args[@]}" 2>&1 | tee -a "${INSTALL_LOG_PATH}"; then
      ROLLBACK_RESOURCES+=("helm:${RELEASE}:${OPENEBS_KUBE_NAMESPACE}")
      info "Helm installation completed successfully"
    else
      error "Helm installation failed. Check logs for details."
    fi
  fi
}

# Offline installation
offline_install_lvmlocalpv() {
  info "Starting offline installation..."
  
  # Check if already installed
  if helm status "${RELEASE}" -n "${OPENEBS_KUBE_NAMESPACE}" &>/dev/null; then
    error "${RELEASE} is already installed in namespace ${OPENEBS_KUBE_NAMESPACE}. 
Please uninstall it first using: helm uninstall ${RELEASE} -n ${OPENEBS_KUBE_NAMESPACE}"
  fi
  
  # Validate and prepare chart directory
  local chart_dir="${OPENEBS_CHART_DIR}"
  
  # Resolve relative path
  if [[ ! "${chart_dir}" =~ ^/ ]]; then
    chart_dir="$(cd "$(dirname "${chart_dir}")" 2>/dev/null && pwd)/$(basename "${chart_dir}")" || \
      chart_dir="$(realpath "${OPENEBS_CHART_DIR}" 2>/dev/null || echo "${OPENEBS_CHART_DIR}")"
  fi
  
  # Check if it's a directory containing tgz file (but not Chart.yaml)
  if [[ -d "${chart_dir}" ]] && [[ ! -f "${chart_dir}/Chart.yaml" ]]; then
    # Look for tgz file in this directory
    local chart_file
    chart_file=$(find "${chart_dir}" -maxdepth 1 -name "lvm-localpv-*.tgz" -type f | head -1)
    if [[ -n "${chart_file}" ]]; then
      local chart_name
      chart_name=$(basename "${chart_file}" .tgz)
      local chart_extracted="${chart_dir}/${chart_name}"
      
      if [[ ! -d "${chart_extracted}" ]]; then
        info "Extracting chart: ${chart_file}"
        tar -xzf "${chart_file}" -C "${chart_dir}" || \
          error "Failed to extract chart"
      fi
      
      # Check what was actually extracted
      # Helm charts can have different structures
      local actual_chart_dir=""
      
      # First, check if Chart.yaml is directly in the extracted directory
      if [[ -f "${chart_extracted}/Chart.yaml" ]]; then
        actual_chart_dir="${chart_extracted}"
      # Check if there's a lvm-localpv subdirectory (without version)
      elif [[ -d "${chart_extracted}/lvm-localpv" ]] && [[ -f "${chart_extracted}/lvm-localpv/Chart.yaml" ]]; then
        actual_chart_dir="${chart_extracted}/lvm-localpv"
      # Check if there's a lvm-localpv directory at the same level
      elif [[ -d "${chart_dir}/lvm-localpv" ]] && [[ -f "${chart_dir}/lvm-localpv/Chart.yaml" ]]; then
        actual_chart_dir="${chart_dir}/lvm-localpv"
      # Search for Chart.yaml but exclude crds and charts subdirs
      else
        actual_chart_dir=$(find "${chart_dir}" -maxdepth 3 -name "Chart.yaml" -type f | \
          grep -v "/charts/" | grep -v "/crds/" | grep -v "crds" | head -1 | xargs dirname 2>/dev/null || echo "")
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
        debug "Extracted directory contents:"
        ls -la "${chart_extracted}" 2>/dev/null | head -10 | while read -r line; do
          debug "  ${line}"
        done
        error "Failed to extract or find valid chart directory with Chart.yaml in: ${chart_dir}"
      fi
    else
      error "Chart directory does not contain Chart.yaml or chart tgz file: ${OPENEBS_CHART_DIR}"
    fi
  elif [[ ! -d "${chart_dir}" ]]; then
    error "Chart directory does not exist: ${OPENEBS_CHART_DIR}"
  fi
  
  # Final validation - check for Chart.yaml in directory
  if [[ ! -f "${chart_dir}/Chart.yaml" ]]; then
    # Look for Chart.yaml in subdirectories, but exclude crds and charts subdirs
    local chart_yaml_path
    chart_yaml_path=$(find "${chart_dir}" -maxdepth 2 -name "Chart.yaml" -type f | \
      grep -v "/charts/" | grep -v "/crds/" | head -1)
    if [[ -n "${chart_yaml_path}" ]]; then
      chart_dir=$(dirname "${chart_yaml_path}")
      OPENEBS_CHART_DIR="${chart_dir}"
      info "Found Chart.yaml in subdirectory, using: ${OPENEBS_CHART_DIR}"
    else
      # Debug: list directory contents
      debug "Chart directory contents:"
      ls -la "${chart_dir}" 2>/dev/null | head -10 | while read -r line; do
        debug "  ${line}"
      done
      find "${chart_dir}" -name "Chart.yaml" -type f 2>/dev/null | while read -r f; do
        debug "  Found Chart.yaml at: ${f}"
      done
      error "Invalid chart directory (missing Chart.yaml): ${chart_dir}"
    fi
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
  
  # Build helm install command
  local helm_args=($(get_common_helm_args))
  helm_args+=(
    "--set-string" "lvmController.resizer.image.registry=${image_registry}"
    "--set-string" "lvmController.snapshotter.image.registry=${image_registry}"
    "--set-string" "lvmController.snapshotController.image.registry=${image_registry}"
    "--set-string" "lvmController.provisioner.image.registry=${image_registry}"
    "--set-string" "lvmNode.driverRegistrar.image.registry=${image_registry}"
    "--set-string" "lvmPlugin.image.registry=${plugin_image_registry}"
  )
  
  info "Installing ${RELEASE} from local chart: ${OPENEBS_CHART_DIR}"
  if [[ "${DRY_RUN}" == "true" ]]; then
    debug "Would run: helm install ${RELEASE} ${OPENEBS_CHART_DIR} ${helm_args[*]}"
    info "Dry-run mode: Skipping actual installation"
  else
    if helm install "${RELEASE}" "${OPENEBS_CHART_DIR}" "${helm_args[@]}" 2>&1 | tee -a "${INSTALL_LOG_PATH}"; then
      ROLLBACK_RESOURCES+=("helm:${RELEASE}:${OPENEBS_KUBE_NAMESPACE}")
      info "Helm installation completed successfully"
    else
      error "Helm installation failed. Check logs for details."
    fi
  fi
}

# Verify installation
verify_installed() {
  info "Verifying installation..."
  
  if [[ "${DRY_RUN}" == "true" ]]; then
    info "Dry-run mode: Skipping verification"
    return 0
  fi
  
  # Check Helm release status
  local status
  status=$(helm status "${RELEASE}" -n "${OPENEBS_KUBE_NAMESPACE}" 2>/dev/null | grep ^STATUS: | awk '{print $2}' || echo "")
  if [[ "${status}" != "deployed" ]]; then
    error "Helm release status is '${status}', expected 'deployed'. Check: helm status ${RELEASE} -n ${OPENEBS_KUBE_NAMESPACE}"
  fi
  info "Helm release status: ${status}"
  
  # Check controller pods
  info "Checking controller pods..."
  local controller_ready
  controller_ready=$(kubectl get pods -n "${OPENEBS_KUBE_NAMESPACE}" -l app=openebs-lvm-controller --no-headers 2>/dev/null | \
    awk '{if ($2 ~ /^[0-9]+\/[0-9]+$/ && $3 == "Running") print $1}' | wc -l || echo "0")
  if [[ "${controller_ready}" -eq 0 ]]; then
    warn "No controller pods are ready yet. This may be normal during initial installation."
  else
    info "Controller pods ready: ${controller_ready}"
  fi
  
  # Check node daemonset
  info "Checking node daemonset..."
  local node_ready node_desired
  node_ready=$(kubectl get daemonset -n "${OPENEBS_KUBE_NAMESPACE}" -l app=openebs-lvm-node --no-headers 2>/dev/null | \
    awk '{print $4}' | cut -d'/' -f1 || echo "0")
  node_desired=$(kubectl get daemonset -n "${OPENEBS_KUBE_NAMESPACE}" -l app=openebs-lvm-node --no-headers 2>/dev/null | \
    awk '{print $4}' | cut -d'/' -f2 || echo "0")
  if [[ "${node_desired}" -gt 0 ]]; then
    info "Node daemonset: ${node_ready}/${node_desired} pods ready"
  else
    warn "No node daemonset found. This may be normal if no nodes are labeled."
  fi
  
  # Check CRDs
  info "Checking CRDs..."
  local crds=("lvmvolumes.local.openebs.io" "lvmnodes.local.openebs.io" "lvmsnapshots.local.openebs.io")
  for crd in "${crds[@]}"; do
    if kubectl get crd "${crd}" &>/dev/null; then
      info "CRD ${crd} exists"
    else
      warn "CRD ${crd} not found"
    fi
  done
  
  # Check LVMNode resources
  info "Checking LVMNode resources..."
  local lvmnode_count
  lvmnode_count=$(kubectl get lvmnode -n "${OPENEBS_KUBE_NAMESPACE}" --no-headers 2>/dev/null | wc -l || echo "0")
  if [[ "${lvmnode_count}" -gt 0 ]]; then
    info "Found ${lvmnode_count} LVMNode resource(s)"
  else
    warn "No LVMNode resources found. This may be normal if nodes haven't been discovered yet."
  fi
  
  info "Installation verification completed"
}

# Create StorageClass from embedded template
create_storageclass() {
  info "Creating StorageClass: ${OPENEBS_STORAGECLASS_NAME}"
  
  if [[ "${DRY_RUN}" == "true" ]]; then
    info "Dry-run mode: Would create StorageClass ${OPENEBS_STORAGECLASS_NAME}"
    return 0
  fi
  
  # Embedded StorageClass template
  local storageclass_yaml
  storageclass_yaml=$(cat <<EOF
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: ${OPENEBS_STORAGECLASS_NAME}
allowVolumeExpansion: true
parameters:
  volgroup: "${OPENEBS_VG_NAME}"
provisioner: local.csi.openebs.io
EOF
)
  
  # Use provided YAML if available
  if [[ -n "${OPENEBS_STORAGECLASS_YAML}" ]] && [[ -f "${OPENEBS_STORAGECLASS_YAML}" ]]; then
    info "Using StorageClass template from: ${OPENEBS_STORAGECLASS_YAML}"
    OPENEBS_STORAGECLASS_NAME="${OPENEBS_STORAGECLASS_NAME}" \
      OPENEBS_VG_NAME="${OPENEBS_VG_NAME}" \
      envsubst < "${OPENEBS_STORAGECLASS_YAML}" | kubectl apply -f - || \
      error "Failed to create StorageClass from template"
  else
    echo "${storageclass_yaml}" | kubectl apply -f - || \
      error "Failed to create StorageClass"
  fi
  
  # Verify StorageClass was created
  if kubectl get storageclass "${OPENEBS_STORAGECLASS_NAME}" &>/dev/null; then
    info "StorageClass ${OPENEBS_STORAGECLASS_NAME} created successfully"
    ROLLBACK_RESOURCES+=("storageclass:${OPENEBS_STORAGECLASS_NAME}")
  else
    error "StorageClass ${OPENEBS_STORAGECLASS_NAME} was not created"
  fi
}

# Rollback on failure
rollback() {
  if [[ ${#ROLLBACK_RESOURCES[@]} -eq 0 ]]; then
    return 0
  fi
  
  warn "Rolling back installed resources..."
  for resource in "${ROLLBACK_RESOURCES[@]}"; do
    local type name namespace
    IFS=':' read -r type name namespace <<< "${resource}"
    case "${type}" in
      helm)
        warn "Uninstalling Helm release: ${name}"
        helm uninstall "${name}" -n "${namespace}" --ignore-not-found &>/dev/null || true
        ;;
      storageclass)
        warn "Deleting StorageClass: ${name}"
        kubectl delete storageclass "${name}" --ignore-not-found &>/dev/null || true
        ;;
    esac
  done
  warn "Rollback completed"
}

# Main installation function
main() {
  # Re-read environment variables (in case they were set after script started)
  # This allows users to set env vars just before running the script
  OFFLINE_INSTALL="${OFFLINE_INSTALL:-false}"
  DRY_RUN="${DRY_RUN:-false}"
  OPENEBS_KUBE_NAMESPACE="${OPENEBS_KUBE_NAMESPACE:-${DEFAULT_NAMESPACE}}"
  OPENEBS_CREATE_STORAGECLASS="${OPENEBS_CREATE_STORAGECLASS:-false}"
  OPENEBS_STORAGECLASS_YAML="${OPENEBS_STORAGECLASS_YAML:-}"
  
  # Parse arguments first (this may override env vars)
  parse_args "$@"
  
  # Initialize logging
  init_log
  
  # Show version info
  info "OpenEBS LVM LocalPV Installation Script v${SCRIPT_VERSION}"
  
  if [[ "${DRY_RUN}" == "true" ]]; then
    info "DRY-RUN MODE: No changes will be made"
  fi
  
  # Check prerequisites
  check_prerequisites
  
  # Validate environment
  validate_environment
  
  # Set chart version if not specified
  if [[ -z "${CHART_VERSION}" ]]; then
    CHART_VERSION=$(get_chart_version)
    if [[ -n "${CHART_VERSION}" ]]; then
      info "Auto-detected chart version: ${CHART_VERSION}"
    fi
  fi
  
  # Label nodes
  info "Labeling controller nodes..."
  label_nodes "controller" "openebs.io/control-plane" "enable" "${OPENEBS_CONTROLLER_NODE_NAMES}"
  
  info "Labeling data nodes..."
  label_nodes "data" "openebs.io/node" "enable" "${OPENEBS_DATA_NODE_NAMES}"
  
  # Load images if in offline mode and requested
  if [[ "${OFFLINE_INSTALL}" == "true" ]]; then
    load_offline_images
  fi
  
  # Install based on mode
  if [[ "${OFFLINE_INSTALL}" == "true" ]]; then
    offline_install_lvmlocalpv
  else
    online_install_lvmlocalpv
  fi
  
  # Verify installation
  verify_installed
  
  # Create StorageClass if requested
  debug "OPENEBS_CREATE_STORAGECLASS=${OPENEBS_CREATE_STORAGECLASS}"
  debug "OPENEBS_STORAGECLASS_NAME=${OPENEBS_STORAGECLASS_NAME:-}"
  debug "OPENEBS_VG_NAME=${OPENEBS_VG_NAME:-}"
  
  if [[ "${OPENEBS_CREATE_STORAGECLASS}" == "true" ]] || [[ "${OPENEBS_CREATE_STORAGECLASS}" == "1" ]]; then
    if [[ -z "${OPENEBS_STORAGECLASS_NAME:-}" ]]; then
      error "OPENEBS_STORAGECLASS_NAME must be set when OPENEBS_CREATE_STORAGECLASS=true"
    fi
    if [[ -z "${OPENEBS_VG_NAME:-}" ]]; then
      error "OPENEBS_VG_NAME must be set when OPENEBS_CREATE_STORAGECLASS=true"
    fi
    create_storageclass
  else
    info "StorageClass creation skipped (set OPENEBS_CREATE_STORAGECLASS=true to enable)"
  fi
  
  info "Installation completed successfully!"
  info "Next steps:"
  info "  1. Verify pods are running: kubectl get pods -n ${OPENEBS_KUBE_NAMESPACE}"
  info "  2. Check LVMNode resources: kubectl get lvmnode -n ${OPENEBS_KUBE_NAMESPACE}"
  if [[ "${OPENEBS_CREATE_STORAGECLASS}" != "true" ]]; then
    info "  3. Create a StorageClass with volgroup: ${OPENEBS_VG_NAME}"
  fi
}

# Trap errors and rollback
trap 'if [[ $? -ne 0 ]]; then rollback; fi' EXIT

# Run main function
main "$@"
