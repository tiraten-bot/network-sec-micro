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

