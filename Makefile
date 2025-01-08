# MIGRATIONS

.PHONY: migrate-up
migrate-up: ## Up goose migrations
	goose -dir database/migrations postgres postgresql://postgres:postgres@127.0.0.1:5435/postgres?sslmode=disable up

.PHONY: migrate-down
migrate-down: ## Up goose migrations
	goose -dir database/migrations postgres postgresql://postgres:postgres@127.0.0.1:5435/postgres?sslmode=disable down

.PHONY: migrate-add
migrate-add: ## Create new migration file, usage: migrate-add [name=<migration_name>]
	goose -dir database/migrations create $(name) sql
