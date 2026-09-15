.DEFAULT_GOAL := help

help: ## コマンド一覧を表示する
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

up: ## コンテナを起動する
	docker compose up -d

seed: ## Seedデータを投入する
	docker compose exec app go run cmd/seed/main.go

test: ## ユニットテストを実行する
	go test -v ./...

logs: ## ログをリアルタイム表示する
	docker compose logs -f

exec-app: ## appコンテナの中に入る (シェル)
	docker compose exec app sh

exec-db: ## dbコンテナの中に入る (psql)
	docker compose exec db psql -U postgres -d road_to_dena

clean: ## コンテナを停止する
	docker compose down

fclean: ## コンテナとボリュームを完全に削除する
	docker compose down -v --remove-orphans

re: fclean up ## 完全削除から再起動し、Seedデータを投入する
	@echo "Waiting for database to be ready..."
	@sleep 5
	@$(MAKE) seed
	@echo "Re-initialization complete"

.PHONY: help up clean fclean re seed test logs exec-app exec-db
