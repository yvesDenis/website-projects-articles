# ── Alert Topics ──────────────────────────────────────────────────────────────

# General operational alerts: CloudWatch alarms (latency, errors, ROC-AUC)
resource "aws_sns_topic" "alerts" {
  name = "${var.project_name}-api-alerts"
}

# Drift alerts: published by drift_report.py when Evidently detects dataset drift
resource "aws_sns_topic" "drift_alerts" {
  name = "${var.project_name}-drift-alerts"
}

# ── Slack Subscriptions ───────────────────────────────────────────────────────
# Set var.slack_webhook_url to wire SNS → Slack.
# Leave it empty during initial Terraform apply; AWS will send a subscription
# confirmation request to the URL, which Slack's incoming webhook accepts
# automatically (it returns HTTP 200 to any POST).

resource "aws_sns_topic_subscription" "alerts_slack" {
  count     = var.slack_webhook_url != "" ? 1 : 0
  topic_arn = aws_sns_topic.alerts.arn
  protocol  = "https"
  endpoint  = var.slack_webhook_url

  # Slack does not support the SNS subscription confirmation URL pattern;
  # raw delivery sends the message body directly without the SNS envelope.
  raw_message_delivery = true
}

resource "aws_sns_topic_subscription" "drift_slack" {
  count     = var.slack_webhook_url != "" ? 1 : 0
  topic_arn = aws_sns_topic.drift_alerts.arn
  protocol  = "https"
  endpoint  = var.slack_webhook_url

  raw_message_delivery = true
}
