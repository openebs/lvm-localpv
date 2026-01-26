#!/usr/bin/env bash

###############################################################################
# OpenEBS LVM LocalPV Offline Media Download Script
#
# This script downloads all required media (Helm charts and Docker images)
# for offline installation of OpenEBS LVM LocalPV.
#
# Usage:
#   ./download.sh [OPTIONS]
#
# Options:
#   --help                  Show this help message
#   --version              Show script version
#   --chart-version VER    Chart version to download (default: latest)
#   --output-dir DIR       Output directory (default: ./offline-media)
#   --pack                 Pack all media into tar.gz after download
#   --log-level LEVEL      Set log level (DEBUG, INFO, WARN, ERROR)
#
###############################################################################

set -euo pipefail

# Script metadata
readonly SCRIPT_NAME="download.sh"
readonly SCRIPT_VERSION="1.0.0"
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly ROOT_DIR="$(cd "${SCRIPT_DIR}/../.." && pwd)"
readonly CHART_DIR="${ROOT_DIR}/deploy/helm/charts"
readonly CHART_YAML="${CHART_DIR}/Chart.yaml"
readonly VALUES_YAML="${CHART_DIR}/values.yaml"

# Default values
readonly DEFAULT_REPO="https://openebs.github.io/lvm-localpv"
readonly DEFAULT_CHART="openebs-lvmlocalpv/lvm-localpv"
readonly DEFAULT_OUTPUT_DIR="./offline-media"
readonly DEFAULT_LOG_LEVEL="INFO"

# Global variables
CHART_VERSION="${CHART_VERSION:-}"
OUTPUT_DIR="${OUTPUT_DIR:-${DEFAULT_OUTPUT_DIR}}"
PACK_MEDIA="${PACK_MEDIA:-false}"
LOG_LEVEL="${LOG_LEVEL:-${DEFAULT_LOG_LEVEL}}"
CONTAINER_TOOL=""

# Log file path
if [[ -w /tmp ]]; then
  DOWNLOAD_LOG_PATH="/tmp/openebs-lvmlocalpv_download-$(date +'%Y-%m-%d_%H-%M-%S').log"
else
  DOWNLOAD_LOG_PATH="./openebs-lvmlocalpv_download-$(date +'%Y-%m-%d_%H-%M-%S').log"
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
  
  # Output to stderr to avoid polluting stdout (important for command substitution)
  echo "[${level}][${timestamp}]: ${message}" | tee -a "${DOWNLOAD_LOG_PATH}" >&2
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
  local missing_tools=()
  
  if ! installed helm; then
    missing_tools+=("helm")
    warn "helm is not installed. Installation instructions:"
    warn "  Linux: curl https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash"
    warn "  macOS: brew install helm"
    warn "  Or visit: https://helm.sh/docs/intro/install/"
  fi
  
  # Check for container tool (docker or podman)
  if installed docker; then
    CONTAINER_TOOL="docker"
  elif installed podman; then
    CONTAINER_TOOL="podman"
  else
    missing_tools+=("docker or podman")
    warn "Neither docker nor podman is installed. Installation instructions:"
    warn "  docker: https://docs.docker.com/get-docker/"
    warn "  podman: https://podman.io/getting-started/installation"
  fi
  
  if [[ ${#missing_tools[@]} -gt 0 ]]; then
    error "Missing required tools: ${missing_tools[*]}

Please install the missing tools before running this script."
  fi
  
  info "Using container tool: ${CONTAINER_TOOL}"
}

init_log() {
  touch "${DOWNLOAD_LOG_PATH}" || error "Cannot create log file: ${DOWNLOAD_LOG_PATH}"
  info "Log file: ${DOWNLOAD_LOG_PATH}"
  debug "Script version: ${SCRIPT_VERSION}"
}

show_help() {
  cat <<EOF
OpenEBS LVM LocalPV Offline Media Download Script v${SCRIPT_VERSION}

USAGE:
    ${SCRIPT_NAME} [OPTIONS]

OPTIONS:
    --help                  Show this help message and exit
    --version               Show script version and exit
    --chart-version VER     Chart version to download (default: latest)
    --output-dir DIR        Output directory (default: ${DEFAULT_OUTPUT_DIR})
    --pack                  Pack all media into tar.gz after download
    --log-level LEVEL       Set log level: DEBUG, INFO, WARN, ERROR (default: ${DEFAULT_LOG_LEVEL})

ENVIRONMENT VARIABLES:
    CHART_VERSION           Chart version to download
    OUTPUT_DIR              Output directory for downloaded media
    PACK_MEDIA              Set to "true" to pack media after download
    LOG_LEVEL               Log level (DEBUG, INFO, WARN, ERROR)

EXAMPLES:
    # Download latest version
    ./download.sh

    # Download specific version
    ./download.sh --chart-version 1.8.0

    # Download and pack
    ./download.sh --chart-version 1.8.0 --pack

    # Custom output directory
    ./download.sh --output-dir /path/to/media --pack

OUTPUT STRUCTURE:
    offline-media/
    ├── charts/
    │   └── lvm-localpv-<version>.tgz
    ├── images/
    │   ├── *.tar (Docker images)
    │   └── images.list
    ├── manifests/
    │   └── VERSION
    └── README.md

For more information, visit: https://github.com/openebs/lvm-localpv
EOF
}

show_version() {
  echo "${SCRIPT_NAME} v${SCRIPT_VERSION}"
}

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
      --chart-version)
        CHART_VERSION="$2"
        shift 2
        ;;
      --output-dir)
        OUTPUT_DIR="$2"
        shift 2
        ;;
      --pack)
        PACK_MEDIA=true
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

