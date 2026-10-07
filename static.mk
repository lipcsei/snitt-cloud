# static.mk - közös statikus elemzés minden lipcsei repóhoz.
#
# FORRÁS: lipcsei/commons, ops/static/static.mk. Ez a fájl a repókban MÁSOLAT: ne itt szerkeszd,
# hanem a commonsban, és onnan frissítsd (make static-sync). Azért másolat és nem submodule-ból
# include, mert a CI-nak így nem kell hozzá a commons deploy kulcsa, és a commons nélküli repókban
# (sso, edge, snitt-cloud ...) is ugyanaz fut.
#
# Használat a repó Makefile-jában:
#     include static.mk
#     check: lint test build static
#
# Minden cél magától megtalálja a rá tartozó fájlokat, és ha nincs ilyen, szó nélkül átugorja -
# ugyanaz a fájl jó egy csak-compose repóra és egy Go+React appra is. A szabályok repónként a
# gyökérben lévő konfigurációkból jönnek (.hadolint.yaml, .yamllint.yaml, .sqlfluff, .golangci.yml,
# .gitleaks.toml, .semgrepignore, .trivyignore, .github/zizmor.yml).
#
# Az eszközök rögzített verziójú Docker-képből futnak (helyben és a CI-ban ugyanaz), kivéve a
# Go-eszközöket: azok `go run ...@verzió`-val, hogy a privát modulokat a saját Go-beállításoddal érjék el.
# GNU Make 3.81-gyel (macOS /usr/bin/make) is működnie kell.

# Amit egyik eszköz se nézzen. Repónként felülírható az include ELŐTT.
STATIC_EXCLUDE ?= (^|/)(\.git|node_modules|\.commons|third_party|vendor|dist)/
# A Go modulok könyvtárai; alapból minden go.mod. Felülírható: STATIC_GO_MODULES := backend shared
STATIC_GO_MODULES ?=
STATIC_GOVET_ARGS ?=
STATIC_GOLANGCI_ARGS ?=
STATIC_SEMGREP_CONFIG ?= p/default
# Kizárt semgrep-szabályok. A secrets-inherit a saját deploy workflow hívására szól; ugyanezt a
# zizmor is nézi, és ott (.github/zizmor.yml) fájlra szűkítve, indoklással van elengedve.
STATIC_SEMGREP_EXCLUDE_RULES ?= yaml.github-actions.security.secrets-inherit.secrets-inherit
# További semgrep kapcsolók repónként.
STATIC_SEMGREP_ARGS ?=
STATIC_SEVERITY ?= HIGH,CRITICAL
# Alapból csak az, amire már van javítás: amin nem tudsz változtatni, az ne állítsa meg a CI-t.
# Szigorúbb repóban üresre állítható (STATIC_TRIVY_VULN_ARGS :=).
STATIC_TRIVY_VULN_ARGS ?= --ignore-unfixed

STATIC_HADOLINT_IMAGE ?= hadolint/hadolint:v2.15.1
STATIC_YAMLLINT_IMAGE ?= cytopia/yamllint:1
STATIC_SQLFLUFF_IMAGE ?= sqlfluff/sqlfluff:4.4.0
STATIC_SHELLCHECK_IMAGE ?= koalaman/shellcheck:v0.11.0
STATIC_ACTIONLINT_IMAGE ?= rhysd/actionlint:1.7.12
STATIC_ZIZMOR_IMAGE ?= ghcr.io/zizmorcore/zizmor:1.30.1
STATIC_GITLEAKS_IMAGE ?= ghcr.io/gitleaks/gitleaks:v8.30.1
STATIC_SEMGREP_IMAGE ?= semgrep/semgrep:1.178.0
STATIC_TRIVY_IMAGE ?= aquasec/trivy:0.75.0

STATIC_GOLANGCI ?= go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.14.0
# Szándékosan @latest: a govulncheck a saját elemzőjével együtt frissül az adatbázishoz.
STATIC_GOVULNCHECK ?= go run golang.org/x/vuln/cmd/govulncheck@latest
STATIC_SQLC ?= go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.31.1

