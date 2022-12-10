resource "aws_ecr_repository" "base-app-repository" {
  name                 = "base-app-repo"
  image_tag_mutability = "IMMUTABLE"

  tags = {
    name = "base-app"
    subject = "github"
  }
}

resource "aws_ecr_repository_policy" "base-app-repo-policy" {
  repository = aws_ecr_repository.base-app-repository.name
  policy     = jsonencode(
    {
        "Version": "2008-10-17",
        "Statement": [
            {
                "Sid": "adds full ecr access to the demo repository",
                "Effect": "Allow",
                "Principal": "*",
                "Action": [
                "ecr:BatchCheckLayerAvailability",
                "ecr:BatchGetImage",
                "ecr:CompleteLayerUpload",
                "ecr:GetDownloadUrlForLayer",
                "ecr:GetLifecyclePolicy",
                "ecr:InitiateLayerUpload",
                "ecr:PutImage",
                "ecr:UploadLayerPart"
                ]
            }
        ]
    }
  )
}