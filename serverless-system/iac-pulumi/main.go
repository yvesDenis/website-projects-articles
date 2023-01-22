package main

import (
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		infra, err := createInfrastructureCodepipeline(ctx)
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