STATIC_DOCKER := docker run --rm -v "$(CURDIR):/work" -w /work
# A trivy adatbázisa és szabálykészlete ne töltődjön le minden futásnál.
# A sima .trivyignore csak azonosítóra szól (az egész repóra); fájlra szűkített kivételhez
# .trivyignore.yaml kell, azt a trivy csak kérésre olvassa.
STATIC_TRIVY_IGNORE := $(if $(wildcard .trivyignore.yaml),--ignorefile .trivyignore.yaml,)
STATIC_TRIVY := docker run --rm -v "$(CURDIR):/work" -w /work -v "$(HOME)/.cache/trivy:/root/.cache/trivy" $(STATIC_TRIVY_IMAGE)
# '**/node_modules': a beágyazott (frontend/node_modules) könyvtárakat is, különben a trivy helyben
# percekig turkál bennük, és időtúllépéssel elszáll.
STATIC_SKIP_DIRS := --skip-dirs '**/node_modules' --skip-dirs .commons --skip-dirs third_party --skip-dirs vendor

# A repó fájljai: a git által ismertek (követett + új, de nem ignorált), git nélkül minden fájl.
# A törölt, de még követett fájlokat a -f szűri ki.
STATIC_LS := { git ls-files -co --exclude-standard 2>/dev/null || find . -type f | sed 's|^\./||'; } \
	| grep -vE '$(STATIC_EXCLUDE)' | while read -r f; do [ -f "$$f" ] && echo "$$f"; done
STATIC_GO_DIRS := if [ -n "$(STATIC_GO_MODULES)" ]; then echo $(STATIC_GO_MODULES) | tr ' ' '\n'; \
	else $(STATIC_LS) | grep -E '(^|/)go\.mod$$' | xargs -n1 dirname 2>/dev/null; fi

.PHONY: static static-docker static-yaml static-sql static-sqlc static-shell static-actions static-go \
        static-secrets static-semgrep static-trivy static-vuln static-sync

static: static-docker static-yaml static-sql static-sqlc static-shell static-actions static-go static-secrets static-semgrep static-trivy ## Minden statikus elemzés (a sebezhetőség-keresés külön: make static-vuln)

static-docker: ## hadolint a Dockerfile-okra
	@files=$$($(STATIC_LS) | grep -E '(^|/)Dockerfile[^/]*$$'); \
	if [ -z "$$files" ]; then echo "hadolint: nincs Dockerfile, kihagyva"; exit 0; fi; \
	echo "hadolint:" $$files; \
	$(STATIC_DOCKER) $(STATIC_HADOLINT_IMAGE) hadolint $$files

static-yaml: ## yamllint minden YAML-fájlra
	@files=$$($(STATIC_LS) | grep -E '\.ya?ml$$'); \
	if [ -z "$$files" ]; then echo "yamllint: nincs YAML, kihagyva"; exit 0; fi; \
	echo "yamllint: $$(echo "$$files" | wc -l | tr -d ' ') fájl"; \
	$(STATIC_DOCKER) $(STATIC_YAMLLINT_IMAGE) $$files

static-sql: ## sqlfluff a migrációkra és a lekérdezésekre
	@files=$$($(STATIC_LS) | grep -E '\.sql$$'); \
	if [ -z "$$files" ]; then echo "sqlfluff: nincs SQL, kihagyva"; exit 0; fi; \
	echo "sqlfluff: $$(echo "$$files" | wc -l | tr -d ' ') fájl"; \
	$(STATIC_DOCKER) $(STATIC_SQLFLUFF_IMAGE) lint --disable-progress-bar $$files

static-sqlc: ## sqlc vet: a sqlc.yaml `rules:` szabályai a lekérdezésekre (pl. WHERE nélküli DELETE)
	@files=$$($(STATIC_LS) | grep -E '(^|/)sqlc\.ya?ml$$'); \
	if [ -z "$$files" ]; then echo "sqlc vet: nincs sqlc.yaml, kihagyva"; exit 0; fi; \
	for f in $$files; do echo "sqlc vet: $$f"; (cd "$$(dirname "$$f")" && $(STATIC_SQLC) vet -f "$$(basename "$$f")") || exit 1; done

static-shell: ## shellcheck a .sh szkriptekre
	@files=$$($(STATIC_LS) | grep -E '\.sh$$'); \
	if [ -z "$$files" ]; then echo "shellcheck: nincs szkript, kihagyva"; exit 0; fi; \
	echo "shellcheck:" $$files; \
	$(STATIC_DOCKER) $(STATIC_SHELLCHECK_IMAGE) -x --severity=warning $$files

static-actions: ## actionlint + zizmor a GitHub Actions workflow-kra
	@if [ ! -d .github/workflows ]; then echo "actionlint, zizmor: nincs workflow, kihagyva"; exit 0; fi; \
	echo "actionlint"; $(STATIC_DOCKER) $(STATIC_ACTIONLINT_IMAGE) -color || exit 1; \
	echo "zizmor"; $(STATIC_DOCKER) $(STATIC_ZIZMOR_IMAGE) --offline --no-progress .

