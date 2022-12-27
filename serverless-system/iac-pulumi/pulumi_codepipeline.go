package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/codebuild"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/codepipeline"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/codestarconnections"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/kms"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type result_infra struct {
	codebuild    *codebuild.Project
	codepipeline *codepipeline.Pipeline
}

// Codepipeline project
func createInfrastructureCodepipeline(ctx *pulumi.Context) (*result_infra, error) {

	serverlesscodebuildProject, err := createInfrastructureCodebuild(ctx)
	if err != nil {
		return nil, err
	}

	serverlessSystemConnection, err := codestarconnections.NewConnection(ctx, "serverlesssystemconnection", &codestarconnections.ConnectionArgs{
		Name:         pulumi.String("serverlesssystemconnection"),
		ProviderType: pulumi.String("GitHub"),
	})
	if err != nil {
		return nil, err
	}

	codepipelineBucket, err := s3.NewBucketV2(ctx, "serverlessCodepipelineBucket", nil)
	if err != nil {
		return nil, err
	}
	codepipelineRole, err := iam.NewRole(ctx, "serverlessCodepipelineRole", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.Any(`{
			"Version": "2012-10-17",
			"Statement": [
				{
				"Effect": "Allow",
				"Principal": {
					"Service": "codepipeline.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
				}
			]
			}
			`),
	})

	if err != nil {
		return nil, err
	}
	cloudformationRole, err := iam.NewRole(ctx, "cloudformationRole", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.Any(`{
			"Version": "2012-10-17",
			"Statement": [
				{
				"Effect": "Allow",
				"Principal": {
					"Service": "cloudformation.amazonaws.com"
				},
				"Action": "sts:AssumeRole"
				}
			]
			}
			`),
	})
	if err != nil {
		return nil, err
	}
	s3kmskey, err := kms.NewKey(ctx, "serverlesskey", &kms.KeyArgs{
		DeletionWindowInDays: pulumi.Int(10),
		Description:          pulumi.String("KMS key for serverless codepipeline"),
	})

	if err != nil {
		return nil, err
	}
	codepipeline, err := codepipeline.NewPipeline(ctx, "serverlesscodepipeline", &codepipeline.PipelineArgs{
		RoleArn: codepipelineRole.Arn,
		Name:    pulumi.String("serverlesscodepipeline"),
		ArtifactStores: codepipeline.PipelineArtifactStoreArray{
			&codepipeline.PipelineArtifactStoreArgs{
				Location: codepipelineBucket.Bucket,
				Type:     pulumi.String("S3"),
				EncryptionKey: &codepipeline.PipelineArtifactStoreEncryptionKeyArgs{
					Id:   s3kmskey.Arn,
					Type: pulumi.String("KMS"),
				},
			},
		},
		Stages: codepipeline.PipelineStageArray{
			&codepipeline.PipelineStageArgs{
				Name: pulumi.String("Source"),
				Actions: codepipeline.PipelineStageActionArray{
					&codepipeline.PipelineStageActionArgs{
						Name:     pulumi.String("Source"),
						Category: pulumi.String("Source"),
						Owner:    pulumi.String("AWS"),
						Provider: pulumi.String("CodeStarSourceConnection"),
						Version:  pulumi.String("1"),
						OutputArtifacts: pulumi.StringArray{
							pulumi.String("source_output"),
						},
						Configuration: pulumi.StringMap{
							"ConnectionArn":    serverlessSystemConnection.Arn,
							"FullRepositoryId": pulumi.String("yvesDenis/website-projects-articles"),
							"BranchName":       pulumi.String("serverless-system"),
						},
					},
				},
			},
			&codepipeline.PipelineStageArgs{
				Name: pulumi.String("Build"),
				Actions: codepipeline.PipelineStageActionArray{
					&codepipeline.PipelineStageActionArgs{
						Name:     pulumi.String("Build"),
						Category: pulumi.String("Build"),
						Owner:    pulumi.String("AWS"),
						Provider: pulumi.String("CodeBuild"),
						InputArtifacts: pulumi.StringArray{
							pulumi.String("source_output"),
						},
						OutputArtifacts: pulumi.StringArray{
							pulumi.String("build_output"),
						},
						Version: pulumi.String("1"),
						Configuration: pulumi.StringMap{
							"ProjectName": pulumi.String("serverlesscodebuildProject"),
						},
					},
				},
			},
			&codepipeline.PipelineStageArgs{
				Name: pulumi.String("Deploy"),
				Actions: codepipeline.PipelineStageActionArray{
					&codepipeline.PipelineStageActionArgs{
						Name:     pulumi.String("Deploy"),
						Category: pulumi.String("Deploy"),
						Owner:    pulumi.String("AWS"),
						Provider: pulumi.String("CloudFormation"),
						InputArtifacts: pulumi.StringArray{
							pulumi.String("build_output"),
						},
						Version: pulumi.String("1"),
						Configuration: pulumi.StringMap{
							"ActionMode":            pulumi.String("REPLACE_ON_FAILURE"),
							"RoleArn":               cloudformationRole.Arn,
							"Capabilities":          pulumi.String("CAPABILITY_AUTO_EXPAND,CAPABILITY_IAM"),
							"OutputFileName":        pulumi.String("CreateStackOutput.json"),
							"StackName":             pulumi.String("serverlessSystemStack"),
							"TemplatePath":          pulumi.String("build_output::template-serverless.yaml"),
							"TemplateConfiguration": pulumi.String("build_output::parameter-configuration.json"),
						},
					},
				},
			},
		},
		Tags: pulumi.StringMap{
			"Environment": pulumi.String("prod"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{serverlesscodebuildProject}))
	if err != nil {
		return nil, err
	}
	_, err = s3.NewBucketAclV2(ctx, "serverlessCodepipelineBucketAcl", &s3.BucketAclV2Args{
		Bucket: codepipelineBucket.ID(),
		Acl:    pulumi.String("private"),
	})
	if err != nil {
		return nil, err
	}

	_, err = iam.NewRolePolicy(ctx, "cloudformationRolePolicy", &iam.RolePolicyArgs{
		Role: cloudformationRole.ID(),
		Policy: pulumi.All(codepipelineBucket.Arn, codepipelineBucket.Arn, serverlessSystemConnection.Arn).ApplyT(func(_args []interface{}) (string, error) {
			return `{
				"Version": "2012-10-17",
				"Statement": [
					{
					"Effect":"Allow",
					"Action": [
						"s3:GetObject",
						"s3:GetObjectVersion",
						"s3:GetBucketVersioning",
						"s3:PutObjectAcl",
						"s3:PutObject"
					],
					"Resource": "*"
					},
					{
					"Effect": "Allow",
					"Action": [
						"lambda:*"
					],
					"Resource": "*"
					},
					{
					"Effect": "Allow",
					"Action": [
						"dynamodb:*"
					],
					"Resource": "*"
					},
					{
					"Effect": "Allow",
					"Action": "kms:*",
					"Resource": "*"
					},
					{
					"Effect": "Allow",
					"Action": [
						"ecr:*"
					],
					"Resource": "*"
					},
					{
					"Effect": "Allow",
					"Action": [
						"apigateway:*"
					],
					"Resource": "*"
					}
				]
				}
				`, nil
		}).(pulumi.StringOutput),
	})

	if err != nil {
		return nil, err
	}

	_, err = iam.NewRolePolicy(ctx, "serverlessCodepipelinePolicy", &iam.RolePolicyArgs{
		Role: codepipelineRole.ID(),
		Policy: pulumi.All(codepipelineBucket.Arn, codepipelineBucket.Arn, serverlessSystemConnection.Arn).ApplyT(func(_args []interface{}) (string, error) {
			codepipelineBucketArn := _args[0].(string)
			codepipelineBucketArn1 := _args[1].(string)
			codepipelineArn := _args[2].(string)
			return fmt.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
					"Effect":"Allow",
					"Action": [
						"s3:GetObject",
						"s3:GetObjectVersion",
						"s3:GetBucketVersioning",
						"s3:PutObjectAcl",
						"s3:PutObject"
					],
					"Resource": [
						"%v",
						"%v/*"
					]
					},
					{
					"Effect": "Allow",
					"Action": [
						"codestar-connections:UseConnection"
					],
					"Resource": "%v"
					},
					{
					"Effect": "Allow",
					"Action": [
						"codebuild:BatchGetBuilds",
						"codebuild:StartBuild"
					],
					"Resource": "*"
					},
					{
					"Effect": "Allow",
					"Action": "kms:*",
					"Resource": "*"
					},
					{
					"Effect": "Allow",
					"Action": [
						"ecr:*"
					],
					"Resource": "*"
					}
				]
				}
				`, codepipelineBucketArn, codepipelineBucketArn1, codepipelineArn), nil
		}).(pulumi.StringOutput),
	})
	if err != nil {
		return nil, err
	}
	return &result_infra{
		codebuild:    serverlesscodebuildProject,
		codepipeline: codepipeline,
	}, nil
}
