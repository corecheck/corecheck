.PHONY: dev
dev:
	docker compose up -d
	# terraform apply

.PHONY: down
down:
	docker compose down