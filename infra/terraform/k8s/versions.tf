terraform {
  required_version = ">= 1.5.0"

  backend "s3" {}

  required_providers {
    kubernetes = {
      source  = "hashicorp/kubernetes"
      version = "~> 2.33"
    }
  }
}
