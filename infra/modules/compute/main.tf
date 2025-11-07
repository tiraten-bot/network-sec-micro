variable "environment" {
  type        = string
  description = "Environment identifier (dev/staging/prod)"
}

variable "vpc_id" {
  type        = string
  description = "VPC hosting the data services"
}

variable "private_subnet_ids" {
  type        = list(string)
  description = "Private subnet IDs for data planes"
}

variable "allowed_cidr_blocks" {
  type        = list(string)
  description = "CIDR blocks allowed to access managed data services"
  default     = []
}

variable "tags" {
  type        = map(string)
  description = "Common tags to apply"
  default     = {}
}

variable "create_postgres" {
  type        = bool
  description = "Whether to provision the PostgreSQL RDS instance"
  default     = true
}

variable "database_name" {
  type        = string
  description = "Default database name for PostgreSQL"
  default     = "networksec"
}

variable "database_username" {
  type        = string
  description = "Master username for PostgreSQL"
  default     = "warrior"
}

variable "database_password" {
  type        = string
  description = "Master password for PostgreSQL"
  sensitive   = true
}

variable "postgres_instance_class" {
  type        = string
  description = "Instance class for the PostgreSQL RDS instance"
  default     = "db.t3.medium"
}

variable "postgres_engine_version" {
  type        = string
  description = "PostgreSQL engine version"
  default     = "14.10"
}

variable "create_redis" {
  type        = bool
  description = "Whether to provision the Redis Elasticache replication group"
  default     = true
}

variable "redis_node_type" {
  type        = string
  description = "Instance class for Redis nodes"
  default     = "cache.t3.small"
}

variable "redis_num_cache_clusters" {
  type        = number
  description = "Number of cache nodes in the Redis replication group"
  default     = 2
}

variable "create_mongo" {
  type        = bool
  description = "Whether to provision the DocumentDB cluster (Mongo-compatible)"
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
  description = "Instance class for DocumentDB instances"
  default     = "db.r6g.large"
}

variable "create_kafka" {
  type        = bool
  description = "Whether to provision the MSK (Managed Kafka) cluster"
  default     = true
}

variable "kafka_version" {
  type        = string
  description = "Kafka version for the MSK cluster"
  default     = "3.6.0"
}

variable "kafka_number_of_broker_nodes" {
  type        = number
  description = "Number of broker nodes for MSK"
  default     = 3
}

locals {
  base_tags = merge(var.tags, {
    project     = "network-sec"
    environment = var.environment
    managed_by  = "terraform"
  })
}

# -----------------------------------------------------------------------------
# PostgreSQL (Amazon RDS)
# -----------------------------------------------------------------------------

