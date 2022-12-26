package main

import (
	"fmt"

	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/codebuild"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/iam"
	"github.com/pulumi/pulumi-aws/sdk/v5/go/aws/s3"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

// Codebuild project
func createInfrastructureCodebuild(ctx *pulumi.Context) (*codebuild.Project, error) {

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
			EnvironmentVariables: codebuild.ProjectEnvironmentEnvironmentVariableArray{
				&codebuild.ProjectEnvironmentEnvironmentVariableArgs{
					Name:  pulumi.String("CODEBUILD_SRC_DIR"),
					Value: pulumi.String("serverless-system/codebuild/buildspec.yml"),
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
			GitSubmodulesConfig: &codebuild.ProjectSourceGitSubmodulesConfigArgs{
				FetchSubmodules: pulumi.Bool(true),
			},
		},
		SourceVersion: pulumi.String("serverless-system-deploy"),
		Tags: pulumi.StringMap{
			"Environment": pulumi.String("prod"),
		},
	})
	if err != nil {
		return nil, err
	}

	return serverlesscodebuildProject, nil
}
