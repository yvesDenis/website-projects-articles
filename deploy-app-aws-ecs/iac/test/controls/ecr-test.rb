# copyright: 2018, The Authors

title "ECR Repository provisioning test"

control "aws-ecr-check" do                                    
  impact 1.0                                                                
  title "Check to see if base-app-ecr exists" 
  describe aws_ecr_repository(repository_name: 'base-app-repo') do
    it                       { should exist }
    its('image_tag_mutability') { should eq 'IMMUTABLE' }
  end
end