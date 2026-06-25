provider "aws" {
  region = var.region
}

provider "kubernetes" {
  config_path    = var.kubeconfig_path
  config_context = var.kubeconfig_context != "" ? var.kubeconfig_context : null
}

# ---------------------------------------------------------------------------
# Read existing EKS cluster to get VPC and node security group
# ---------------------------------------------------------------------------

data "aws_eks_cluster" "main" {
  name = var.cluster_name
}

data "aws_subnets" "private" {
  filter {
    name   = "vpc-id"
    values = [data.aws_eks_cluster.main.vpc_config[0].vpc_id]
  }
  filter {
    name   = "tag:kubernetes.io/role/internal-elb"
    values = ["1"]
  }
}

data "aws_security_group" "nodes" {
  name   = "${var.cluster_name}-nodes-sg"
  vpc_id = data.aws_eks_cluster.main.vpc_config[0].vpc_id
}

locals {
  default_tags = merge(
    {
      Project   = "velane-licensing"
      ManagedBy = "terraform"
    },
    var.tags,
  )
  tls_enabled = var.acm_certificate_arn != ""
  alb_listen_ports = local.tls_enabled ? "[{\"HTTP\":80},{\"HTTPS\":443}]" : "[{\"HTTP\":80}]"
}

# ---------------------------------------------------------------------------
# RDS — isolated Postgres for licensing data
# ---------------------------------------------------------------------------

resource "random_password" "db" {
  length  = 32
  special = false
}

resource "aws_security_group" "licensing_db" {
  name        = "velane-licensing-db-sg"
  description = "Allow licensing server pods to reach the licensing RDS instance."
  vpc_id      = data.aws_eks_cluster.main.vpc_config[0].vpc_id

  ingress {
    description     = "Postgres from EKS nodes"
    from_port       = 5432
    to_port         = 5432
    protocol        = "tcp"
    security_groups = [data.aws_security_group.nodes.id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.default_tags
}

resource "aws_db_subnet_group" "licensing" {
  name       = "velane-licensing-subnet-group"
  subnet_ids = data.aws_subnets.private.ids
  tags       = local.default_tags
}

resource "aws_db_instance" "licensing" {
  identifier        = "velane-licensing"
  engine            = "postgres"
  engine_version    = "16"
  instance_class    = var.db_instance_class
  allocated_storage = var.db_allocated_storage
  storage_encrypted = true

  db_name  = "licensing"
  username = "licensing"
  password = random_password.db.result

  db_subnet_group_name   = aws_db_subnet_group.licensing.name
  vpc_security_group_ids = [aws_security_group.licensing_db.id]
  publicly_accessible    = false

  backup_retention_period = 7
  skip_final_snapshot     = false
  final_snapshot_identifier = "velane-licensing-final"

  tags = local.default_tags
}

# ---------------------------------------------------------------------------
# Kubernetes — namespace, secret, deployment, service, ingress
# ---------------------------------------------------------------------------

resource "kubernetes_namespace_v1" "licensing" {
  metadata {
    name = "velane-licensing"
    labels = {
      "app.kubernetes.io/part-of"    = "velane"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }
}

resource "kubernetes_secret_v1" "licensing" {
  metadata {
    name      = "license-server-secrets"
    namespace = kubernetes_namespace_v1.licensing.metadata[0].name
  }

  type = "Opaque"

  data = {
    DATABASE_URL    = "postgres://licensing:${random_password.db.result}@${aws_db_instance.licensing.address}:5432/licensing"
    PRIVATE_KEY_PEM = var.private_key_pem
  }
}

resource "kubernetes_deployment_v1" "licensing" {
  metadata {
    name      = "license-server"
    namespace = kubernetes_namespace_v1.licensing.metadata[0].name
    labels = {
      "app.kubernetes.io/name"       = "license-server"
      "app.kubernetes.io/part-of"    = "velane"
      "app.kubernetes.io/managed-by" = "terraform"
    }
  }

  spec {
    replicas = 1

    selector {
      match_labels = { app = "license-server" }
    }

    template {
      metadata {
        labels = { app = "license-server" }
      }

      spec {
        container {
          name              = "license-server"
          image             = var.license_server_image
          image_pull_policy = var.image_pull_policy

          port {
            container_port = 8070
            name           = "http"
          }

          env {
            name  = "PORT"
            value = "8070"
          }

          env {
            name = "DATABASE_URL"
            value_from {
              secret_key_ref {
                name = kubernetes_secret_v1.licensing.metadata[0].name
                key  = "DATABASE_URL"
              }
            }
          }

          env {
            name = "PRIVATE_KEY_PEM"
            value_from {
              secret_key_ref {
                name = kubernetes_secret_v1.licensing.metadata[0].name
                key  = "PRIVATE_KEY_PEM"
              }
            }
          }

          liveness_probe {
            http_get {
              path = "/healthz"
              port = 8070
            }
            initial_delay_seconds = 5
            period_seconds        = 10
          }

          readiness_probe {
            http_get {
              path = "/healthz"
              port = 8070
            }
            initial_delay_seconds = 3
            period_seconds        = 5
          }
        }
      }
    }
  }

  depends_on = [aws_db_instance.licensing]
}

resource "kubernetes_service_v1" "licensing" {
  metadata {
    name      = "license-server"
    namespace = kubernetes_namespace_v1.licensing.metadata[0].name
  }

  spec {
    selector = { app = "license-server" }

    port {
      name        = "http"
      port        = 8070
      target_port = 8070
    }

    type = "ClusterIP"
  }
}

# Ingress joins the existing Velane ALB via group.name annotation.
# This adds a new host rule without provisioning a second load balancer.
resource "kubernetes_ingress_v1" "licensing" {
  metadata {
    name      = "license-server"
    namespace = kubernetes_namespace_v1.licensing.metadata[0].name

    annotations = merge(
      {
        "kubernetes.io/ingress.class"                = "alb"
        "alb.ingress.kubernetes.io/scheme"           = "internet-facing"
        "alb.ingress.kubernetes.io/target-type"      = "ip"
        "alb.ingress.kubernetes.io/listen-ports"     = local.alb_listen_ports
        "alb.ingress.kubernetes.io/group.name"       = "velane"
      },
      local.tls_enabled ? {
        "alb.ingress.kubernetes.io/certificate-arn" = var.acm_certificate_arn
        "alb.ingress.kubernetes.io/ssl-redirect"    = "443"
      } : {}
    )
  }

  spec {
    ingress_class_name = "alb"

    rule {
      host = "license.${var.base_domain}"
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = kubernetes_service_v1.licensing.metadata[0].name
              port { number = 8070 }
            }
          }
        }
      }
    }
  }
}
