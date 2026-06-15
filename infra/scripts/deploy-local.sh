#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
TOFU_DIR="$ROOT/infra/terraform/k8s"
REGISTRY="${IMAGE_REGISTRY:-ghcr.io/abskrj}"

if [ -f "$ROOT/.env" ]; then
  set -a
  # shellcheck disable=SC1091
  source "$ROOT/.env"
  set +a
fi

: "${AWS_ACCESS_KEY_ID:?AWS_ACCESS_KEY_ID not set — add it to .env}"
: "${AWS_SECRET_ACCESS_KEY:?AWS_SECRET_ACCESS_KEY not set — add it to .env}"

AWS_REGION="${AWS_REGION:-us-east-1}"
EKS_CLUSTER_NAME="${EKS_CLUSTER_NAME:-velane-us-east-1}"
IMAGE_TAG="${1:-}"

if [ -z "$IMAGE_TAG" ]; then
  cd "$ROOT"
  git fetch --tags --force >/dev/null 2>&1 || true
  if IMAGE_TAG=$(git describe --tags --exact-match 2>/dev/null); then
    IMAGE_TAG="${IMAGE_TAG#v}"
  else
    IMAGE_TAG="sha-$(git rev-parse --short=7 HEAD)"
  fi
fi

export AWS_DEFAULT_REGION="$AWS_REGION"

aws eks update-kubeconfig \
  --region "$AWS_REGION" \
  --name "$EKS_CLUSTER_NAME" \
  --alias "$EKS_CLUSTER_NAME"

bash "$ROOT/infra/scripts/ci-prepare-k8s-tfvars.sh" \
  "$TOFU_DIR/terraform.tfvars" \
  "$REGISTRY" \
  "$IMAGE_TAG" \
  "$HOME/.kube/config" \
  "$EKS_CLUSTER_NAME"

cd "$TOFU_DIR"
tofu init -backend-config=backend.hcl -input=false
tofu apply -auto-approve

echo "Deployed Velane images tagged: $IMAGE_TAG"
