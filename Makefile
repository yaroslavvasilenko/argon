# MIGRATIONS

.PHONY: migrate-add
migrate-add: ## Create new migration file, usage: migrate-add [name=<migration_name>]
	goose -dir database/migrations create $(name) sql
