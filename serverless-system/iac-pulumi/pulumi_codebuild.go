package main

import (
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/codebuild"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/ecr"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

type config_data struct {
	awsRegion string
	accountId string
}

func loadConfig(ctx *pulumi.Context) *config_data {
	config := config.New(ctx, "serverless")
	return &config_data{
		awsRegion: config.Get("region"),
		accountId: config.Get("account"),
	}
}

var (
	ecr_repo_name    = []string{"createorders", "getorders", "deleteorder"}
	ecr_resource_map [3]*ecr.Repository
)

// Codebuild project
func createInfrastructureCodebuild(ctx *pulumi.Context) (*codebuild.Project, error) {

	configData := loadConfig(ctx)

	for key, repoName := range ecr_repo_name {
		repository, err := createInfrastructureECR(ctx, repoName)
		if err != nil {
			return nil, err
		}
		ecr_resource_map[key] = repository
	}

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
			return `{
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
					"Resource": "*"
					},
					{
					"Effect": "Allow",
					"Action": [
						"kms:*"
					],
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
				}`, nil
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
			PrivilegedMode:           pulumi.Bool(true),
			EnvironmentVariables: codebuild.ProjectEnvironmentEnvironmentVariableArray{
				&codebuild.ProjectEnvironmentEnvironmentVariableArgs{
					Name:  pulumi.String("AWS_DEFAULT_REGION"),
					Value: pulumi.String(configData.awsRegion),
				},
				&codebuild.ProjectEnvironmentEnvironmentVariableArgs{
					Name:  pulumi.String("AWS_ACCOUNT_ID"),
					Value: pulumi.String(configData.accountId),
				},
				&codebuild.ProjectEnvironmentEnvironmentVariableArgs{
					Name:  pulumi.String("IMAGE_REPO_CREATE"),
					Value: ecr_resource_map[0].Name,
				},
				&codebuild.ProjectEnvironmentEnvironmentVariableArgs{
					Name:  pulumi.String("IMAGE_REPO_GET"),
					Value: ecr_resource_map[1].Name,
				},
				&codebuild.ProjectEnvironmentEnvironmentVariableArgs{
					Name:  pulumi.String("IMAGE_REPO_DELETE"),
					Value: ecr_resource_map[2].Name,
				},
			},
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
			GitCloneDepth: pulumi.Int(5),
			Buildspec:     pulumi.String("serverless-system/buildspec.yml"),
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

	return serverlesscodebuildProject, nil
}
