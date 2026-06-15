#!/usr/bin/env bash
set -euo pipefail

TFVARS="${1:?terraform.tfvars path required}"
REGISTRY="${2:?image registry required, e.g. ghcr.io/owner}"
TAG="${3:?image tag required}"
KUBECONFIG_PATH="${4:-/home/runner/.kube/config}"
KUBECONFIG_CONTEXT="${5:?kubeconfig context required}"

sed_inplace() {
  if [[ "${OSTYPE:-}" == darwin* ]]; then
    sed -i '' "$@"
  else
    sed -i "$@"
  fi
}

update_image() {
  local var="$1"
  local image="$2"
  sed_inplace "s|^${var}[[:space:]]*=.*|${var}   = \"${REGISTRY}/${image}:${TAG}\"|" "$TFVARS"
}

update_image control_plane_image velane-control-plane
update_image bun_executor_image velane-bun-executor
update_image python_executor_image velane-python-executor
update_image admin_image velane-admin
update_image mcp_server_image velane-mcp-server

sed_inplace "s|^kubeconfig_path[[:space:]]*=.*|kubeconfig_path    = \"${KUBECONFIG_PATH}\"|" "$TFVARS"
sed_inplace "s|^kubeconfig_context[[:space:]]*=.*|kubeconfig_context = \"${KUBECONFIG_CONTEXT}\"|" "$TFVARS"
