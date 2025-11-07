variable "environment" {
  type        = string
  description = "Environment identifier"
}

variable "cluster_name" {
  type        = string
  description = "Target Kubernetes cluster name"
}

variable "grafana_admin_user" {
  type        = string
  description = "Grafana admin username"
}

variable "grafana_admin_password" {
  type        = string
  description = "Grafana admin password"
  sensitive   = true
}

variable "tags" {
  type        = map(string)
  description = "Tags passed for metadata (informational only)"
  default     = {}
}

provider "helm" {
  kubernetes {
    host                   = data.aws_eks_cluster.this.endpoint
    cluster_ca_certificate = base64decode(data.aws_eks_cluster.this.certificate_authority[0].data)
    token                  = data.aws_eks_cluster_auth.this.token
  }
}

data "aws_eks_cluster" "this" {
  name = var.cluster_name
}

data "aws_eks_cluster_auth" "this" {
  name = var.cluster_name
}

resource "kubernetes_namespace" "monitoring" {
  metadata {
    name = "monitoring"
    labels = {
      "project"     = "network-sec"
      "environment" = var.environment
      "managed-by"  = "terraform"
    }
  }
}

resource "helm_release" "prometheus" {
  name       = "kube-prometheus-stack"
  namespace  = kubernetes_namespace.monitoring.metadata[0].name
  repository = "https://prometheus-community.github.io/helm-charts"
  chart      = "kube-prometheus-stack"
  version    = "58.3.0"

  values = [
    yamlencode({
      grafana = {
        adminUser     = var.grafana_admin_user
        adminPassword = var.grafana_admin_password
      }
      defaultRules = {
        create = true
      }
      prometheus = {
        prometheusSpec = {
          retention           = "15d"
          retentionSize       = "50GiB"
          serviceMonitorSelectorNilUsesHelmValues = false
        }
      }
    })
  ]

  depends_on = [kubernetes_namespace.monitoring]
}

output "namespace" {
  description = "Namespace where monitoring stack is installed"
  value       = kubernetes_namespace.monitoring.metadata[0].name
}

