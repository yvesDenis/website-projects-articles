package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecr"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// ECR Repository project
func createInfrastructureECR(ctx *pulumi.Context) (*ecr.Repository, error) {

	serverlessrepository, err := ecr.NewRepository(ctx, "serverlessrepository", &ecr.RepositoryArgs{
		Name:               pulumi.String("serverlessrepository"),
		ImageTagMutability: pulumi.String("MUTABLE"),
		Tags: pulumi.StringMap{
			"Environment": pulumi.String("prod"),
		},
	})

	if err != nil {
		return nil, err
	}
	_, err = ecr.NewRepositoryPolicy(ctx, "serverlessrepositorypolicy", &ecr.RepositoryPolicyArgs{
		Repository: serverlessrepository.Name,
		Policy: pulumi.Any(fmt.Sprintf(`{
			"Version": "2008-10-17",
			"Statement": [
				{
					"Sid": "serverless repository policy",
					"Effect": "Allow",
					"Principal": "*",
					"Action": [
						"ecr:GetDownloadUrlForLayer",
						"ecr:BatchGetImage",
						"ecr:BatchCheckLayerAvailability",
						"ecr:PutImage",
						"ecr:InitiateLayerUpload",
						"ecr:UploadLayerPart",
						"ecr:CompleteLayerUpload",
						"ecr:DescribeRepositories",
						"ecr:GetRepositoryPolicy",
						"ecr:ListImages",
						"ecr:DeleteRepository",
						"ecr:BatchDeleteImage",
						"ecr:SetRepositoryPolicy",
						"ecr:DeleteRepositoryPolicy"
					]
				}
			]
			}`),
		),
	})

	if err != nil {
		return nil, err
	}

	return serverlessrepository, nil
}