# Get chart version from Chart.yaml or Helm repo
get_chart_version() {
  if [[ -n "${CHART_VERSION}" ]]; then
    echo "${CHART_VERSION}"
    return 0
  fi
  
  # Try to get from Chart.yaml using awk
  if [[ -f "${CHART_YAML}" ]]; then
    local version
    version=$(awk '/^version:/ {gsub(/^version:[[:space:]]*/, ""); gsub(/[\"'\''"]/, ""); gsub(/[[:space:]]*$/, ""); print; exit}' "${CHART_YAML}" 2>/dev/null || echo "")
    if [[ -n "${version}" ]] && [[ "${version}" != *"develop"* ]]; then
      echo "${version}"
      return 0
    fi
  fi
  
  # Get latest from Helm repo
  info "Fetching latest chart version from Helm repository..."
  helm repo add openebs-lvmlocalpv "${DEFAULT_REPO}" &>/dev/null || true
  helm repo update openebs-lvmlocalpv &>/dev/null || true
  
  local latest_version
  latest_version=$(helm search repo openebs-lvmlocalpv/lvm-localpv --versions --output json 2>/dev/null | \
    grep -oP '"version":\s*"\K[^"]+' | grep -v "develop" | head -1 || echo "")
  
  if [[ -n "${latest_version}" ]]; then
    echo "${latest_version}"
  else
    error "Could not determine chart version. Please specify with --chart-version"
    return 1
  fi
}

# Create media directory structure
create_media_structure() {
  info "Creating media directory structure..."
  
  mkdir -p "${OUTPUT_DIR}/charts"
  mkdir -p "${OUTPUT_DIR}/images"
  mkdir -p "${OUTPUT_DIR}/manifests"
  
  info "Media directory: ${OUTPUT_DIR}"
}

# List available chart versions
list_available_versions() {
  info "Fetching available chart versions..."
  helm repo add openebs-lvmlocalpv "${DEFAULT_REPO}" &>/dev/null || true
  helm repo update openebs-lvmlocalpv &>/dev/null || true
  
  local versions
  versions=$(helm search repo openebs-lvmlocalpv/lvm-localpv --versions --output json 2>/dev/null | \
    grep -oP '"version":\s*"\K[^"]+' | grep -v "develop" | head -10 || echo "")
  
  if [[ -n "${versions}" ]]; then
    warn "Available chart versions (showing first 10):"
    echo "${versions}" | while read -r v; do
      warn "  - ${v}"
    done
  else
    warn "Could not fetch available versions. The repository might be unreachable."
  fi
}

# Validate chart version exists
validate_chart_version() {
  local version="$1"
  
  # Add and update repo
  helm repo add openebs-lvmlocalpv "${DEFAULT_REPO}" &>/dev/null || true
  helm repo update openebs-lvmlocalpv &>/dev/null || true
  
  # Get all available versions and check if the requested version exists
  local all_versions
  all_versions=$(helm search repo openebs-lvmlocalpv/lvm-localpv --versions --output json 2>/dev/null | \
    grep -oP '"version":\s*"\K[^"]+' | grep -v "develop" || echo "")
  
  # Check if version exists in the list
  local version_exists=0
  if echo "${all_versions}" | grep -qFx "${version}"; then
    version_exists=1
  fi
  
  if [[ "${version_exists}" -eq 0 ]]; then
    # Try to find similar versions (same major version)
    local major_version="${version%%.*}"
    local similar_versions
    similar_versions=$(echo "${all_versions}" | grep "^${major_version}\." | head -5 || echo "")
    
    warn "Chart version ${version} not found in repository."
    
    if [[ -n "${similar_versions}" ]]; then
      warn "Similar versions found:"
      echo "${similar_versions}" | while read -r v; do
        [[ -n "${v}" ]] && warn "  - ${v}"
      done
    else
      list_available_versions
    fi
    
    return 1
  fi
  
  return 0
}

# Download Helm chart
download_chart() {
  local version="$1"
  local chart_file="${OUTPUT_DIR}/charts/lvm-localpv-${version}.tgz"
  
  info "Downloading Helm chart version ${version}..."
  
  # Add repo if not exists
  if ! helm repo list | grep -q "openebs-lvmlocalpv"; then
    info "Adding Helm repository: ${DEFAULT_REPO}"
    helm repo add openebs-lvmlocalpv "${DEFAULT_REPO}" || \
      error "Failed to add Helm repository"
  fi
  
  # Update repo
  info "Updating Helm repository..."
  helm repo update openebs-lvmlocalpv || \
    error "Failed to update Helm repository"
  
  # Download chart (version already validated in main flow)
  info "Downloading chart to: ${chart_file}"
  local pull_output
  pull_output=$(helm pull "${DEFAULT_CHART}" --version "${version}" --destination "${OUTPUT_DIR}/charts" 2>&1)
  local pull_exit_code=$?
  
  echo "${pull_output}" | tee -a "${DOWNLOAD_LOG_PATH}"
  
  if [[ ${pull_exit_code} -eq 0 ]]; then
    # Find the downloaded file (may have different naming)
    local downloaded_file
    downloaded_file=$(find "${OUTPUT_DIR}/charts" -name "lvm-localpv-*.tgz" -type f | head -1)
    
    if [[ -n "${downloaded_file}" ]]; then
      # Rename to expected name if different
      if [[ "${downloaded_file}" != "${chart_file}" ]]; then
        mv "${downloaded_file}" "${chart_file}"
        info "Renamed chart file to: ${chart_file}"
      fi
      info "Chart downloaded successfully: ${chart_file}"
    else
      error "Chart downloaded but file not found in expected location"
    fi
  else
    # Try to provide helpful error message
    if echo "${pull_output}" | grep -q "not found"; then
      error "Chart version ${version} not found.

$(list_available_versions)

Please specify a valid version using --chart-version."
    else
      error "Failed to download chart. Error: ${pull_output}"
    fi
  fi
  
  # Verify chart
  if [[ -f "${chart_file}" ]]; then
    local chart_size
    chart_size=$(du -h "${chart_file}" | cut -f1)
    info "Chart size: ${chart_size}"
    
    # Verify chart integrity
    if tar -tzf "${chart_file}" &>/dev/null; then
      info "Chart integrity verified"
    else
      error "Chart file appears to be corrupted: ${chart_file}"
    fi
  else
    error "Chart file not found: ${chart_file}"
  fi
}

# Extract image information from values.yaml
extract_images_from_chart() {
  local chart_file="$1"
  local temp_dir
  temp_dir=$(mktemp -d)
  
  # Extract chart
  tar -xzf "${chart_file}" -C "${temp_dir}" || error "Failed to extract chart"
  
  # Find values.yaml - it could be in different locations
  local values_file
  values_file=$(find "${temp_dir}" -name "values.yaml" -type f | head -1)
  
  if [[ -z "${values_file}" ]] || [[ ! -f "${values_file}" ]]; then
    # Debug: list extracted files
    debug "Chart extraction directory contents:"
    find "${temp_dir}" -type f | head -10 | while read -r f; do
      debug "  ${f}"
    done
    error "values.yaml not found in chart. Chart structure may be unexpected."
  fi
  
  debug "Found values.yaml at: ${values_file}"
  
  # Parse images using awk
  local images=()
  
  # Extract images using awk - track YAML indentation levels
    local temp_awk
    temp_awk=$(mktemp)
    cat > "${temp_awk}" << 'AWKEOF'
BEGIN {
  in_lvmNode = 0
  in_lvmController = 0
  in_lvmPlugin = 0
  in_driverRegistrar = 0
  in_resizer = 0
  in_snapshotter = 0
  in_snapshotController = 0
  in_provisioner = 0
  in_image = 0
  dr_reg = ""; dr_repo = ""; dr_tag = ""
  res_reg = ""; res_repo = ""; res_tag = ""
  snap_reg = ""; snap_repo = ""; snap_tag = ""
  snapc_reg = ""; snapc_repo = ""; snapc_tag = ""
  prov_reg = ""; prov_repo = ""; prov_tag = ""
  plugin_reg = ""; plugin_repo = ""; plugin_tag = ""
}

# Track section boundaries
/^[[:space:]]*lvmNode:/ { 
  in_lvmNode = 1
  in_driverRegistrar = 0
  in_image = 0
  dr_reg = ""; dr_repo = ""; dr_tag = ""
  next 
}

/^[[:space:]]*lvmController:/ { 
  in_lvmController = 1
  in_resizer = 0
  in_snapshotter = 0
  in_snapshotController = 0
  in_provisioner = 0
  in_image = 0
  res_reg = ""; res_repo = ""; res_tag = ""
  snap_reg = ""; snap_repo = ""; snap_tag = ""
  snapc_reg = ""; snapc_repo = ""; snapc_tag = ""
  prov_reg = ""; prov_repo = ""; prov_tag = ""
  next 
}

/^[[:space:]]*lvmPlugin:/ { 
  in_lvmPlugin = 1
  in_image = 0
  plugin_reg = ""; plugin_repo = ""; plugin_tag = ""
  next 
}

# Driver Registrar section
/^[[:space:]]*driverRegistrar:/ && in_lvmNode { 
  in_driverRegistrar = 1
  in_image = 0
  dr_reg = ""; dr_repo = ""; dr_tag = ""
  next 
}

/^[[:space:]]*image:/ && in_driverRegistrar { 
  in_image = 1
  next 
}

/^[[:space:]]*registry:/ && in_image && in_driverRegistrar && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*registry:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  dr_reg = $0
  next 
}

/^[[:space:]]*repository:/ && in_image && in_driverRegistrar && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*repository:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  dr_repo = $0
  next 
}

