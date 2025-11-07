variable "environment" {
  type        = string
  description = "Environment identifier (dev/staging/prod)"
}

variable "cluster_name" {
  type        = string
  description = "EKS cluster where the gateway will run"
}

variable "namespace" {
  type        = string
  description = "Kubernetes namespace for the gateway deployment"
  default     = "gateway"
}

variable "image_repository" {
  type        = string
  description = "Container image repository for the fiber gateway"
}

variable "image_tag" {
  type        = string
  description = "Container image tag for the fiber gateway"
}

variable "replica_count" {
  type        = number
  description = "Desired number of replicas"
  default     = 2
}

variable "service_port" {
  type        = number
  description = "Service port exposed by the gateway"
  default     = 8090
}

variable "ingress_host" {
  type        = string
  description = "Hostname for ingress exposure"
}

variable "ingress_class_name" {
  type        = string
  description = "Ingress class name (e.g., alb, nginx)"
  default     = "alb"
}

variable "ingress_annotations" {
  type        = map(string)
  description = "Additional annotations for ingress resource"
  default     = {}
}

variable "env_vars" {
  type        = map(string)
  description = "Environment variables injected into the deployment"
  default     = {}
}

variable "resources" {
  type = object({
    requests = map(string)
    limits   = map(string)
  })
  description = "Container resource requests/limits"
  default = {
    requests = {
      cpu    = "250m"
      memory = "256Mi"
    }
    limits = {
      cpu    = "500m"
      memory = "512Mi"
    }
  }
}

locals {
  labels = {
    "app.kubernetes.io/name"       = "fiber-gateway"
    "app.kubernetes.io/part-of"    = "network-sec"
    "app.kubernetes.io/component"  = "api-gateway"
    "app.kubernetes.io/managed-by" = "terraform"
    "app.kubernetes.io/environment" = var.environment
  }
}

data "aws_eks_cluster" "this" {
  name = var.cluster_name
}

data "aws_eks_cluster_auth" "this" {
  name = var.cluster_name
}

provider "kubernetes" {
  host                   = data.aws_eks_cluster.this.endpoint
  cluster_ca_certificate = base64decode(data.aws_eks_cluster.this.certificate_authority[0].data)
  token                  = data.aws_eks_cluster_auth.this.token
}

resource "kubernetes_namespace" "this" {
  metadata {
    name   = var.namespace
    labels = {
      "project"     = "network-sec"
      "environment" = var.environment
      "managed-by"  = "terraform"
    }
  }
}

resource "kubernetes_config_map" "gateway_env" {
  metadata {
    name      = "gateway-env"
    namespace = kubernetes_namespace.this.metadata[0].name
  }

  data = var.env_vars
}

resource "kubernetes_deployment" "gateway" {
  metadata {
    name      = "fiber-gateway"
    namespace = kubernetes_namespace.this.metadata[0].name
    labels    = local.labels
  }

  spec {
    replicas = var.replica_count

    selector {
      match_labels = local.labels
    }

    template {
      metadata {
        labels = local.labels
      }

      spec {
        container {
          name  = "gateway"
          image = "${var.image_repository}:${var.image_tag}"

          port {
            name           = "http"
            container_port = var.service_port
          }

          env_from {
            config_map_ref {
              name = kubernetes_config_map.gateway_env.metadata[0].name
            }
          }

          liveness_probe {
            http_get {
              path = "/health"
              port = var.service_port
            }
            initial_delay_seconds = 15
            period_seconds        = 20
          }

          readiness_probe {
            http_get {
              path = "/health"
              port = var.service_port
            }
            initial_delay_seconds = 10
            period_seconds        = 10
          }

          resources {
            limits   = var.resources.limits
            requests = var.resources.requests
          }
        }

        security_context {
          run_as_non_root = true
          fs_group        = 2000
        }

        termination_grace_period_seconds = 30
      }
    }
  }

  depends_on = [kubernetes_config_map.gateway_env]
}

resource "kubernetes_service" "gateway" {
  metadata {
    name      = "fiber-gateway"
    namespace = kubernetes_namespace.this.metadata[0].name
    labels    = local.labels
  }

  spec {
    selector = local.labels

    port {
      name        = "http"
      port        = 80
      target_port = var.service_port
    }

    type = "ClusterIP"
  }
}

resource "kubernetes_ingress_v1" "gateway" {
  metadata {
    name      = "fiber-gateway"
    namespace = kubernetes_namespace.this.metadata[0].name
    labels    = local.labels
    annotations = merge({
      "alb.ingress.kubernetes.io/scheme"          = "internet-facing"
      "alb.ingress.kubernetes.io/target-type"     = "ip"
      "external-dns.alpha.kubernetes.io/hostname" = var.ingress_host
    }, var.ingress_annotations)
  }

  spec {
    ingress_class_name = var.ingress_class_name

    rule {
      host = var.ingress_host
      http {
        path {
          path      = "/"
          path_type = "Prefix"
          backend {
            service {
              name = kubernetes_service.gateway.metadata[0].name
              port {
                number = 80
              }
            }
          }
        }
      }
    }

    tls {
      hosts      = [var.ingress_host]
      secret_name = "fiber-gateway-tls"
    }
  }
}

output "namespace" {
  description = "Namespace where the fiber gateway is deployed"
  value       = kubernetes_namespace.this.metadata[0].name
}

output "service_name" {
  description = "Cluster service name for the gateway"
  value       = kubernetes_service.gateway.metadata[0].name
}

output "ingress_host" {
  description = "Ingress hostname used to reach the gateway"
  value       = var.ingress_host
}
