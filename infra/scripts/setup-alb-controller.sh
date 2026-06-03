#!/usr/bin/env bash
# =============================================================================
# setup-alb-controller.sh
#
# Installs the AWS Load Balancer Controller into an EKS cluster via Helm.
# Run this once after "terraform apply" on the aws-eks module.
#
# Usage:
#   bash infra/scripts/setup-alb-controller.sh \
#     --cluster velane-us-east-1 \
#     --region  us-east-1 \
#     --role-arn arn:aws:iam::123456789012:role/velane-us-east-1-alb-controller
#
# Or set the env vars directly:
#   CLUSTER_NAME, AWS_REGION, ALB_CONTROLLER_ROLE_ARN
#
# The role ARN comes from:
#   terraform -chdir=infra/terraform/aws-eks output -raw alb_controller_role_arn
# =============================================================================

set -euo pipefail

# ── helpers ────────────────────────────────────────────────────────────────

info()  { echo "[INFO]  $*"; }
error() { echo "[ERROR] $*" >&2; exit 1; }

check_dep() {
  command -v "$1" &>/dev/null || error "'$1' is required but not installed. $2"
}

# ── argument parsing ────────────────────────────────────────────────────────

CLUSTER_NAME="${CLUSTER_NAME:-}"
AWS_REGION="${AWS_REGION:-us-east-1}"
ALB_CONTROLLER_ROLE_ARN="${ALB_CONTROLLER_ROLE_ARN:-}"

while [[ $# -gt 0 ]]; do
  case $1 in
    --cluster)   CLUSTER_NAME="$2";             shift 2 ;;
    --region)    AWS_REGION="$2";               shift 2 ;;
    --role-arn)  ALB_CONTROLLER_ROLE_ARN="$2";  shift 2 ;;
    *)           error "Unknown argument: $1" ;;
  esac
done

[[ -n "$CLUSTER_NAME" ]]           || error "--cluster (or CLUSTER_NAME env var) is required"
[[ -n "$ALB_CONTROLLER_ROLE_ARN" ]] || error "--role-arn (or ALB_CONTROLLER_ROLE_ARN env var) is required"

# ── dependency checks ────────────────────────────────────────────────────────

check_dep helm    "Install from https://helm.sh/docs/intro/install/"
check_dep kubectl "Install from https://kubernetes.io/docs/tasks/tools/"
check_dep aws     "Install from https://aws.amazon.com/cli/"

# ── refresh kubeconfig ────────────────────────────────────────────────────────

info "Refreshing kubeconfig for cluster '$CLUSTER_NAME' in '$AWS_REGION'..."
aws eks update-kubeconfig \
  --region "$AWS_REGION" \
  --name   "$CLUSTER_NAME" \
  --alias  "$CLUSTER_NAME"

# ── add Helm repo ─────────────────────────────────────────────────────────────

info "Adding eks Helm repo..."
helm repo add eks https://aws.github.io/eks-charts 2>/dev/null || true
helm repo update eks

# ── install / upgrade ─────────────────────────────────────────────────────────

info "Installing aws-load-balancer-controller..."
helm upgrade --install aws-load-balancer-controller eks/aws-load-balancer-controller \
  --namespace kube-system \
  --set clusterName="$CLUSTER_NAME" \
  --set serviceAccount.create=true \
  --set serviceAccount.name=aws-load-balancer-controller \
  --set "serviceAccount.annotations.eks\\.amazonaws\\.com/role-arn=$ALB_CONTROLLER_ROLE_ARN" \
  --set region="$AWS_REGION" \
  --set vpcId="$(aws eks describe-cluster --name "$CLUSTER_NAME" --region "$AWS_REGION" --query 'cluster.resourcesVpcConfig.vpcId' --output text)" \
  --wait \
  --timeout 3m

# ── verify ────────────────────────────────────────────────────────────────────

info "Waiting for controller pods to be ready..."
kubectl rollout status deployment/aws-load-balancer-controller -n kube-system --timeout=120s

info ""
info "AWS Load Balancer Controller installed successfully."
info ""
info "Next steps:"
info "  1. Validate your ACM certificate in Cloudflare (add the CNAME from terraform output acm_dns_validation_records)"
info "  2. Once the cert status is ISSUED, run: cd infra/terraform/k8s && terraform apply"
info "  3. Get the ALB hostname:  kubectl get ingress -n velane velane -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'"
info "  4. Add a wildcard CNAME in Cloudflare: *.yourdomain.com → <ALB hostname> (Proxied / orange cloud)"