/^[[:space:]]*tag:/ && in_image && in_driverRegistrar && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*tag:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  dr_tag = $0
  if (dr_repo != "" && dr_tag != "") {
    if (dr_reg == "") {
      dr_reg = "docker.io/"
    }
    gsub(/\/$/, "", dr_reg)
    print dr_reg "/" dr_repo ":" dr_tag
  }
  in_image = 0
  next
}

# Resizer section
/^[[:space:]]*resizer:/ && in_lvmController { 
  in_resizer = 1
  in_image = 0
  res_reg = ""; res_repo = ""; res_tag = ""
  next 
}

/^[[:space:]]*image:/ && in_resizer { 
  in_image = 1
  next 
}

/^[[:space:]]*registry:/ && in_image && in_resizer && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*registry:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  res_reg = $0
  next 
}

/^[[:space:]]*repository:/ && in_image && in_resizer && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*repository:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  res_repo = $0
  next 
}

/^[[:space:]]*tag:/ && in_image && in_resizer && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*tag:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  res_tag = $0
  if (res_repo != "" && res_tag != "") {
    if (res_reg == "") {
      res_reg = "docker.io/"
    }
    gsub(/\/$/, "", res_reg)
    print res_reg "/" res_repo ":" res_tag
  }
  in_image = 0
  next
}

