variable "region" {
  type        = string
  description = "AWS region used for the dev environment"
  default     = "eu-central-1"
}

variable "environment" {
  type        = string
  description = "Logical environment name"
  default     = "dev"
}

variable "account_id" {
  type        = string
  description = "AWS account ID that hosts the dev environment"
}

variable "vpc_cidr" {
  type        = string
  description = "Base CIDR block for the environment VPC"
  default     = "10.50.0.0/16"
}

variable "public_subnets" {
  type        = list(string)
  description = "Public subnet CIDR blocks"
  default     = ["10.50.0.0/20", "10.50.16.0/20", "10.50.32.0/20"]
}

variable "private_subnets" {
  type        = list(string)
  description = "Private subnet CIDR blocks"
  default     = ["10.50.48.0/20", "10.50.64.0/20", "10.50.80.0/20"]
}

variable "grafana_admin_user" {
  type        = string
  description = "Grafana admin username for the monitoring stack"
  default     = "admin"
}

variable "grafana_admin_password" {
  type        = string
  description = "Grafana admin password (use Vault/SSM in real deployments)"
  sensitive   = true
}

variable "database_name" {
  type        = string
  description = "Primary PostgreSQL database name"
  default     = "networksec"
}

variable "database_username" {
  type        = string
  description = "Primary PostgreSQL username"
  default     = "warrior"
}

variable "database_password" {
  type        = string
  description = "Primary PostgreSQL password"
  sensitive   = true
}

variable "postgres_instance_class" {
  type        = string
  description = "Instance class for PostgreSQL RDS"
  default     = "db.t3.medium"
}

variable "create_postgres" {
  type        = bool
  description = "Toggle for provisioning PostgreSQL"
  default     = true
}

variable "create_redis" {
  type        = bool
  description = "Toggle for provisioning Redis"
  default     = true
}

variable "redis_node_type" {
  type        = string
  description = "Instance class for Redis"
  default     = "cache.t3.small"
}

variable "redis_num_cache_clusters" {
  type        = number
  description = "Number of Redis nodes"
  default     = 2
}

variable "create_mongo" {
  type        = bool
  description = "Toggle for provisioning DocumentDB"
  default     = true
}

variable "mongo_username" {
  type        = string
  description = "Master username for DocumentDB"
  default     = "root"
}

variable "mongo_password" {
  type        = string
  description = "Master password for DocumentDB"
  sensitive   = true
}

variable "mongo_instance_class" {
  type        = string
  description = "Instance class for DocumentDB"
  default     = "db.r6g.large"
}

variable "create_kafka" {
  type        = bool
  description = "Toggle for provisioning MSK cluster"
  default     = true
}

variable "kafka_version" {
  type        = string
  description = "Kafka version"
  default     = "3.6.0"
}

variable "kafka_number_of_broker_nodes" {
  type        = number
  description = "MSK broker count"
  default     = 3
}

variable "compute_allowed_cidr_blocks" {
  type        = list(string)
  description = "CIDR blocks with access to managed data services"
  default     = ["10.0.0.0/8"]
}

variable "vault_enable_ha" {
  type        = bool
  description = "Deploy Vault with integrated Raft HA mode"
  default     = true
}

variable "vault_ingress_host" {
  type        = string
  description = "Hostname for Vault ingress"
  default     = "vault.dev.network-sec.local"
}

variable "gateway_namespace" {
  type        = string
  description = "Kubernetes namespace for the fiber gateway"
  default     = "gateway"
}

variable "gateway_image_repository" {
  type        = string
  description = "Container repository for the fiber gateway"
  default     = "ghcr.io/network-sec/fiber-gateway"
}

variable "gateway_image_tag" {
  type        = string
  description = "Container image tag for the fiber gateway"
  default     = "latest"
}

variable "gateway_replica_count" {
  type        = number
  description = "Number of fiber gateway replicas"
  default     = 2
}

variable "gateway_service_port" {
  type        = number
  description = "Container port exposed by fiber gateway"
  default     = 8090
}

variable "gateway_ingress_host" {
  type        = string
  description = "Hostname for fiber gateway ingress"
  default     = "api.dev.network-sec.local"
}

variable "gateway_ingress_class_name" {
  type        = string
  description = "Ingress class for fiber gateway"
  default     = "alb"
}

variable "gateway_ingress_annotations" {
  type        = map(string)
  description = "Additional annotations for the fiber gateway ingress"
  default     = {}
}

variable "gateway_env" {
  type        = map(string)
  description = "Environment variables for the fiber gateway deployment"
  default = {
    "UPSTREAM_WARRIOR" = "http://warrior:8080"
    "UPSTREAM_WEAPON"  = "http://weapon:8081"
    "UPSTREAM_BATTLE"  = "http://battle:8085"
  }
}

variable "gateway_resources" {
  type = object({
    requests = map(string)
    limits   = map(string)
  })
  description = "Resource requests and limits for the fiber gateway"
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

