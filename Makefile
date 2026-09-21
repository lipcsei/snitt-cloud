.DEFAULT_GOAL := help
SHELL := /bin/bash

COMPOSE := docker compose -f deploy/docker-compose.yml
PROD_CONFIG := docker compose --env-file deploy/.env.prod.example -f deploy/docker-compose.yml -f deploy/docker-compose.prod.yml config
APPS := landing admin

# A mentő szkriptek compose-a ebből az env-fájlból helyettesít. Helyben nem kell (mindennek van
# alapértéke); élesben, a VPS-en: make backup DOTENV=deploy/.env.prod BACKUP_DIR=/var/backups/snitt
DOTENV ?=
SCRIPT_ENV := $(if $(DOTENV),COMPOSE_ENV_FILES=$(abspath $(DOTENV)))

.PHONY: help need-sso up down logs ps install dev-account dev-landing dev-admin og \
        fmt lint test build compose-check check backup restore

help: ## Elérhető parancsok
	@grep -E '^[a-z-]+:.*## ' $(MAKEFILE_LIST) | awk -F':.*## ' '{printf "  make %-14s %s\n", $$1, $$2}'

need-sso:
	@docker network inspect sso-net >/dev/null 2>&1 || { echo "Nincs sso-net hálózat: előbb a közös Keycloakot indítsd (make -C ../sso up)." >&2; exit 1; }

## ---- Fejlesztői stack (Docker Compose) ----
up: need-sso ## Postgres + account + admin + landing (a közös sso Keycloaknak előbb futnia kell)
	$(COMPOSE) up --build -d
	@echo "API: http://localhost:8090   Admin: http://localhost:5175   Landing: http://localhost:8088   Keycloak: http://localhost:8081"

down: ## Leállítás (az adatok megmaradnak; teljes törlés: docker compose -f deploy/docker-compose.yml down -v)
	$(COMPOSE) down

logs: ## Az account szolgáltatás naplója
	$(COMPOSE) logs -f account

ps: ## A konténerek állapota
	$(COMPOSE) ps

## ---- Helyi fejlesztés, konténer nélkül ----
install: ## A landing és az admin függőségei (npm ci)
	for app in $(APPS); do (cd $$app && npm ci) || exit 1; done

services/account/.env:
	cp services/account/.env.example services/account/.env

dev-account: services/account/.env ## Account szolgáltatás helyben (:8090; a compose account konténerét leállítja)
	$(COMPOSE) stop account
	cd services/account && set -a && source .env && set +a && go run ./cmd/server

dev-landing: ## Landing Vite dev szerver (http://localhost:5174)
	cd landing && npm run dev

dev-admin: ## Admin Vite dev szerver (http://localhost:5175; a compose admin konténerét leállítja)
	$(COMPOSE) stop admin
	cd admin && npm run dev

og: ## A landing megosztási képeinek (og:image) újragyártása
	cd landing && npm run og

## ---- Ellenőrzés ----
fmt: ## gofmt az account szolgáltatásra
	cd services/account && gofmt -w .

lint: ## gofmt-ellenőrzés, go vet, és tsc a landingre meg az adminra
	@cd services/account && unformatted=$$(gofmt -l .) && if [ -n "$$unformatted" ]; then echo "Nem gofmt-elt fájlok (make fmt):"; echo "$$unformatted"; exit 1; fi
	cd services/account && go vet ./...
	for app in $(APPS); do (cd $$app && npx tsc --noEmit) || exit 1; done

test: ## Az account szolgáltatás tesztjei (adatbázis és Keycloak nélkül)
	cd services/account && go test -count=1 ./...

build: ## Az account szolgáltatás, a landing és az admin fordítása
	cd services/account && go build ./...
	for app in $(APPS); do (cd $$app && npm run build) || exit 1; done

compose-check: ## A fejlesztői és az éles compose rendben van
	$(COMPOSE) config -q
	$(PROD_CONFIG) -q

check: lint test build ## Amit a CI futtat

## ---- Mentés ----
backup: ## A "snitt" Postgres-adatbázis mentése: make backup BACKUP_DIR=/mentesek/helye
	@test -n "$(BACKUP_DIR)" || { echo "Add meg, hova mentsen: make backup BACKUP_DIR=/mentesek/helye" >&2; exit 1; }
	$(SCRIPT_ENV) ./deploy/scripts/backup.sh "$(BACKUP_DIR)"

restore: ## Visszaállítás (FELÜLÍRJA az adatbázist!): make restore FILE=/mentesek/snitt-….sql.gz
	@test -n "$(FILE)" || { echo "Add meg a mentést: make restore FILE=/mentesek/snitt-….sql.gz" >&2; exit 1; }
	$(SCRIPT_ENV) ./deploy/scripts/restore.sh "$(FILE)"