# Snapshotter section
/^[[:space:]]*snapshotter:/ && in_lvmController { 
  in_snapshotter = 1
  in_image = 0
  snap_reg = ""; snap_repo = ""; snap_tag = ""
  next 
}

/^[[:space:]]*image:/ && in_snapshotter { 
  in_image = 1
  next 
}

/^[[:space:]]*registry:/ && in_image && in_snapshotter && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*registry:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  snap_reg = $0
  next 
}

/^[[:space:]]*repository:/ && in_image && in_snapshotter && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*repository:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  snap_repo = $0
  next 
}

/^[[:space:]]*tag:/ && in_image && in_snapshotter && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*tag:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  snap_tag = $0
  if (snap_repo != "" && snap_tag != "") {
    if (snap_reg == "") {
      snap_reg = "docker.io/"
    }
    gsub(/\/$/, "", snap_reg)
    print snap_reg "/" snap_repo ":" snap_tag
  }
  in_image = 0
  next
}

# Snapshot Controller section
/^[[:space:]]*snapshotController:/ && in_lvmController { 
  in_snapshotController = 1
  in_image = 0
  snapc_reg = ""; snapc_repo = ""; snapc_tag = ""
  next 
}

/^[[:space:]]*image:/ && in_snapshotController { 
  in_image = 1
  next 
}

