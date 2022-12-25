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

func createInfrastructure(ctx *pulumi.Context) (*result_infra, error) {

	// Codebuild project
	serverlessCodebuildBucketV2, err := s3.NewBucketV2(ctx, "serverless-codebuild-bucket-v2", nil)
	if err != nil {
		return nil, err
	}
	_, err = s3.NewBucketAclV2(ctx, "serverlessCodebuildBucketAclV2", &s3.BucketAclV2Args{
		Bucket: serverlessCodebuildBucketV2.ID(),
		Acl:    pulumi.String("private"),
	})
	if err != nil {
		return nil, err
	}
	serverlesscodebuildrole, err := iam.NewRole(ctx, "serverlesscodebuildrole", &iam.RoleArgs{
		AssumeRolePolicy: pulumi.Any(`{
			"Version": "2012-10-17",
			"Statement": [
				{
				"Effect": "Allow",
				"Principal": {
					"Service": "codebuild.amazonaws.com"
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
	_, err = iam.NewRolePolicy(ctx, "serverlessCodebuildRolePolicy", &iam.RolePolicyArgs{
		Role: serverlesscodebuildrole.Name,
		Policy: pulumi.All(serverlessCodebuildBucketV2.Arn, serverlessCodebuildBucketV2.Arn).ApplyT(func(_args []interface{}) (string, error) {
			serverlessCodebuildBucketV2Arn := _args[0].(string)
			serverlessCodebuildBucketV2Arn1 := _args[1].(string)
			return fmt.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
					"Effect": "Allow",
					"Resource": [
						"*"
					],
					"Action": [
						"logs:CreateLogGroup",
						"logs:CreateLogStream",
						"logs:PutLogEvents"
					]
					},
					{
					"Effect": "Allow",
					"Action": [
						"ec2:CreateNetworkInterface",
						"ec2:DescribeDhcpOptions",
						"ec2:DescribeNetworkInterfaces",
						"ec2:DeleteNetworkInterface",
						"ec2:DescribeSubnets",
						"ec2:DescribeSecurityGroups",
						"ec2:DescribeVpcs"
					],
					"Resource": "*"
					},
					{
					"Effect": "Allow",
					"Action": [
						"ec2:CreateNetworkInterfacePermission"
					],
					"Resource": [
						"arn:aws:ec2:us-east-1:123456789012:network-interface/*"
					]
					},
					{
					"Effect": "Allow",
					"Action": [
						"s3:*"
					],
					"Resource": [
						"%v",
						"%v/*"
					]
					}
				]
				}`, serverlessCodebuildBucketV2Arn, serverlessCodebuildBucketV2Arn1), nil
		}).(pulumi.StringOutput),
	})
	if err != nil {
		return nil, err
	}
	serverlesscodebuildProject, err := codebuild.NewProject(ctx, "serverlesscodebuildProject", &codebuild.ProjectArgs{
		Description:  pulumi.String("serverless_codebuild_project"),
		Name:         pulumi.String("serverlesscodebuildProject"),
		BuildTimeout: pulumi.Int(5),
		ServiceRole:  serverlesscodebuildrole.Arn,
		Artifacts: &codebuild.ProjectArtifactsArgs{
			Type: pulumi.String("NO_ARTIFACTS"),
		},
		Cache: &codebuild.ProjectCacheArgs{
			Type:     pulumi.String("S3"),
			Location: serverlessCodebuildBucketV2.Bucket,
		},
		Environment: &codebuild.ProjectEnvironmentArgs{
			ComputeType:              pulumi.String("BUILD_GENERAL1_SMALL"),
			Image:                    pulumi.String("aws/codebuild/standard:6.0"),
			Type:                     pulumi.String("LINUX_CONTAINER"),
			ImagePullCredentialsType: pulumi.String("CODEBUILD"),
		},
		LogsConfig: &codebuild.ProjectLogsConfigArgs{
			CloudwatchLogs: &codebuild.ProjectLogsConfigCloudwatchLogsArgs{
				GroupName:  pulumi.String("serverless-codebuild-log-group"),
				StreamName: pulumi.String("serverless-codebuild-log-stream"),
			},
		},
		Source: &codebuild.ProjectSourceArgs{
			Type:          pulumi.String("GITHUB"),
			Location:      pulumi.String("https://github.com/yvesDenis/website-projects-articles.git"),
			GitCloneDepth: pulumi.Int(1),
			GitSubmodulesConfig: &codebuild.ProjectSourceGitSubmodulesConfigArgs{
				FetchSubmodules: pulumi.Bool(true),
			},
		},
		SourceVersion: pulumi.String("serverless-system"),
		Tags: pulumi.StringMap{
			"Environment": pulumi.String("prod"),
		},
	})
	if err != nil {
		return nil, err
	}

	// Codepipeline project
	serverlessSystemConnection, err := codestarconnections.NewConnection(ctx, "serverlesssystemconnection", &codestarconnections.ConnectionArgs{
		Name:         pulumi.string("serverlesssystemconnection"),
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
					Id:   pulumi.StringOutput(s3kmskey.Arn),
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
						Version:  pulumi.String("2"),
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
							"ActionMode":     pulumi.String("REPLACE_ON_FAILURE"),
							"Capabilities":   pulumi.String("CAPABILITY_AUTO_EXPAND,CAPABILITY_IAM"),
							"OutputFileName": pulumi.String("CreateStackOutput.json"),
							"StackName":      pulumi.String("MyStack"),
							"TemplatePath":   pulumi.String("build_output::sam-templated.yaml"),
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

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		infra, err := createInfrastructure(ctx)
		if err != nil {
			return err
		}

		ctx.Export("codebuildName", infra.codebuild.Name)
		ctx.Export("codebuildTags", infra.codebuild.Tags)
		ctx.Export("codepipelineName", infra.codepipeline.Name)
		ctx.Export("codepipelineTags", infra.codepipeline.Tags)

		return nil
	})
}
