#!/usr/bin/env bash

###############################################################################
# Compatibility Script - Redirects to new location
#
# This script has been moved to deploy/scripts/install.sh
# This file is kept for backward compatibility.
#
###############################################################################

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
NEW_SCRIPT="${SCRIPT_DIR}/scripts/install.sh"

if [[ -f "${NEW_SCRIPT}" ]]; then
  exec "${NEW_SCRIPT}" "$@"
else
  echo "Error: Installation script not found at ${NEW_SCRIPT}" >&2
  echo "Please ensure you have the latest version of the repository." >&2
  exit 1
fi
