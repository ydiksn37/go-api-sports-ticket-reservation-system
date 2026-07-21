.DEFAULT_GOAL := help

help: ## コマンド一覧を表示する
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)

up: ## Dockerコンテナを起動する (バックグラウンド)
	docker compose up -d

down: ## Dockerコンテナを停止・削除する
	docker compose down

logs: ## コンテナのログをリアルタイムで表示する (Ctrl+Cで終了)
	docker compose logs -f

reset-db: ## DBを完全にリセットする
	docker compose down -v

db-re-up: ## DBを完全にリセット -> 再起動
	docker compose down -v
	docker compose up -d

seed: ## SeedデータをDBに投入する
	docker compose exec app go run cmd/seed/main.go

setup: db-re-up ## 環境の初期セットアップを行う (DBリセット -> Seed投入)
	@echo "Waiting for database and server to start..."
	@sleep 5
	@echo "Inserting seed data..."
	$(MAKE) seed
	@echo "Setup complete! You can access http://localhost:8080/swagger/"

exec-app: ## appコンテナの中に入る (シェル)
	docker compose exec app sh

exec-db: ## dbコンテナの中に入り、直接SQLを叩ける状態にする
	docker compose exec db psql -U postgres -d road_to_dena

ps: ## 稼働中のコンテナ一覧を表示する (docker container ls)
	docker container ls

ps-a: ## 停止中も含めた全てのコンテナ一覧を表示する (docker container ls -a)
	docker container ls -a

images: ## 取得済みのDockerイメージ一覧を表示する (docker image ls)
	docker image ls

images-a: ## 中間レイヤー等を含めた全てのイメージ一覧を表示する (docker image ls -a)
	docker image ls -a

project-ps: ## このプロジェクトに関わるコンテナの状態だけを表示する
	docker compose ps -a

.PHONY: help up down reset-db seed setup logs exec-app exec-db ps ps-a images images-a project-ps
