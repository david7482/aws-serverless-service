variable "aws_profile" {
  default = "david74"
}

variable "region" {
  default = "us-west-2"
}

variable "env" {
  default = "production"
}

variable "vpc_id" {
  default = "vpc-07808722ae95b965d"
}

variable "app_name" {
  default = "chatbot"
}

variable "domain" {
  default = "david74.dev"
}

variable "alb_listen_port" {
  default = 443
}

variable "task_service_port" {
  default = 8000
}

variable "ecr_repository_url" {
  default = "553321195691.dkr.ecr.us-west-2.amazonaws.com/chatbot-serverless-service"
}

variable "task_cpu" {
  default = 256
}

variable "task_memory" {
  default = 512
}

variable "ecs_cluster_name" {
  default = "ecs-playground"
}

variable "desired_task_count" {
  default = 1
}