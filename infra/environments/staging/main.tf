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

module "compute" {
  source = "../../modules/compute"

  environment              = var.environment
  vpc_id                   = module.network.vpc_id
  private_subnet_ids       = module.network.private_subnet_ids
  allowed_cidr_blocks      = var.compute_allowed_cidr_blocks
  database_name            = var.database_name
  database_username        = var.database_username
  database_password        = var.database_password
  postgres_instance_class  = var.postgres_instance_class
  create_postgres          = var.create_postgres
  create_redis             = var.create_redis
  redis_node_type          = var.redis_node_type
  redis_num_cache_clusters = var.redis_num_cache_clusters
  create_mongo             = var.create_mongo
  mongo_username           = var.mongo_username
  mongo_password           = var.mongo_password
  mongo_instance_class     = var.mongo_instance_class
  create_kafka             = var.create_kafka
  kafka_version            = var.kafka_version
  kafka_number_of_broker_nodes = var.kafka_number_of_broker_nodes
  tags                     = local.tags
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
  cluster_name = module.kubernetes.cluster_name
  aws_region   = var.region
  enable_ha    = var.vault_enable_ha
  ingress_host = var.vault_ingress_host
  tags         = local.tags
}

module "gateway" {
  source = "../../modules/gateway"

  environment        = var.environment
  cluster_name       = module.kubernetes.cluster_name
  namespace          = var.gateway_namespace
  image_repository   = var.gateway_image_repository
  image_tag          = var.gateway_image_tag
  replica_count      = var.gateway_replica_count
  service_port       = var.gateway_service_port
  ingress_host       = var.gateway_ingress_host
  ingress_class_name = var.gateway_ingress_class_name
  ingress_annotations = var.gateway_ingress_annotations
  env_vars           = var.gateway_env
  resources          = var.gateway_resources
}

