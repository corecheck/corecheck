GO_VERSION=1.21.1

build-lambdas: build-api-lambdas build-compute-lambdas build-compute-stats-lambda

build-api-lambdas: ./functions/api/*
	for path in $(notdir $^) ; do \
		echo \\n Building $${path} \\n; \
		docker run -e GOOS=linux -e GOARCH=arm64 -e CGO_ENABLED=0 -e GOFLAGS=-trimpath -v ./:/app -w /app golang:$(GO_VERSION) bash -c "cd functions/api/$${path} && go build -modfile=../../../go.mod -mod=readonly -ldflags='-s -w' -o bootstrap && chmod 777 bootstrap" ; \
		zip -j deploy/lambdas/api/$${path}.zip functions/api/$${path}/bootstrap ; \
		rm functions/api/$${path}/bootstrap ; \
	done;


build-compute-lambdas: ./functions/compute/*
	# Filter out stats, it uses a separate module configuration.
	for path in $(notdir $(filter-out functions/compute/stats, $^)) ; do \
		echo \\n Building $${path} \\n; \
		docker run -e GOOS=linux -e GOARCH=arm64 -e CGO_ENABLED=0 -e GOFLAGS=-trimpath -v ./:/app -w /app golang:$(GO_VERSION) bash -c "cd functions/compute/$${path} && go build -modfile=../../../go.mod -mod=readonly -ldflags='-s -w' -o bootstrap && chmod 777 bootstrap" ; \
		zip -j deploy/lambdas/compute/$${path}.zip functions/compute/$${path}/bootstrap ; \
		rm functions/compute/$${path}/bootstrap ; \
	done;

# Stats is a special case. It uses Go 1.22.3 and has its own go.mod file.
build-compute-stats-lambda:
	echo \\n Building stats \\n; \
	docker run -e GOOS=linux -e GOARCH=arm64 -e CGO_ENABLED=0 -e GOFLAGS=-trimpath -v ./:/app -w /app golang:1.22.3 bash -c "cd functions/compute/stats && go build -mod=readonly -ldflags='-s -w' -o bootstrap && chmod 777 bootstrap"
	zip -j deploy/lambdas/compute/stats.zip functions/compute/stats/bootstrap
	rm functions/compute/stats/bootstrap

build-shell:
	docker run -e GOOS=linux -e GOARCH=arm64 -e CGO_ENABLED=0 -e GOFLAGS=-trimpath -v ./:/app -w /app -it golang:$(GO_VERSION) bash

# Requires AWS credentials set along with environment variables for
#  - CORECHECK_S3_API_BUCKET
#  - CORECHECK_S3_COMPUTE_BUCKET
# set to the values consistent with the Terraform configuration for the deploy environment.
deploy-lambdas:
	aws s3 cp --recursive --exclude "*" --include "*.zip" deploy/lambdas/api/ s3://$${CORECHECK_S3_API_BUCKET}/
	aws s3 cp --recursive --exclude "*" --include "*.zip" deploy/lambdas/compute/ s3://$${CORECHECK_S3_COMPUTE_BUCKET}/

.PHONY: build-lambdas build-api-lambdas build-compute-lambdas build-compute-stats-lambda deploy-lambdas
