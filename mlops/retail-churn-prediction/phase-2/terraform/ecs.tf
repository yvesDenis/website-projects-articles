# ── Application Load Balancer ─────────────────────────────────────────────────

resource "aws_lb" "main" {
  name               = "${var.project_name}-alb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.alb.id]
  subnets            = aws_subnet.public[*].id
}

resource "aws_lb_target_group" "churn_api" {
  name        = "${var.project_name}-tg"
  port        = 8000
  protocol    = "HTTP"
  vpc_id      = aws_vpc.main.id
  target_type = "ip"

  health_check {
    path                = "/health"
    interval            = 30
    timeout             = 5
    healthy_threshold   = 2
    unhealthy_threshold = 3
    matcher             = "200"
  }
}

resource "aws_lb_listener" "http" {
  load_balancer_arn = aws_lb.main.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.churn_api.arn
  }
}

# ── ECS Cluster ───────────────────────────────────────────────────────────────

resource "aws_ecs_cluster" "main" {
  name = var.ecs_cluster_name

  setting {
    name  = "containerInsights"
    value = "enabled"
  }
}

# ── Task Definition ───────────────────────────────────────────────────────────
# This Terraform-managed task definition is the initial version.
# Once CI/CD is running, GitHub Actions registers new revisions via:
#   aws ecs register-task-definition + aws ecs update-service
# The `ignore_changes = [task_definition]` lifecycle block below ensures
# Terraform stops managing the revision pointer after the first apply.

resource "aws_ecs_task_definition" "churn_api" {
  family                   = "churn-api"
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.ecs_execution.arn
  task_role_arn            = aws_iam_role.ecs_task.arn

  container_definitions = jsonencode([{
    name = "churn-api"

    # Placeholder image used on first apply before CI has pushed a real image.
    # CI will push to ECR and deploy a new task definition revision.
    image = "${aws_ecr_repository.churn_api.repository_url}:latest"

    portMappings = [{
      containerPort = 8000
      protocol      = "tcp"
    }]

    environment = [
      { name = "MODELS_DIR",          value = "/app/models" },
      { name = "AWS_DEFAULT_REGION",  value = var.aws_region },
      { name = "DRIFT_SNS_TOPIC_ARN", value = aws_sns_topic.drift_alerts.arn },
    ]

    secrets = [
      {
        name      = "DATABASE_URL"
        valueFrom = aws_ssm_parameter.db_url.arn
      }
    ]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.churn_api.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "ecs"
      }
    }

    healthCheck = {
      command     = ["CMD-SHELL", "curl -f http://localhost:8000/health || exit 1"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 15
    }
  }])
}

# ── ECS Service ───────────────────────────────────────────────────────────────

resource "aws_ecs_service" "churn_api" {
  name            = var.ecs_service_name
  cluster         = aws_ecs_cluster.main.id
  task_definition = aws_ecs_task_definition.churn_api.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets         = aws_subnet.public[*].id
    security_groups = [aws_security_group.ecs_tasks.id]

    # assign_public_ip = true because tasks live in public subnets (no NAT).
    # The ECS task still needs a public IP to pull images from ECR and reach S3.
    # Inbound traffic is locked to the ALB security group only.
    assign_public_ip = true
  }

  load_balancer {
    target_group_arn = aws_lb_target_group.churn_api.arn
    container_name   = "churn-api"
    container_port   = 8000
  }

  depends_on = [aws_lb_listener.http]

  lifecycle {
    # GitHub Actions owns the running task definition revision;
    # prevent Terraform from rolling it back on the next `terraform apply`.
    ignore_changes = [task_definition]
  }
}
