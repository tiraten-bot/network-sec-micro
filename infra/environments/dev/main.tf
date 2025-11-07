locals {
  tags = {
    project     = "network-sec"
    environment = var.environment
    managed_by  = "terraform"
  }
}

provider "aws" {
  region = var.region

  assume_role {
    role_arn = "arn:aws:iam::${var.account_id}:role/terraform-admin"
  }

  default_tags {
    tags = local.tags
  }
}

# Optional: configure Kubernetes and Helm providers once the cluster exists.
# provider "kubernetes" {
#   host                   = module.eks.cluster_endpoint
#   cluster_ca_certificate = base64decode(module.eks.cluster_certificate_authority_data)
#   token                  = data.aws_eks_cluster_auth.this.token
# }

module "network" {
  source = "../../modules/network"

  environment     = var.environment
  vpc_cidr        = var.vpc_cidr
  public_subnets  = var.public_subnets
  private_subnets = var.private_subnets
  tags            = local.tags
}

module "security" {
  source = "../../modules/security"

  environment = var.environment
  vpc_id      = module.network.vpc_id
  tags        = local.tags
}

module "kubernetes" {
  source = "../../modules/kubernetes"

  environment        = var.environment
  cluster_name       = "${var.environment}-network-sec-eks"
  vpc_id             = module.network.vpc_id
  private_subnet_ids = module.network.private_subnet_ids
  tags               = local.tags
}

module "monitoring" {
  source = "../../modules/monitoring"

  environment            = var.environment
  cluster_name           = module.kubernetes.cluster_name
  grafana_admin_user     = var.grafana_admin_user
  grafana_admin_password = var.grafana_admin_password
  tags                   = local.tags
}

module "vault" {
  source = "../../modules/vault"

  environment = var.environment
  vpc_id      = module.network.vpc_id
  tags        = local.tags
}

module "gateway" {
  source = "../../modules/gateway"

  environment        = var.environment
  vpc_id             = module.network.vpc_id
  public_subnet_ids  = module.network.public_subnet_ids
  private_subnet_ids = module.network.private_subnet_ids
  tags               = local.tags
}

output "cluster_name" {
  description = "Name of the EKS cluster deployed in the environment"
  value       = module.kubernetes.cluster_name
}

output "vpc_id" {
  description = "ID of the environment VPC"
  value       = module.network.vpc_id
}

