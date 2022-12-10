locals {
  vpc_id = "vpc-09c827d20524ded81"
}


resource "aws_ecs_cluster" "base_app_aws_ecs_cluster" {
  name = "base-app-cluster"

  tags = {
    name = "base-app"
    subject = "github"
  }
}

resource "aws_ecs_cluster_capacity_providers" "base_app_aws_ecs_cluster_capacity_providers" {
  cluster_name = aws_ecs_cluster.base_app_aws_ecs_cluster.name

  capacity_providers = [aws_ecs_capacity_provider.base_app_aws_ecs_capacity_provider.name]
}

resource "aws_ecs_capacity_provider" "base_app_aws_ecs_capacity_provider" {
  name = "base-app-capacity-provider"

  auto_scaling_group_provider {
    auto_scaling_group_arn         = aws_autoscaling_group.base_app_aws_autoscaling_group.arn
    managed_termination_protection = "ENABLED"

    managed_scaling {
      maximum_scaling_step_size = 3
      minimum_scaling_step_size = 2
      status                    = "ENABLED"
      target_capacity           = 60
    }
  }
}

data "aws_launch_template" "base_app_aws_launch_template" {
  name = "demo-template"
}

resource "aws_autoscaling_group" "base_app_aws_autoscaling_group" {
  vpc_zone_identifier = data.aws_subnets.base_app_aws_subnet_ids.ids
  desired_capacity   = 3
  max_size           = 3
  min_size           = 2
  health_check_grace_period = 300
  health_check_type         = "ELB"
  protect_from_scale_in  = true

  launch_template {
    id      = data.aws_launch_template.base_app_aws_launch_template.id
    version = "$Latest"
  }
}

data "aws_security_groups" "base_app_aws_security_groups" {
  filter {
    name   = "group-name"
    values = ["myFirstALB"]
  }
}

data "aws_subnets" "base_app_aws_subnet_ids" {
  filter {
    name   = "tag:use"
    values = ["auto-scaling"]
  }
}

resource "aws_lb" "base_app_aws_lb" {
  name               = "base-app-alb"
  internal           = false
  security_groups    = data.aws_security_groups.base_app_aws_security_groups.ids
  subnets            = data.aws_subnets.base_app_aws_subnet_ids.ids

  enable_deletion_protection = true

  tags = {
    name = "base-app"
    subject = "github"
  }
}

resource "aws_lb_listener" "base_app_aws_lb_listener" {
  load_balancer_arn = aws_lb.base_app_aws_lb.arn
  port              = "80"
  protocol          = "HTTP"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.base_app_aws_lb_target_group.arn
  }
}

resource "aws_lb_target_group" "base_app_aws_lb_target_group" {
   name     = "base-app-lb-target-group"
   port     = 8080
   protocol = "HTTP"
   vpc_id   = local.vpc_id

   health_check {
    interval   = 200
    path       = "/hello"
   }
 }

resource "aws_autoscaling_attachment" "base_app_aws_autoscaling_attachment" {
  autoscaling_group_name = aws_autoscaling_group.base_app_aws_autoscaling_group.id
  alb_target_group_arn   = aws_lb_target_group.base_app_aws_lb_target_group.arn
}

resource "aws_ecs_task_definition" "base_app_aws_ecs_task_definition" {
  family = "base-app-task-definition"
  container_definitions = jsonencode([
    {
      name      = "base-app-container"
      image     = "516193157210.dkr.ecr.ca-central-1.amazonaws.com/base-app-repo:9fd96f8d919ff33e07a01dd2c6da8e9d69b73f7f"
      cpu       = 300
      memory    = 400
      essential = true
      portMappings = [
        {
          containerPort = 8080
          hostPort      = 0
        }
      ]
    }
  ])
  execution_role_arn = "arn:aws:iam::516193157210:role/ecsTaskExecutionRole"
  network_mode = "bridge"
  runtime_platform {
    cpu_architecture = "X86_64"
    operating_system_family = "LINUX"
  }
  requires_compatibilities = ["EC2"]

  tags = {
    name = "base-app"
    subject = "github"
  }
}

resource "aws_ecs_service" "base_app_aws_ecs_service" {
  name            = "base-app-ecs-service"
  cluster         = aws_ecs_cluster.base_app_aws_ecs_cluster.id
  task_definition = aws_ecs_task_definition.base_app_aws_ecs_task_definition.arn
  desired_count   = 1
  iam_role        = "arn:aws:iam::516193157210:role/ecsInstanceRole"

  load_balancer {
    target_group_arn = aws_lb_target_group.base_app_aws_lb_target_group.arn
    container_name   = "base-app-container"
    container_port   = 8080
  }

  tags = {
    name = "base-app"
    subject = "github"
  }
}