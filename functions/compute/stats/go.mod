module github.com/corecheck/corecheck/functions/compute/stats

go 1.26.3

require (
	github.com/aws/aws-lambda-go v1.46.0
	github.com/aws/aws-sdk-go v1.50.9
	github.com/corecheck/corecheck v0.0.0
)

require (
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	golang.org/x/net v0.53.0 // indirect
)

replace github.com/corecheck/corecheck => ../../..
