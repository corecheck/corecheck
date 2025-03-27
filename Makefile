build-lambdas: build-api-lambdas build-compute-lambdas build-compute-stats-lambda

build-api-lambdas: ./functions/api/*
	for path in $(notdir $^) ; do \
		echo \\n Building $${path} \\n; \
		docker compose run --remove-orphans --rm build-lambda bash -c "\
			cd functions/api/$${path} && \
			go build -modfile=../../../go.mod -mod=readonly -ldflags='-s -w' -o bootstrap && \
			chmod 777 bootstrap" ; \
		docker compose run --remove-orphans --rm util bash -c "\
			zip -j -X deploy/lambdas/api/$${path}.zip functions/api/$${path}/bootstrap && \
			rm functions/api/$${path}/bootstrap" ; \
	done;

build-compute-lambdas: ./functions/compute/*
	# Filter out stats, it uses a separate module configuration.
	for path in $(notdir $(filter-out functions/compute/stats, $^)) ; do \
		echo \\n Building $${path} \\n; \
		docker compose run --remove-orphans --rm build-lambda bash -c "\
			cd functions/compute/$${path} && \
			go build -modfile=../../../go.mod -mod=readonly -ldflags='-s -w' -o bootstrap && \
			chmod 777 bootstrap" ; \
		docker compose run --remove-orphans --rm util bash -c "\
			zip -j -X deploy/lambdas/compute/$${path}.zip functions/compute/$${path}/bootstrap && \
			rm functions/compute/$${path}/bootstrap" ; \
	done;

# Stats is a special case. It uses Go 1.22.3 and has its own go.mod file.
build-compute-stats-lambda:
	echo \\n Building stats \\n; \
	docker compose run --remove-orphans --rm build-compute-stats-lambda bash -c "\
		cd functions/compute/stats && \
		go build -mod=readonly -ldflags='-s -w' -o bootstrap && \
		chmod 777 bootstrap"
	docker compose run --remove-orphans --rm util bash -c "\
		zip -j -X deploy/lambdas/compute/stats.zip functions/compute/stats/bootstrap && \
		rm functions/compute/stats/bootstrap"

build-shell:
	docker compose run --remove-orphans --rm -it build-lambda bash

.PHONY: build-lambdas build-api-lambdas build-compute-lambdas build-compute-stats-lambda