resource "aws_security_group" "postgres" {
  count       = var.create_postgres ? 1 : 0
  name        = "${var.environment}-postgres-sg"
  description = "PostgreSQL access rules"
  vpc_id      = var.vpc_id

  dynamic "ingress" {
    for_each = var.allowed_cidr_blocks
    content {
      description = "PostgreSQL ingress"
      from_port   = 5432
      to_port     = 5432
      protocol    = "tcp"
      cidr_blocks = [ingress.value]
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.base_tags, {
    Name = "${var.environment}-postgres-sg"
  })
}

resource "aws_db_subnet_group" "postgres" {
  count      = var.create_postgres ? 1 : 0
  name       = "${var.environment}-postgres-subnets"
  subnet_ids = var.private_subnet_ids

  tags = merge(local.base_tags, {
    Name = "${var.environment}-postgres-subnets"
  })
}

resource "aws_db_instance" "postgres" {
  count = var.create_postgres ? 1 : 0

  identifier              = "${var.environment}-postgres"
  engine                  = "postgres"
  engine_version          = var.postgres_engine_version
  instance_class          = var.postgres_instance_class
  allocated_storage       = 50
  storage_encrypted       = true
  username                = var.database_username
  password                = var.database_password
  db_name                 = var.database_name
  multi_az                = false
  publicly_accessible     = false
  backup_retention_period = 7
  deletion_protection     = false
  skip_final_snapshot     = true

  db_subnet_group_name = aws_db_subnet_group.postgres[0].name
  vpc_security_group_ids = var.create_postgres ? [
    aws_security_group.postgres[0].id
  ] : []

  tags = merge(local.base_tags, {
    Name = "${var.environment}-postgres"
  })
}

# -----------------------------------------------------------------------------
# Redis (Amazon ElastiCache)
# -----------------------------------------------------------------------------

resource "aws_security_group" "redis" {
  count       = var.create_redis ? 1 : 0
  name        = "${var.environment}-redis-sg"
  description = "Redis access rules"
  vpc_id      = var.vpc_id

  dynamic "ingress" {
    for_each = var.allowed_cidr_blocks
    content {
      description = "Redis ingress"
      from_port   = 6379
      to_port     = 6379
      protocol    = "tcp"
      cidr_blocks = [ingress.value]
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.base_tags, {
    Name = "${var.environment}-redis-sg"
  })
}

resource "aws_elasticache_subnet_group" "redis" {
  count      = var.create_redis ? 1 : 0
  name       = "${var.environment}-redis-subnets"
  subnet_ids = var.private_subnet_ids

  tags = merge(local.base_tags, {
    Name = "${var.environment}-redis-subnets"
  })
}

resource "aws_elasticache_replication_group" "redis" {
  count = var.create_redis ? 1 : 0

  replication_group_id          = "${var.environment}-redis"
  replication_group_description = "Redis for network-sec microservices"
  engine                        = "redis"
  engine_version                = "7.0"
  node_type                     = var.redis_node_type
  parameter_group_name          = "default.redis7"
  port                          = 6379
  automatic_failover_enabled    = true
  multi_az_enabled              = true
  number_cache_clusters         = var.redis_num_cache_clusters
  subnet_group_name             = aws_elasticache_subnet_group.redis[0].name
  security_group_ids            = [aws_security_group.redis[0].id]
  at_rest_encryption_enabled    = true
  transit_encryption_enabled    = true

  tags = merge(local.base_tags, {
    Name = "${var.environment}-redis"
  })
}

# -----------------------------------------------------------------------------
# Mongo-compatible (Amazon DocumentDB)
# -----------------------------------------------------------------------------

resource "aws_security_group" "mongo" {
  count       = var.create_mongo ? 1 : 0
  name        = "${var.environment}-docdb-sg"
  description = "DocumentDB access rules"
  vpc_id      = var.vpc_id

  dynamic "ingress" {
    for_each = var.allowed_cidr_blocks
    content {
      description = "Mongo ingress"
      from_port   = 27017
      to_port     = 27017
      protocol    = "tcp"
      cidr_blocks = [ingress.value]
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.base_tags, {
    Name = "${var.environment}-docdb-sg"
  })
}

resource "aws_docdb_subnet_group" "mongo" {
  count      = var.create_mongo ? 1 : 0
  name       = "${var.environment}-docdb-subnets"
  subnet_ids = var.private_subnet_ids

  tags = merge(local.base_tags, {
    Name = "${var.environment}-docdb-subnets"
  })
}

resource "aws_docdb_cluster" "mongo" {
  count               = var.create_mongo ? 1 : 0
  cluster_identifier  = "${var.environment}-docdb"
  master_username     = var.mongo_username
  master_password     = var.mongo_password
  engine_version      = "5.0.0"
  db_subnet_group_name = aws_docdb_subnet_group.mongo[0].name
  vpc_security_group_ids = [aws_security_group.mongo[0].id]
  storage_encrypted   = true
  backup_retention_period = 7
  deletion_protection = false
  skip_final_snapshot = true
  preferred_backup_window = "04:00-05:00"

  tags = merge(local.base_tags, {
    Name = "${var.environment}-docdb"
  })
}

resource "aws_docdb_cluster_instance" "mongo" {
  count              = var.create_mongo ? 2 : 0
  identifier         = "${var.environment}-docdb-${count.index}"
  cluster_identifier = aws_docdb_cluster.mongo[0].id
  instance_class     = var.mongo_instance_class
  apply_immediately  = true

  tags = merge(local.base_tags, {
    Name = "${var.environment}-docdb-${count.index}"
  })
}

# -----------------------------------------------------------------------------
# Kafka (Amazon MSK)
# -----------------------------------------------------------------------------

resource "aws_security_group" "kafka" {
  count       = var.create_kafka ? 1 : 0
  name        = "${var.environment}-msk-sg"
  description = "MSK access rules"
  vpc_id      = var.vpc_id

  dynamic "ingress" {
    for_each = var.allowed_cidr_blocks
    content {
      description = "Kafka client ingress"
      from_port   = 9092
      to_port     = 9094
      protocol    = "tcp"
      cidr_blocks = [ingress.value]
    }
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = merge(local.base_tags, {
    Name = "${var.environment}-msk-sg"
  })
}

resource "aws_msk_configuration" "this" {
  count          = var.create_kafka ? 1 : 0
  name           = "${var.environment}-msk-config"
  kafka_versions = [var.kafka_version]
  server_properties = <<EOT
auto.create.topics.enable = true
delete.topic.enable = true
compression.type = producer
log.retention.hours = 168
EOT

  description = "MSK configuration for network-sec microservices"
}

resource "aws_msk_cluster" "this" {
  count                     = var.create_kafka ? 1 : 0
  cluster_name              = "${var.environment}-msk"
  kafka_version             = var.kafka_version
  number_of_broker_nodes    = var.kafka_number_of_broker_nodes
  configuration_info {
    arn      = aws_msk_configuration.this[0].arn
    revision = aws_msk_configuration.this[0].latest_revision
  }

  broker_node_group_info {
    instance_type   = "kafka.m5.large"
    ebs_volume_size = 100
    client_subnets  = var.private_subnet_ids
    security_groups = [aws_security_group.kafka[0].id]
  }

  client_authentication {
    sasl {
      scram = true
    }
    tls {
      enabled = true
    }
  }

  encryption_info {
    encryption_in_transit {
      client_broker = "TLS"
      in_cluster    = true
    }
  }

  enhanced_monitoring = "PER_TOPIC_PER_PARTITION"

  tags = merge(local.base_tags, {
    Name = "${var.environment}-msk"
  })
}

# -----------------------------------------------------------------------------
# Outputs
# -----------------------------------------------------------------------------

output "postgres_endpoint" {
  description = "DNS endpoint for PostgreSQL"
  value       = try(aws_db_instance.postgres[0].address, null)
}

output "redis_primary_endpoint" {
  description = "Redis primary endpoint"
  value       = try(aws_elasticache_replication_group.redis[0].primary_endpoint_address, null)
}

output "mongo_endpoint" {
  description = "DocumentDB endpoint"
  value       = try(aws_docdb_cluster.mongo[0].endpoint, null)
}

output "kafka_bootstrap_brokers" {
  description = "Bootstrap brokers string for MSK"
  value       = try(aws_msk_cluster.this[0].bootstrap_brokers_tls, null)
}

