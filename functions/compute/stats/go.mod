module github.com/corecheck/corecheck/functions/compute/stats

go 1.22.3

require (
	github.com/artdarek/go-unzip v1.0.0
	github.com/aws/aws-lambda-go v1.46.0
	github.com/corecheck/corecheck v0.0.0
)

require (
	github.com/aws/aws-sdk-go v1.50.9 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
)

replace github.com/corecheck/corecheck => ../../..
