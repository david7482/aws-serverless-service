resource "aws_cloudwatch_event_bus" "message_bus" {
  name = "message-bus-${var.env}"

  tags = {
    Name        = "${var.app_name}-${var.env}"
    Environment = var.env
  }
}

resource "aws_cloudwatch_event_rule" "all" {
  name           = "message-bus-rule-all-${var.env}"
  event_bus_name = aws_cloudwatch_event_bus.message_bus.name
  is_enabled     = true
  event_pattern  = <<EOF
{
  "detail-type": [
    "line-message"
  ]
}
EOF
}

resource "aws_cloudwatch_event_target" "chatbot_worker" {
  target_id      = "chatbot_worker_lambda"
  event_bus_name = aws_cloudwatch_event_bus.message_bus.name
  rule           = aws_cloudwatch_event_rule.all.name
  arn            = aws_lambda_function.worker.arn
}

resource "aws_cloudwatch_event_archive" "archive" {
  name             = "message-bus-archive-${var.env}"
  event_source_arn = aws_cloudwatch_event_bus.message_bus.arn
  retention_days   = 7
}