package main

import (
	"sync"
	"testing"

	"github.com/pulumi/pulumi/sdk/v3/go/common/resource"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/stretchr/testify/assert"
)

type mocks int

// Create the mock.
func (mocks) NewResource(args pulumi.MockResourceArgs) (string, resource.PropertyMap, error) {
	outputs := args.Inputs.Mappable()
	if args.TypeToken == "aws:s3/bucket:BucketV2" {
		outputs["id"] = "mockIdResourceBucket"
	}
	return args.Name + "_id", resource.NewPropertyMapFromMap(outputs), nil
}

func (mocks) Call(args pulumi.MockCallArgs) (resource.PropertyMap, error) {
	return args.Args, nil
}

func TestMain(t *testing.T) {
	err := pulumi.RunErr(func(ctx *pulumi.Context) error {
		infra, err := createInfrastructureCodepipeline(ctx)
		assert.NoError(t, err)

		var wg sync.WaitGroup
		wg.Add(2)

		pulumi.All(infra.codebuild.Name, infra.codebuild.Tags).ApplyT(func(all []interface{}) error {
			name := all[0].(string)
			tags := all[1].(map[string]string)

			assert.Equal(t, "serverlesscodebuildProject", name, "Codebuild project name is wrong , the actual one is %v", name)
			assert.Containsf(t, tags, "Environment", "codebuild project doesn't contain the environment tag name")

			wg.Done()
			return nil
		})

		pulumi.All(infra.codepipeline.Name, infra.codepipeline.Tags).ApplyT(func(all []interface{}) error {
			name := all[0].(string)
			tags := all[1].(map[string]string)

			assert.Equal(t, "serverlesscodepipeline", name, "Codepipeline project name is wrong , the actual one is %v", name)
			assert.Containsf(t, tags, "Environment", "Codepipeline project doesn't contain the environment tag name")

			wg.Done()
			return nil
		})

		wg.Wait()
		return nil
	}, pulumi.WithMocks("iac-codepipeline", "prod", mocks(0)))
	assert.NoError(t, err)
}
