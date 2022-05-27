locals {
  lambda_file        = "lambda.zip"
  lambda_binary_name = "chatbot-worker"
}

data "archive_file" "lambda" {
  type        = "zip"
  source_file = "../bin/${local.lambda_binary_name}"
  output_path = local.lambda_file
}

resource "aws_cloudwatch_log_group" "worker_log_group" {
  name              = "/aws/lambda/${var.app_name}-worker-${var.env}"
  retention_in_days = 7

  tags = {
    Name        = "${var.app_name}-worker-${var.env}"
    Environment = var.env
  }
}

resource "aws_iam_role" "chatbot_worker_role" {
  name = "david74-chatbot-worker-lambda-${var.region}-${var.env}-role"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow"
    }
  ]
}
EOF
}

data "aws_iam_policy_document" "chatbot_worker_role_inline_policy" {
  statement {
    effect  = "Allow"
    actions = [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["arn:aws:logs:*:*:*"]
  }
}

resource "aws_iam_role_policy" "post_camera_snapshot_lambda_inline_policy" {
  name   = "inline-policy"
  role   = aws_iam_role.chatbot_worker_role.name
  policy = data.aws_iam_policy_document.chatbot_worker_role_inline_policy.json
}

resource "aws_lambda_function" "worker" {
  function_name    = "${var.app_name}-worker-${var.env}"
  filename         = data.archive_file.lambda.output_path
  role             = aws_iam_role.chatbot_worker_role.arn
  handler          = local.lambda_binary_name
  runtime          = "go1.x"
  source_code_hash = filebase64sha256(data.archive_file.lambda.output_path)
  memory_size      = 128
  timeout          = 30
  publish          = "true"

  environment {
    variables = {
      DATABASE_DSN = data.aws_ssm_parameter.database_dsn.value
    }
  }

  tags = {
    Name        = "${var.app_name}-worker-${var.env}"
    Environment = var.env
  }
  depends_on = [aws_cloudwatch_log_group.log_group]
}

resource "aws_lambda_permission" "allow_cloudwatch_to_call_check_foo" {
  statement_id  = "AllowExecutionFromEventBridge"
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.worker.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.all.arn
}