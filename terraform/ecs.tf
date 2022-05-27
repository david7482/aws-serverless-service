################################
# Locals
################################
locals {
  ecs_task_definition_name = "${var.app_name}-${var.env}"
  ecs_service_name         = "${var.app_name}-${var.env}-${random_id.ecs_service_random.hex}"
}

locals {
  task_container_name  = var.app_name
  task_container_image = "${var.ecr_repository_url}:${var.env}"
}

################################
# Cloudwatch Logs
################################

# Set up cloudwatch group and log stream and retain logs for 14 days
resource "aws_cloudwatch_log_group" "log_group" {
  name              = "/aws/ecs/${var.app_name}-${var.env}"
  retention_in_days = 14

  tags = {
    Name        = "${var.app_name}-${var.env}-log-group"
    Environment = var.env
  }
}

################################
# SSM Parameter Store
################################

data "aws_ssm_parameter" "database_dsn" {
  name = "chatbot-postgres-dsn-${var.env}"
}

################################
# ECS Task Definition
################################
module "container_definition" {
  source = "github.com/cloudposse/terraform-aws-ecs-container-definition?ref=0.58.1"

  container_name  = local.task_container_name
  container_image = local.task_container_image
  essential       = "true"

  log_configuration = {
    logDriver = "awslogs"
    options   = {
      awslogs-group         = aws_cloudwatch_log_group.log_group.name
      awslogs-region        = var.region
      awslogs-stream-prefix = "logs"
    },
    secretOptions = []
  }

  environment = [
    {
      name  = "AWS_EVENTBRIDGE_NAME"
      value = aws_cloudwatch_event_bus.message_bus.name
    },
  ]

  secrets = [
    {
      name      = "DATABASE_DSN"
      valueFrom = data.aws_ssm_parameter.database_dsn.arn
    },
  ]

  // These 3 fields are optional for Fargate. So we just set to 0 to overwrite the default values.
  #  container_cpu                = 0
  #  container_memory             = var.task_memory
  #  container_memory_reservation = null

  #  mount_points = []
  #  volumes_from = []
  #  user         = "0"

  port_mappings = [
    {
      containerPort = var.task_service_port
      protocol      = "tcp"
      hostPort      = var.task_service_port
    },
  ]

  // we only need "initProcessEnabled" so omit other fields
  linux_parameters = {
    capabilities       = null
    devices            = null
    maxSwap            = null
    sharedMemorySize   = null
    swappiness         = null
    tmpfs              = null
    initProcessEnabled = true
  }

  ulimits = [
    {
      name      = "nofile"
      hardLimit = 102400
      softLimit = 102400
    },
  ]
}

resource "aws_ecs_task_definition" "main" {
  family = local.ecs_task_definition_name

  execution_role_arn = aws_iam_role.service_task_execution_role.arn
  task_role_arn      = aws_iam_role.service_task_execution_role.arn

  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.task_cpu
  memory                   = var.task_memory
  container_definitions    = jsonencode([
    module.container_definition.json_map_object,
  ])
}

################################
# ECS
################################
data "aws_ecs_cluster" "main" {
  cluster_name = var.ecs_cluster_name
}

data "aws_ecs_task_definition" "main" {
  task_definition = aws_ecs_task_definition.main.family
}

resource "random_id" "ecs_service_random" {
  keepers = {
    version = 2
  }

  byte_length = 2
}

resource "aws_ecs_service" "main" {
  name                              = local.ecs_service_name
  cluster                           = data.aws_ecs_cluster.main.id
  desired_count                     = var.desired_task_count
  platform_version                  = "1.4.0"
  health_check_grace_period_seconds = 600

  # Always use the latest active version of the task definition
  task_definition = "${aws_ecs_task_definition.main.family}:${max(aws_ecs_task_definition.main.revision, data.aws_ecs_task_definition.main.revision)}"

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 100

  capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
  }

  network_configuration {
    security_groups  = [data.aws_security_group.default.id]
    subnets          = data.aws_subnets.private.ids
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.alb_tg.id
    container_name   = local.task_container_name
    container_port   = var.task_service_port
  }

  lifecycle {
    # Allow external changes without Terraform plan difference
    ignore_changes = [
      desired_count,
    ]
  }

  enable_execute_command  = true
  enable_ecs_managed_tags = true
  propagate_tags          = "SERVICE"

  tags = {
    Name        = "${var.app_name}-${var.env}"
    Environment = var.env
  }
}
