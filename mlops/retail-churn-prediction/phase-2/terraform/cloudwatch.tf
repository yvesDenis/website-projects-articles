# ── Log Group ─────────────────────────────────────────────────────────────────

resource "aws_cloudwatch_log_group" "churn_api" {
  name              = "/ecs/churn-api"
  retention_in_days = 30
}

# ── Metric Filters ────────────────────────────────────────────────────────────
# The FastAPI middleware in src/api/main.py logs every request as structured
# JSON to stdout, e.g.:
#   {"event":"request","path":"/predict","status_code":200,"latency_ms":34.7}
#
# These filters parse that JSON and create CloudWatch metrics automatically —
# no SDK calls needed from the application code.

resource "aws_cloudwatch_log_metric_filter" "latency" {
  name           = "churn-api-latency"
  log_group_name = aws_cloudwatch_log_group.churn_api.name

  # Match any log event that contains a latency_ms field
  pattern = "{ $.latency_ms = * }"

  metric_transformation {
    name          = "RequestLatencyMs"
    namespace     = "ChurnAPI"
    value         = "$.latency_ms"
    default_value = "0"
    unit          = "Milliseconds"
  }
}

resource "aws_cloudwatch_log_metric_filter" "errors" {
  name           = "churn-api-errors"
  log_group_name = aws_cloudwatch_log_group.churn_api.name

  pattern = "{ $.status_code >= 500 }"

  metric_transformation {
    name          = "ErrorCount"
    namespace     = "ChurnAPI"
    value         = "1"
    default_value = "0"
    unit          = "Count"
  }
}

# ── Alarms ────────────────────────────────────────────────────────────────────

resource "aws_cloudwatch_metric_alarm" "p95_latency" {
  alarm_name          = "churn-api-p95-latency-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 3
  metric_name         = "RequestLatencyMs"
  namespace           = "ChurnAPI"
  period              = 60

  # p95 is only available via extended_statistic, not the basic `statistic` field.
  # CloudWatch computes it from the distribution of data points in the period.
  extended_statistic = "p95"

  threshold          = 200
  treat_missing_data = "notBreaching"
  alarm_description  = "p95 request latency exceeded 200ms for 3 consecutive minutes"
  alarm_actions      = [aws_sns_topic.alerts.arn]
  ok_actions         = [aws_sns_topic.alerts.arn]
}

resource "aws_cloudwatch_metric_alarm" "error_rate" {
  alarm_name          = "churn-api-error-rate-high"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 1
  metric_name         = "ErrorCount"
  namespace           = "ChurnAPI"
  period              = 300
  statistic           = "Sum"
  threshold           = 5
  treat_missing_data  = "notBreaching"
  alarm_description   = "More than 5 API errors in a 5-minute window"
  alarm_actions       = [aws_sns_topic.alerts.arn]
  ok_actions          = [aws_sns_topic.alerts.arn]
}

# ── Training ROC-AUC Alarm ────────────────────────────────────────────────────
# train.py publishes roc_auc to this namespace; if a run scores below 0.78
# the alarm fires before CI would have gated on it — useful for scheduled
# retraining jobs that run outside GitHub Actions.

resource "aws_cloudwatch_metric_alarm" "training_roc_auc" {
  alarm_name          = "churn-model-roc-auc-below-threshold"
  comparison_operator = "LessThanThreshold"
  evaluation_periods  = 1
  metric_name         = "RocAuc"
  namespace           = "ChurnModel"
  period              = 3600
  statistic           = "Maximum"
  threshold           = 0.78
  treat_missing_data  = "notBreaching"
  alarm_description   = "Latest training run scored ROC-AUC below 0.78"
  alarm_actions       = [aws_sns_topic.alerts.arn]
}
