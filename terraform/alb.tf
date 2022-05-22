################################
# Locals
################################
locals {
  ecs_alb_name              = format("%s-%s-alb", var.app_name, var.env)
  ecs_alb_sg_name           = format("%s-%s-alb-sg", var.app_name, var.env)
  ecs_alb_listen_https_port = var.alb_listen_port
}

locals {
  ecs_alb_tg_name     = format("%s-%s-tg", var.app_name, var.env)
  ecs_alb_tg_port     = var.task_service_port
  ecs_alb_tg_protocol = "HTTP"
}

################################
# Subnets
################################
data "aws_subnets" "public" {
  filter {
    name   = "vpc-id"
    values = [var.vpc_id]
  }

  tags = {
    Purpose = "public-subnet"
  }
}

data "aws_subnets" "private" {
  filter {
    name   = "vpc-id"
    values = [var.vpc_id]
  }

  tags = {
    Purpose = "private-subnet"
  }
}

################################
# Security Groups
################################
data "aws_security_group" "default" {
  vpc_id = var.vpc_id
  name   = "default"
}

resource "aws_security_group" "alb_sg" {
  name        = local.ecs_alb_sg_name
  description = "Control access to the ALB"
  vpc_id      = var.vpc_id

  ingress {
    protocol  = "tcp"
    from_port = var.alb_listen_port
    to_port   = var.alb_listen_port

    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name        = local.ecs_alb_sg_name
    Environment = var.env
  }
}

################################
# Application Load Balancer
################################
resource "aws_alb" "alb" {
  name            = local.ecs_alb_name
  subnets         = data.aws_subnets.public.ids
  security_groups = [aws_security_group.alb_sg.id, data.aws_security_group.default.id]

  tags = {
    Name        = local.ecs_alb_name
    Environment = var.env
  }
}

resource "random_id" "alb_tg_random" {
  keepers = {
    name     = local.ecs_alb_tg_name
    port     = local.ecs_alb_tg_port
    protocol = local.ecs_alb_tg_protocol
    vpc_id   = var.vpc_id
  }

  byte_length = 2
}

resource "aws_lb_target_group" "alb_tg" {
  name                 = "${local.ecs_alb_tg_name}-${random_id.alb_tg_random.hex}"
  port                 = local.ecs_alb_tg_port
  protocol             = local.ecs_alb_tg_protocol
  vpc_id               = var.vpc_id
  deregistration_delay = 60
  target_type          = "ip"

  health_check {
    healthy_threshold   = 3
    unhealthy_threshold = 3
    interval            = 60
    protocol            = local.ecs_alb_tg_protocol
    timeout             = 5
    matcher             = 200
    path                = "/api/v1/health"
  }

  # To achieve zero downtime when we update this target group
  lifecycle {
    create_before_destroy = true
  }

  # Avoid error: The target group does not have an associated load balancer.
  depends_on = [aws_alb.alb]

  tags = {
    Name        = local.ecs_alb_tg_name
    Environment = var.env
  }
}

data "aws_acm_certificate" "cert" {
  domain = "*.david74.dev"
}

resource "aws_alb_listener" "https" {
  load_balancer_arn = aws_alb.alb.id
  port              = local.ecs_alb_listen_https_port
  protocol          = "HTTPS"
  certificate_arn   = data.aws_acm_certificate.cert.arn

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.alb_tg.id
  }
}