# Velane Kubernetes Terraform

This stack deploys the **core Velane workloads + optional Nango + subdomain-based Ingress** to an existing Kubernetes cluster (EKS, GKE, AKS, or self-managed).

## What this stack deploys

### Always
- `control-plane`
- `bun-executor`
- `python-executor`
- `admin`
- `mcp-server`
- Namespace (optional)
- Secrets for sensitive values

### When `deploy_nango = true` (default)
- Full Nango server Deployment (`nango`) with its own Service
- One-time Job that creates a **separate `nango` database** on your existing Postgres (unless you provide `nango_database_url`)
- Control-plane is automatically wired to talk to the in-cluster Nango

### When `enable_ingress = true` (default)
- Single `Ingress` resource with **subdomain-based routing** (works with ALB, nginx-ingress, etc.)

Default subdomain layout (assuming `base_domain = "example.com"`):
| Subdomain               | Backend             | Port |
|-------------------------|---------------------|------|
| `admin.example.com`     | admin               | 80   |
| `api.example.com`       | control-plane       | 8080 |
| `mcp.example.com`       | mcp-server          | 8090 |
| `connect.example.com`   | Nango Connect UI    | 3009 |
| `nango.example.com`     | Nango API           | 3003 |

## Cloud-agnostic Ingress (ALB and friends)

Set `ingress_class_name`:

- **AWS EKS** → `"alb"` (requires aws-load-balancer-controller installed)
- **nginx-ingress** → `"nginx"`
- **GKE** → `"gce"` or `"nginx"`
- **AKS** → `"azure/application-gateway"` or nginx, etc.

Recommended ALB annotations are automatically added when you choose `ingress_class_name = "alb"`. You can add more via `ingress_annotations`.

No domain/TLS is configured by this stack (you can terminate at the ALB or add cert-manager later).

## Nango Database Strategy (exactly what you asked for)

- You point the stack at your main Velane Postgres via `database_url`
- If you leave `nango_database_url` empty (the common case):
  - The stack creates a **dedicated `nango` database** on the **same Postgres server**
  - A Kubernetes Job runs `CREATE DATABASE nango` (idempotent) the first time
- If you provide a full `nango_database_url`, the stack uses it as-is and skips DB creation

This gives you separate databases while still using one managed Postgres instance.

## Quick Start

```bash
cd infra/terraform/k8s
cp terraform.tfvars.example terraform.tfvars
# edit the file — at minimum set images + database_url + redis_url + keys
terraform init
terraform plan
terraform apply
```

After apply on EKS with ALB:

```bash
kubectl get ingress -n velane velane
# copy the hostname from status.loadBalancer.ingress[0].hostname
```

Then point your DNS records for the subdomains to the ALB DNS name. The Nango URLs are automatically derived from your `base_domain` and subdomains, so no extra configuration is needed.

## Prerequisites

- Terraform ≥ 1.5
- Existing Kubernetes cluster with `kubectl` access
- (For ALB) aws-load-balancer-controller installed on the EKS cluster
- Container images available to the cluster

## EKS ALB Checklist

1. Install the AWS Load Balancer Controller (official docs).
2. Use `ingress_class_name = "alb"`.
3. After the ALB is created, point your DNS records (A/CNAME) for all subdomains (`admin`, `api`, `mcp`, `connect`, `nango`) to the ALB's DNS name.
4. The control-plane configures itself to use these subdomains automatically based on your `base_domain` variable.

## What this stack still does NOT do

- Does not create the Kubernetes cluster itself
- Does not provision Postgres / Redis / ClickHouse (you bring them)
- Does not set up DNS or TLS certificates (you can add cert-manager or ALB ACM later)