# --max-same-issues/--max-issues-per-linter 0: a golangci alapból azonos szövegű találatból 3, linterenként
# 50 után elhallgat, így egy javítás után újabbak bukkannának elő.
static-go: ## go vet + golangci-lint (benne a gosec) minden Go modulra
	@mods=$$($(STATIC_GO_DIRS)); \
	if [ -z "$$mods" ]; then echo "go vet, golangci-lint: nincs Go modul, kihagyva"; exit 0; fi; \
	for m in $$mods; do \
	  echo "go vet + golangci-lint: $$m"; \
	  (cd "$$m" && go vet $(STATIC_GOVET_ARGS) ./... && $(STATIC_GOLANGCI) run --timeout 10m --max-same-issues 0 --max-issues-per-linter 0 $(STATIC_GOLANGCI_ARGS) ./...) || exit 1; \
	done

# A közös git-könyvtárat külön is csatoljuk: worktree-ben a .git csak mutató a fő repóra, és nélküle
# a gitleaks 0 commitot látna, mégis zölden lépne ki - ezért 0 commitnál el is bukik.
static-secrets: ## gitleaks: titkok a git-történetben (git nélkül a munkafában)
	@if git rev-parse --git-dir >/dev/null 2>&1 && git rev-parse HEAD >/dev/null 2>&1; then \
	  echo "gitleaks (git-történet)"; \
	  gd=$$(cd "$$(git rev-parse --git-common-dir)" && pwd); \
	  out=$$($(STATIC_DOCKER) -v "$$gd:$$gd:ro" \
	    -e GIT_CONFIG_COUNT=1 -e GIT_CONFIG_KEY_0=safe.directory -e 'GIT_CONFIG_VALUE_0=*' \
	    $(STATIC_GITLEAKS_IMAGE) git --no-banner --no-color --redact . 2>&1); rc=$$?; echo "$$out"; \
	  [ $$rc -eq 0 ] || exit $$rc; \
	  if echo "$$out" | grep -q ' 0 commits scanned'; then \
	    echo "gitleaks: 0 commitot látott - nem érte el a git-történetet, ez NEM tiszta eredmény" >&2; exit 1; \
	  fi; \
	else \
	  echo "gitleaks (munkafa, nincs git-történet)"; \
	  $(STATIC_DOCKER) $(STATIC_GITLEAKS_IMAGE) dir --no-banner --redact .; \
	fi

static-semgrep: ## semgrep: biztonsági minták a kódban és a konfigurációban
	@echo "semgrep ($(STATIC_SEMGREP_CONFIG))"; \
	$(STATIC_DOCKER) $(STATIC_SEMGREP_IMAGE) semgrep scan --config $(STATIC_SEMGREP_CONFIG) \
	  --metrics=off --error --quiet --exclude .commons --exclude third_party --exclude vendor \
	  $(addprefix --exclude-rule ,$(STATIC_SEMGREP_EXCLUDE_RULES)) $(STATIC_SEMGREP_ARGS) .

static-trivy: ## trivy: a Dockerfile-ok és az infrastruktúra-fájlok hibás beállításai
	@echo "trivy (misconfig)"; mkdir -p "$(HOME)/.cache/trivy"; \
	$(STATIC_TRIVY) fs --quiet --scanners misconfig --severity $(STATIC_SEVERITY) --exit-code 1 $(STATIC_TRIVY_IGNORE) $(STATIC_SKIP_DIRS) .

# Nem a `static` része: friss adatbázist tölt le, ezért a kimenete a kód változása nélkül is
# változhat - egy ma bejelentett CVE holnap megbuktatja, és ez a szándék. A CI külön jobban futtatja.
static-vuln: ## Ismert sebezhetőségek a függőségekben (trivy + govulncheck)
	@echo "trivy (függőségek)"; mkdir -p "$(HOME)/.cache/trivy"; \
	$(STATIC_TRIVY) fs --quiet --scanners vuln $(STATIC_TRIVY_VULN_ARGS) --severity $(STATIC_SEVERITY) --exit-code 1 $(STATIC_TRIVY_IGNORE) $(STATIC_SKIP_DIRS) . || exit 1; \
	for m in $$($(STATIC_GO_DIRS)); do echo "govulncheck: $$m"; (cd "$$m" && $(STATIC_GOVULNCHECK) ./...) || exit 1; done

COMMONS ?= ../commons
static-sync: ## A static.mk frissítése a commonsból (COMMONS=../commons)
	cp "$(COMMONS)/ops/static/static.mk" static.mk
