################################
# IAM
################################
resource "aws_iam_role" "service_task_execution_role" {
  name = "david74-${var.app_name}-${var.region}-${var.env}-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ecs-tasks.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF

  tags = {
    Name        = "${var.app_name}-${var.region}-${var.env}-role"
    Environment = var.env
  }
}

resource "aws_iam_role_policy_attachment" "service_task_execution_policy" {
  role       = aws_iam_role.service_task_execution_role.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

###############################
# Attach service inline policy
###############################
data "aws_caller_identity" "current" {
}

data "aws_iam_policy_document" "inline_policy" {
  statement {
    effect  = "Allow"
    actions = [
      "ssm:GetParameters",
    ]
    resources = [
      "arn:aws:ssm:${var.region}:${data.aws_caller_identity.current.account_id}:parameter/${var.app_name}-*",
    ]
  }

  statement {
    effect  = "Allow"
    actions = [
      "events:PutEvents",
    ]
    resources = [
      aws_cloudwatch_event_bus.message_bus.arn,
    ]
  }

  statement {
    effect  = "Allow"
    actions = [
      "ssmmessages:CreateControlChannel",
      "ssmmessages:CreateDataChannel",
      "ssmmessages:OpenControlChannel",
      "ssmmessages:OpenDataChannel"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role_policy" "inline_policy" {
  name   = "inline-policy"
  role   = aws_iam_role.service_task_execution_role.name
  policy = data.aws_iam_policy_document.inline_policy.json
}
