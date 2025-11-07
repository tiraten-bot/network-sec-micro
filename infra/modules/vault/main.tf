variable "environment" {
  type        = string
  description = "Environment identifier"
}

variable "cluster_name" {
  type        = string
  description = "EKS cluster name where Vault will be installed"
}

variable "aws_region" {
  type        = string
  description = "AWS region hosting the cluster (required for auth helper)"
}

variable "enable_ha" {
  type        = bool
  description = "Whether to deploy Vault in HA mode with Raft storage"
  default     = true
}

variable "tags" {
  type        = map(string)
  description = "Metadata tags"
  default     = {}
}

variable "vault_image" {
  type        = string
  description = "Container image for Vault"
  default     = "hashicorp/vault:1.17.3"
}

variable "ingress_host" {
  type        = string
  description = "Ingress hostname for Vault UI/API"
  default     = ""
}

locals {
  namespace_labels = {
    "project"     = "network-sec"
    "environment" = var.environment
    "managed-by"  = "terraform"
  }
}

data "aws_eks_cluster" "this" {
  name = var.cluster_name
}

data "aws_eks_cluster_auth" "this" {
  name = var.cluster_name
}

provider "helm" {
  kubernetes {
    host                   = data.aws_eks_cluster.this.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.this.certificate_authority[0].data)
    token                  = data.aws_eks_cluster_auth.this.token
  }
}

resource "kubernetes_namespace" "vault" {
  metadata {
    name   = "vault"
    labels = local.namespace_labels
  }
}

resource "helm_release" "vault" {
  name       = "vault"
  repository = "https://helm.releases.hashicorp.com"
  chart      = "vault"
  namespace  = kubernetes_namespace.vault.metadata[0].name
  version    = "0.28.0"

  values = [
    yamlencode({
      global = {
        enabled = true
        image   = var.vault_image
      }
      server = {
        affinity = {
          podAntiAffinity = {
            preferredDuringSchedulingIgnoredDuringExecution = [{
              weight = 100
              podAffinityTerm = {
                topologyKey = "kubernetes.io/hostname"
                labelSelector = {
                  matchLabels = {
                    "app.kubernetes.io/name" = "vault"
                  }
                }
              }
            }]
          }
        }
        ha = {
          enabled = var.enable_ha
          raft = {
            enabled    = var.enable_ha
            setNodeId  = true
          }
        }
        ingress = {
          enabled          = length(var.ingress_host) > 0
          ingressClassName = "alb"
          hosts            = length(var.ingress_host) > 0 ? [var.ingress_host] : []
          annotations = {
            "alb.ingress.kubernetes.io/scheme"           = "internet-facing"
            "alb.ingress.kubernetes.io/target-type"      = "ip"
            "external-dns.alpha.kubernetes.io/hostname"  = var.ingress_host
          }
        }
        standalone = {
          enabled = !var.enable_ha
        }
        extraEnvironmentVars = {
          AWS_REGION = var.aws_region
        }
        serviceAccount = {
          create = true
          name   = "vault"
        }
        auditStorage = {
          enabled = true
        }
        dataStorage = {
          enabled = true
          size    = "20Gi"
        }
      }
    })
  ]

  depends_on = [kubernetes_namespace.vault]
}

output "namespace" {
  description = "Namespace where Vault is deployed"
  value       = kubernetes_namespace.vault.metadata[0].name
}

output "release_name" {
  description = "Helm release name"
  value       = helm_release.vault.name
}