/^[[:space:]]*registry:/ && in_image && in_snapshotController && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*registry:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  snapc_reg = $0
  next 
}

/^[[:space:]]*repository:/ && in_image && in_snapshotController && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*repository:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  snapc_repo = $0
  next 
}

/^[[:space:]]*tag:/ && in_image && in_snapshotController && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*tag:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  snapc_tag = $0
  if (snapc_repo != "" && snapc_tag != "") {
    if (snapc_reg == "") {
      snapc_reg = "docker.io/"
    }
    gsub(/\/$/, "", snapc_reg)
    print snapc_reg "/" snapc_repo ":" snapc_tag
  }
  in_image = 0
  next
}

# Provisioner section
/^[[:space:]]*provisioner:/ && in_lvmController { 
  in_provisioner = 1
  in_image = 0
  prov_reg = ""; prov_repo = ""; prov_tag = ""
  next 
}

/^[[:space:]]*image:/ && in_provisioner { 
  in_image = 1
  next 
}

/^[[:space:]]*registry:/ && in_image && in_provisioner && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*registry:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  prov_reg = $0
  next 
}

/^[[:space:]]*repository:/ && in_image && in_provisioner && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*repository:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  prov_repo = $0
  next 
}

/^[[:space:]]*tag:/ && in_image && in_provisioner && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*tag:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  prov_tag = $0
  if (prov_repo != "" && prov_tag != "") {
    if (prov_reg == "") {
      prov_reg = "docker.io/"
    }
    gsub(/\/$/, "", prov_reg)
    print prov_reg "/" prov_repo ":" prov_tag
  }
  in_image = 0
  next
}

# LVM Plugin section
/^[[:space:]]*image:/ && in_lvmPlugin { 
  in_image = 1
  next 
}

/^[[:space:]]*registry:/ && in_image && in_lvmPlugin && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*registry:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  plugin_reg = $0
  next 
}

/^[[:space:]]*repository:/ && in_image && in_lvmPlugin && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*repository:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  plugin_repo = $0
  next 
}

/^[[:space:]]*tag:/ && in_image && in_lvmPlugin && $0 !~ /^[[:space:]]*#/ { 
  gsub(/^[[:space:]]*tag:[[:space:]]*/, "")
  gsub(/[\"'\''"]/, "")
  gsub(/[[:space:]]*$/, "")
  plugin_tag = $0
  if (plugin_repo != "" && plugin_tag != "") {
    if (plugin_reg == "") {
      plugin_reg = "docker.io/"
    }
    gsub(/\/$/, "", plugin_reg)
    print plugin_reg "/" plugin_repo ":" plugin_tag
  }
  in_image = 0
  next
}

# Reset image state when leaving a section (non-indented line that's not a comment)
/^[^[:space:]#]/ {
  if (in_image) {
    in_image = 0
  }
}
AWKEOF
    
    while IFS= read -r line; do
      [[ -n "${line}" ]] && images+=("${line}")
    done < <(awk -f "${temp_awk}" "${values_file}" 2>/dev/null)
    
    rm -f "${temp_awk}"
  
  # Clean up
  rm -rf "${temp_dir}"
  
  # Filter out empty entries and return
  for img in "${images[@]}"; do
    if [[ -n "${img}" ]] && [[ "${img}" != "null" ]]; then
      echo "${img}"
    fi
  done
}

