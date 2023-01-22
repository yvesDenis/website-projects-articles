package main

import (
	"github.com/aws/aws-lambda-go/events"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("the Main function", func() {
	var (
		response events.APIGatewayCustomAuthorizerResponse
		request  events.APIGatewayCustomAuthorizerRequestTypeRequest
		err      error
	)

	JustBeforeEach(func() {
		response, err = HandleRequest(request)
	})

	AfterEach(func() {
		request = events.APIGatewayCustomAuthorizerRequestTypeRequest{}
		response = events.APIGatewayCustomAuthorizerResponse{}
	})

	Context("When the auth header is not set", func() {
		It("Denies", func() {
			Expect(err).To(BeNil())
			Expect(response.PolicyDocument.Version).To(Equal("2012-10-17"))
			Expect(response.PolicyDocument.Statement).To(Equal([]events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   "Deny",
					Resource: []string{""},
				},
			}))
		})
	})

	Context("When the auth header is set but it's invalid", func() {
		Context("and auth fails", func() {
			BeforeEach(func() {
				request = events.APIGatewayCustomAuthorizerRequestTypeRequest{
					Headers:   map[string]string{"AUTH": "auth_secrets"},
					MethodArn: "testARN",
				}
			})

			It("Denies", func() {
				Expect(err).To(BeNil())
				Expect(response.PolicyDocument.Version).To(Equal("2012-10-17"))
				Expect(response.PolicyDocument.Statement).To(Equal([]events.IAMPolicyStatement{
					{
						Action:   []string{"execute-api:Invoke"},
						Effect:   "Deny",
						Resource: []string{"testARN"},
					},
				}))
			})
		})

		Context("and auth succeeds", func() {
			BeforeEach(func() {
				request = events.APIGatewayCustomAuthorizerRequestTypeRequest{
					Headers:   map[string]string{"AUTH": "auth_secret"},
					MethodArn: "testARN",
				}
			})

			It("authorizes", func() {
				Expect(err).To(BeNil())
				Expect(response.PolicyDocument.Version).To(Equal("2012-10-17"))
				Expect(response.PolicyDocument.Statement).To(Equal([]events.IAMPolicyStatement{
					{
						Action:   []string{"execute-api:Invoke"},
						Effect:   "Allow",
						Resource: []string{"testARN"},
					},
				}))
			})
		})
	})
})
