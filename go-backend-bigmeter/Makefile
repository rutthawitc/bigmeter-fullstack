# BigMeter — developer helpers

SHELL := /bin/bash

.PHONY: help docker-up docker-down docker-restart docker-logs api-build api-local psql migrate seed seed-sample sync-init sync-month sync-scheduler build-sync build-api fmt vet tidy release-publish

help:
	@echo "Targets:"
	@echo "  docker-up       - Build and start Postgres + API"
	@echo "  docker-down     - Stop all containers"
	@echo "  docker-restart  - Restart API container"
	@echo "  docker-logs     - Tail API logs"
	@echo "  api-build       - Build API image only"
	@echo "  api-local       - Run API locally (requires POSTGRES_DSN)"
	@echo "  psql            - Open psql shell into Postgres container"
	@echo "  migrate         - Re-run migrations job"
	@echo "  seed            - Re-run branch seeding job"
	@echo "  seed-sample     - Insert demo rows into Postgres (custcodes/details)"
	@echo "  sync-init       - Run yearly init once (requires ORACLE_DSN, POSTGRES_DSN)"
	@echo "  sync-month      - Run monthly details once (requires ORACLE_DSN, POSTGRES_DSN, YM)"
	@echo "  sync-scheduler  - Run scheduler with cron (requires ORACLE_DSN, POSTGRES_DSN)"
	@echo "  build-sync      - Build sync binary with oracle tag"
	@echo "  build-api       - Build api binary"
	@echo "  fmt vet tidy    - Go formatting, vet, and tidy"
	@echo "  release-publish - Push tag and publish GitHub release (TAG=...)"

docker-up:
	docker compose up -d --build

docker-down:
	docker compose down

docker-restart:
	docker compose restart api

docker-logs:
	docker compose logs -f api

api-build:
	docker compose build api

api-local:
	@[[ -n "$(POSTGRES_DSN)" ]] || (echo "POSTGRES_DSN is required"; exit 1)
	TIMEZONE?=Asia/Bangkok
	BRANCHES?=
	POSTGRES_DSN=$(POSTGRES_DSN) TIMEZONE=$(TIMEZONE) BRANCHES=$(BRANCHES) go run cmd/api/main.go

psql:
	docker compose exec -it postgres psql -U $$POSTGRES_USER -d $$POSTGRES_DB

migrate:
	docker compose --profile setup up -d migrate

seed:
	docker compose --profile setup up -d seed_branches

seed-sample:
	@echo "Seeding sample rows into bm_custcode_init and bm_meter_details..."
	docker compose exec -T postgres psql -U $$POSTGRES_USER -d $$POSTGRES_DB -v ON_ERROR_STOP=1 -c \
	  "INSERT INTO bm_custcode_init (fiscal_year, branch_code, cust_code, use_type, meter_no, meter_state, debt_ym) \
	    VALUES (2025,'BA01','C12345','R','M-0001','N','202410') \
	    ON CONFLICT (fiscal_year, branch_code, cust_code) DO NOTHING;"
	docker compose exec -T postgres psql -U $$POSTGRES_USER -d $$POSTGRES_DB -v ON_ERROR_STOP=1 -c \
	  "INSERT INTO bm_meter_details (year_month, branch_code, org_name, cust_code, use_type, use_name, cust_name, address, route_code, meter_no, meter_size, meter_brand, meter_state, average, present_meter_count, present_water_usg, debt_ym) \
	    VALUES ('202410','BA01','BA01','C12345','R','Residential','John Doe','','RT01','M-0001','1/2','XYZ','N',12.5,300,15,'202410') \
	    ON CONFLICT (year_month, branch_code, cust_code) DO NOTHING;"

# One-shot yearly init (YM is Gregorian). Example:
#   make sync-init YM=202410 POSTGRES_DSN=postgres://... ORACLE_DSN=oracle://...
sync-init:
	@[[ -n "$(ORACLE_DSN)" && -n "$(POSTGRES_DSN)" ]] || (echo "ORACLE_DSN and POSTGRES_DSN required"; exit 1)
	TIMEZONE?=Asia/Bangkok
	YM?=$(shell date +%Y)10
	MODE=init-once YM=$(YM) TIMEZONE=$(TIMEZONE) POSTGRES_DSN=$(POSTGRES_DSN) ORACLE_DSN=$(ORACLE_DSN) go run -tags oracle cmd/sync/main.go

# One-shot monthly details. Example:
#   make sync-month YM=202410 POSTGRES_DSN=postgres://... ORACLE_DSN=oracle://...
sync-month:
	@[[ -n "$(ORACLE_DSN)" && -n "$(POSTGRES_DSN)" ]] || (echo "ORACLE_DSN and POSTGRES_DSN required"; exit 1)
	@[[ -n "$(YM)" ]] || (echo "YM=YYYYMM required"; exit 1)
	TIMEZONE?=Asia/Bangkok
	MODE=month-once YM=$(YM) TIMEZONE=$(TIMEZONE) POSTGRES_DSN=$(POSTGRES_DSN) ORACLE_DSN=$(ORACLE_DSN) go run -tags oracle cmd/sync/main.go

sync-scheduler:
	@[[ -n "$(ORACLE_DSN)" && -n "$(POSTGRES_DSN)" ]] || (echo "ORACLE_DSN and POSTGRES_DSN required"; exit 1)
	TIMEZONE?=Asia/Bangkok
	CRON_YEARLY?=
	CRON_MONTHLY?=
	BRANCHES?=
	TIMEZONE=$(TIMEZONE) POSTGRES_DSN=$(POSTGRES_DSN) ORACLE_DSN=$(ORACLE_DSN) CRON_YEARLY=$(CRON_YEARLY) CRON_MONTHLY=$(CRON_MONTHLY) BRANCHES=$(BRANCHES) go run -tags oracle cmd/sync/main.go

build-sync:
	CGO_ENABLED=1 go build -tags oracle -o bin/sync ./cmd/sync

build-api:
	CGO_ENABLED=0 go build -o bin/api ./cmd/api

fmt:
	gofmt -s -w .

vet:
	go vet ./...

tidy:
	go mod tidy

# Publish a GitHub release from a tag using GitHub CLI (gh)
# Usage: make release-publish TAG=v2025.09.20
release-publish:
	@[[ -n "$(TAG)" ]] || (echo "TAG is required, e.g. make release-publish TAG=v2025.09.20"; exit 1)
	@echo "Pushing git tag $(TAG) to origin..."
	git push origin $(TAG)
	@notes="docs/releases/$(TAG).md"; \
	if [[ -f "$$notes" ]]; then \
	  echo "Creating GitHub release $(TAG) with notes file $$notes"; \
	  gh release create $(TAG) --title "Release $(TAG)" --notes-file "$$notes"; \
	else \
	  echo "Notes file $$notes not found; using CHANGELOG.md"; \
	  gh release create $(TAG) --title "Release $(TAG)" --notes "See CHANGELOG.md for details."; \
	fi
