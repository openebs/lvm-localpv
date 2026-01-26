#!/usr/bin/env bash

###############################################################################
# OpenEBS LVM LocalPV Image Loading Script
#
# This script loads Docker images from tar files into the local container
# registry (Docker or Podman).
#
# Usage:
#   ./load-images.sh [OPTIONS]
#
# Options:
#   --help                  Show this help message
#   --images-dir DIR       Directory containing image tar files (default: ./images)
#   --registry REGISTRY    Target registry to tag and push images (optional)
#   --push                 Push images to registry after loading
#   --log-level LEVEL      Set log level (DEBUG, INFO, WARN, ERROR)
#
###############################################################################

set -euo pipefail

# Script metadata
readonly SCRIPT_NAME="load-images.sh"
readonly SCRIPT_VERSION="1.0.0"
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Default values
readonly DEFAULT_IMAGES_DIR="./images"
readonly DEFAULT_LOG_LEVEL="INFO"

# Global variables
IMAGES_DIR="${IMAGES_DIR:-${DEFAULT_IMAGES_DIR}}"
TARGET_REGISTRY="${TARGET_REGISTRY:-}"
PUSH_IMAGES="${PUSH_IMAGES:-false}"
LOG_LEVEL="${LOG_LEVEL:-${DEFAULT_LOG_LEVEL}}"
CONTAINER_TOOL=""

# Log file path
if [[ -w /tmp ]]; then
  LOAD_LOG_PATH="/tmp/openebs-lvmlocalpv_load-images-$(date +'%Y-%m-%d_%H-%M-%S').log"
else
  LOAD_LOG_PATH="./openebs-lvmlocalpv_load-images-$(date +'%Y-%m-%d_%H-%M-%S').log"
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
  echo "[${level}][${timestamp}]: ${message}" | tee -a "${LOAD_LOG_PATH}" >&2
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
  # Check for container tool
  if installed docker; then
    CONTAINER_TOOL="docker"
  elif installed podman; then
    CONTAINER_TOOL="podman"
  else
    error "Neither docker nor podman is installed. Please install one of them."
  fi
  
  info "Using container tool: ${CONTAINER_TOOL}"
  
  # Check if tool is running
  if ! ${CONTAINER_TOOL} info &>/dev/null; then
    error "${CONTAINER_TOOL} daemon is not running. Please start it first."
  fi
}

init_log() {
  touch "${LOAD_LOG_PATH}" || error "Cannot create log file: ${LOAD_LOG_PATH}"
  info "Log file: ${LOAD_LOG_PATH}"
  debug "Script version: ${SCRIPT_VERSION}"
}

show_help() {
  cat <<EOF
OpenEBS LVM LocalPV Image Loading Script v${SCRIPT_VERSION}

USAGE:
    ${SCRIPT_NAME} [OPTIONS]

OPTIONS:
    --help                  Show this help message and exit
    --images-dir DIR       Directory containing image tar files (default: ${DEFAULT_IMAGES_DIR})
    --registry REGISTRY    Target registry to tag and push images (optional)
    --push                 Push images to registry after loading and tagging
    --log-level LEVEL      Set log level: DEBUG, INFO, WARN, ERROR (default: ${DEFAULT_LOG_LEVEL})

ENVIRONMENT VARIABLES:
    IMAGES_DIR              Directory containing image tar files
    TARGET_REGISTRY         Target registry for tagging and pushing
    PUSH_IMAGES             Set to "true" to push images after loading
    LOG_LEVEL               Log level (DEBUG, INFO, WARN, ERROR)

EXAMPLES:
    # Load images from default directory
    ./load-images.sh

    # Load images from custom directory
    ./load-images.sh --images-dir /path/to/images

    # Load and push to registry
    ./load-images.sh --registry registry.example.com --push

    # Load from offline media directory
    ./load-images.sh --images-dir ./offline-media/images

NOTES:
    - Images are loaded into local ${CONTAINER_TOOL} registry
    - If --registry is specified, images will be tagged with registry prefix
    - Use --push to push tagged images to the registry
    - Original image names are preserved from the tar files

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
      --images-dir)
        IMAGES_DIR="$2"
        shift 2
        ;;
      --registry)
        TARGET_REGISTRY="$2"
        shift 2
        ;;
      --push)
        PUSH_IMAGES=true
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

# Load image from tar file
load_image() {
  local image_file="$1"
  local image_name
  image_name=$(basename "${image_file}" .tar)
  
  info "Loading image: ${image_file}"
  
  # Load image and capture output separately
  local load_output
  load_output=$(${CONTAINER_TOOL} load -i "${image_file}" 2>&1)
  local load_exit_code=$?
  
  # Log the output
  echo "${load_output}" | tee -a "${LOAD_LOG_PATH}" >&2
  
  if [[ ${load_exit_code} -eq 0 ]]; then
    info "Image loaded successfully: ${image_name}"
    
    # Get the actual image name from load output (only from stdout, not stderr)
    local loaded_image
    loaded_image=$(echo "${load_output}" | grep "Loaded image" | sed 's/.*Loaded image: //' | head -1 || echo "")
    
    if [[ -n "${loaded_image}" ]]; then
      echo "${loaded_image}"
    fi
  else
    error "Failed to load image: ${image_file}"
  fi
}