# Download Docker image and save as tar
download_image() {
  local image="$1"
  local image_name
  image_name=$(echo "${image}" | sed 's|[\/:]|_|g')
  local image_file="${OUTPUT_DIR}/images/${image_name}.tar"
  
  info "Downloading image: ${image}"
  
  # Pull image
  if ${CONTAINER_TOOL} pull "${image}" 2>&1 | tee -a "${DOWNLOAD_LOG_PATH}"; then
    info "Image pulled successfully"
  else
    error "Failed to pull image: ${image}"
  fi
  
  # Save image
  info "Saving image to: ${image_file}"
  if ${CONTAINER_TOOL} save "${image}" -o "${image_file}" 2>&1 | tee -a "${DOWNLOAD_LOG_PATH}"; then
    local image_size
    image_size=$(du -h "${image_file}" | cut -f1)
    info "Image saved: ${image_file} (${image_size})"
    
    # Verify file exists and has content
    if [[ ! -s "${image_file}" ]]; then
      error "Image file is empty: ${image_file}"
    fi
  else
    error "Failed to save image: ${image}"
  fi
}

# Download all images
download_images() {
  local chart_file="${OUTPUT_DIR}/charts/lvm-localpv-${CHART_VERSION}.tgz"
  
  if [[ ! -f "${chart_file}" ]]; then
    error "Chart file not found: ${chart_file}"
  fi
  
  info "Extracting image list from chart..."
  local images
  readarray -t images < <(extract_images_from_chart "${chart_file}")
  
  if [[ ${#images[@]} -eq 0 ]]; then
    error "No images found in chart"
  fi
  
  info "Found ${#images[@]} image(s) to download"
  info "Using container tool: ${CONTAINER_TOOL}"
  
  # Create images list file
  local images_list="${OUTPUT_DIR}/images/images.list"
  > "${images_list}"
  
  # Download each image
  for image in "${images[@]}"; do
    if [[ -n "${image}" ]]; then
      download_image "${image}"
      echo "${image}" >> "${images_list}"
    fi
  done
  
  info "All images downloaded successfully"
  info "Image list saved to: ${images_list}"
}

# Create manifest files
create_manifests() {
  info "Creating manifest files..."
  
  # Version file
  echo "${CHART_VERSION}" > "${OUTPUT_DIR}/manifests/VERSION"
  
  # Create README
  cat > "${OUTPUT_DIR}/README.md" <<EOF
# OpenEBS LVM LocalPV Offline Media

This directory contains all media required for offline installation of OpenEBS LVM LocalPV.

## Contents

- **charts/**: Helm charts
- **images/**: Docker images (as tar files)
- **manifests/**: Version and metadata files

## Version

Chart Version: ${CHART_VERSION}

## Usage

### 1. Load Images

First, load the Docker images into your local registry:

\`\`\`bash
# Using load-images.sh script (recommended)
./deploy/scripts/load-images.sh --images-dir ./images

# Or manually
for img in images/*.tar; do
  docker load -i "\$img"
done
\`\`\`

### 2. Install Using Media

\`\`\`bash
export OFFLINE_INSTALL=true
export OPENEBS_CHART_DIR="./charts/lvm-localpv-${CHART_VERSION}"
export OPENEBS_IMAGE_REGISTRY="your-registry.example.com"  # or use local registry

./deploy/scripts/install.sh --offline
\`\`\`

## Image List

\`\`\`
$(cat "${OUTPUT_DIR}/images/images.list" 2>/dev/null || echo "No images list found")
\`\`\`

## Notes

- All images are saved as tar files in the images/ directory
- Chart is saved as a tgz file in the charts/ directory
- After loading images, tag and push them to your private registry if needed
- See installation-scripts.md for detailed usage instructions

Generated by download.sh v${SCRIPT_VERSION}
EOF
  
  info "Manifest files created"
}

# Pack media into tar.gz
pack_media() {
  info "Packing media into tar.gz..."
  
  local pack_file="offline-media-${CHART_VERSION}.tar.gz"
  local pack_path="${OUTPUT_DIR}/../${pack_file}"
  
  info "Creating package: ${pack_path}"
  
  # Create checksum file
  info "Generating checksums..."
  find "${OUTPUT_DIR}" -type f -exec sha256sum {} \; > "${OUTPUT_DIR}/manifests/checksums.txt" 2>/dev/null || true
  
  # Pack
  if tar -czf "${pack_path}" -C "$(dirname "${OUTPUT_DIR}")" "$(basename "${OUTPUT_DIR}")" 2>&1 | tee -a "${DOWNLOAD_LOG_PATH}"; then
    local pack_size
    pack_size=$(du -h "${pack_path}" | cut -f1)
    info "Package created successfully: ${pack_path} (${pack_size})"
    info "You can transfer this file to your offline environment"
  else
    error "Failed to create package"
  fi
}

# Verify downloaded media
verify_media() {
  info "Verifying downloaded media..."
  
  local chart_file="${OUTPUT_DIR}/charts/lvm-localpv-${CHART_VERSION}.tgz"
  
  # Verify chart
  if [[ -f "${chart_file}" ]]; then
    if tar -tzf "${chart_file}" &>/dev/null; then
      info "Chart verification: OK"
    else
      error "Chart verification failed: ${chart_file}"
    fi
  else
    error "Chart file not found: ${chart_file}"
  fi
  
  # Verify images
  local image_count
  image_count=$(find "${OUTPUT_DIR}/images" -name "*.tar" -type f | wc -l)
  if [[ ${image_count} -gt 0 ]]; then
    info "Image verification: Found ${image_count} image file(s)"
    
    # Check a few images
    local checked=0
    shopt -s nullglob  # Handle case when no files match
    for img_file in "${OUTPUT_DIR}"/images/*.tar; do
      [[ ! -f "${img_file}" ]] && continue
      if [[ -s "${img_file}" ]]; then
        ((checked++))
        if [[ ${checked} -ge 3 ]]; then
          break
        fi
      fi
    done
    shopt -u nullglob  # Restore default behavior
    if [[ ${checked} -gt 0 ]]; then
      info "Sample image verification: OK"
    fi
  else
    error "No image files found"
  fi
  
  info "Media verification completed successfully"
}

main() {
  parse_args "$@"
  
  init_log
  info "OpenEBS LVM LocalPV Offline Media Download Script v${SCRIPT_VERSION}"
  
  # Record start time
  local start_time
  start_time=$(date +%s)
  
  check_prerequisites
  
  # Get chart version
  CHART_VERSION=$(get_chart_version)
  info "Chart version: ${CHART_VERSION}"
  
  # Validate version exists
  if ! validate_chart_version "${CHART_VERSION}"; then
    error "Invalid chart version: ${CHART_VERSION}

Please specify a valid version using --chart-version, or omit it to use the latest version."
  fi
  
  # Create directory structure
  create_media_structure
  
  # Download chart
  download_chart "${CHART_VERSION}"
  
  # Download images
  download_images
  
  # Create manifests
  create_manifests
  
  # Verify
  verify_media
  
  # Pack if requested
  if [[ "${PACK_MEDIA}" == "true" ]]; then
    pack_media
  fi
  
  # Calculate elapsed time
  local end_time elapsed_time hours minutes seconds
  end_time=$(date +%s)
  elapsed_time=$((end_time - start_time))
  
  hours=$((elapsed_time / 3600))
  minutes=$(((elapsed_time % 3600) / 60))
  seconds=$((elapsed_time % 60))
  
  # Format time string
  local time_str=""
  if [[ ${hours} -gt 0 ]]; then
    time_str="${hours}h ${minutes}m ${seconds}s"
  elif [[ ${minutes} -gt 0 ]]; then
    time_str="${minutes}m ${seconds}s"
  else
    time_str="${seconds}s"
  fi
  
  info "Download completed successfully!"
  info "Total time: ${time_str}"
  info "Media location: ${OUTPUT_DIR}"
  if [[ "${PACK_MEDIA}" == "true" ]]; then
    info "Package location: $(dirname "${OUTPUT_DIR}")/offline-media-${CHART_VERSION}.tar.gz"
  fi
  info "Next steps:"
  info "  1. Transfer media to offline environment"
  info "  2. Load images using: ./deploy/scripts/load-images.sh"
  info "  3. Install using: ./deploy/scripts/install.sh --offline"
}

main "$@"