# Tag image for registry
tag_image() {
  local source_image="$1"
  local target_image="$2"
  
  info "Tagging image: ${source_image} -> ${target_image}"
  
  if ${CONTAINER_TOOL} tag "${source_image}" "${target_image}" 2>&1 | tee -a "${LOAD_LOG_PATH}"; then
    info "Image tagged successfully"
  else
    error "Failed to tag image: ${source_image}"
  fi
}

# Push image to registry
push_image() {
  local image="$1"
  
  info "Pushing image: ${image}"
  
  if ${CONTAINER_TOOL} push "${image}" 2>&1 | tee -a "${LOAD_LOG_PATH}"; then
    info "Image pushed successfully: ${image}"
  else
    error "Failed to push image: ${image}"
  fi
}

# Process images list file if exists
load_images_from_list() {
  local images_list="${IMAGES_DIR}/images.list"
  
  if [[ -f "${images_list}" ]]; then
    info "Found images list: ${images_list}"
    # This can be used for reference, but we'll load from tar files
  fi
}

# Main loading function
load_all_images() {
  if [[ ! -d "${IMAGES_DIR}" ]]; then
    error "Images directory does not exist: ${IMAGES_DIR}"
  fi
  
  local image_files
  readarray -t image_files < <(find "${IMAGES_DIR}" -name "*.tar" -type f | sort)
  
  if [[ ${#image_files[@]} -eq 0 ]]; then
    error "No image tar files found in: ${IMAGES_DIR}"
  fi
  
  info "Found ${#image_files[@]} image file(s) to load"
  
  local loaded_images=()
  
  # Load each image
  for image_file in "${image_files[@]}"; do
    local loaded_image
    loaded_image=$(load_image "${image_file}")
    if [[ -n "${loaded_image}" ]]; then
      loaded_images+=("${loaded_image}")
    fi
  done
  
  info "All images loaded successfully"
  
  # Tag and push if registry specified
  if [[ -n "${TARGET_REGISTRY}" ]]; then
    info "Tagging images for registry: ${TARGET_REGISTRY}"
    
    local registry_prefix="${TARGET_REGISTRY%/}"
    
    for source_image in "${loaded_images[@]}"; do
      # Extract image name without registry prefix
      # Examples:
      #   docker.io/openebs/lvm-driver:1.8.0 -> openebs/lvm-driver:1.8.0
      #   registry.k8s.io/sig-storage/csi-node-driver-registrar:v2.13.0 -> sig-storage/csi-node-driver-registrar:v2.13.0
      #   openebs/lvm-driver:1.8.0 -> openebs/lvm-driver:1.8.0 (already without registry, keep as is)
      local image_name
      
      # Handle different registry formats:
      # 1. docker.io/openebs/lvm-driver:1.8.0 -> openebs/lvm-driver:1.8.0
      # 2. registry.k8s.io/sig-storage/... -> sig-storage/...
      # 3. openebs/lvm-driver:1.8.0 -> openebs/lvm-driver:1.8.0 (no change, keep namespace)
      
      # Check if it starts with a known registry (contains . or : in the first segment)
      # Known registries: docker.io, registry.k8s.io, quay.io, gcr.io, etc.
      if [[ "${source_image}" =~ ^([^/]+[.:][^/]*)/(.+)$ ]]; then
        # Has registry prefix like docker.io/ or registry.k8s.io/ or localhost:5000/
        image_name="${BASH_REMATCH[2]}"
      else
        # No registry prefix (e.g., openebs/lvm-driver:1.8.0), use as is
        # This preserves the namespace/repository:tag structure
        image_name="${source_image}"
      fi
      
      # Create target image name
      # If registry ends with a path (e.g., localhost:35000/base), append image name
      # If registry is just host:port (e.g., localhost:35000), append image name directly
      local target_image="${registry_prefix}/${image_name}"
      
      tag_image "${source_image}" "${target_image}"
      
      # Push if requested
      if [[ "${PUSH_IMAGES}" == "true" ]]; then
        push_image "${target_image}"
      fi
    done
    
    if [[ "${PUSH_IMAGES}" == "true" ]]; then
      info "All images pushed to registry: ${TARGET_REGISTRY}"
    else
      info "Images tagged but not pushed. Use --push to push them."
    fi
  fi
  
  # List loaded images
  info "Loaded images:"
  for img in "${loaded_images[@]}"; do
    info "  - ${img}"
  done
}

main() {
  parse_args "$@"
  
  init_log
  info "OpenEBS LVM LocalPV Image Loading Script v${SCRIPT_VERSION}"
  
  check_prerequisites
  
  if [[ "${PUSH_IMAGES}" == "true" ]] && [[ -z "${TARGET_REGISTRY}" ]]; then
    error "--push requires --registry to be specified"
  fi
  
  load_images_from_list
  load_all_images
  
  info "Image loading completed successfully!"
  info "Next steps:"
  if [[ -z "${TARGET_REGISTRY}" ]]; then
    info "  1. Images are loaded in local ${CONTAINER_TOOL} registry"
    info "  2. Use them directly or tag/push to your registry"
  elif [[ "${PUSH_IMAGES}" != "true" ]]; then
    info "  1. Images are tagged for registry: ${TARGET_REGISTRY}"
    info "  2. Push them using: ${CONTAINER_TOOL} push <image>"
    info "  3. Or run this script again with --push"
  else
    info "  1. Images are loaded and pushed to: ${TARGET_REGISTRY}"
    info "  2. You can now use them in your installation"
  fi
}

main "$@"
