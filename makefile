# ------------------------------------------------------------
# Warmhouse / SmartHome — Makefile
# Генерация диаграмм (PlantUML), REST-доков (OpenAPI/Redoc),
# и AsyncAPI-доков (MQTT) через локальный шаблон.
# ------------------------------------------------------------

# --- Исходники / каталоги ---
C4_DIR         := docs/c4
ERD_DIR        := docs/erd
PUML_GLOBS     := $(C4_DIR)/*.puml $(ERD_DIR)/*.puml
IMG_DIR        := ../images

OUT_DIR        := docs/api
DIAGS_OUT      := $(OUT_DIR)

OPENAPI_SRC    := schemas/openapi-smarthome.yml
OPENAPI_OUT    := $(OUT_DIR)/openapi/openapi.html

ASYNCAPI_SRC         := schemas/asyncapi-smarthome.yml
ASYNCAPI_HTML_DIR    := $(OUT_DIR)/asyncapi
ASYNCAPI_INDEX       := $(ASYNCAPI_HTML_DIR)/index.html

# Локальный шаблон для AsyncAPI (совместим с генератором v1)
# ASYNCAPI_TEMPLATE_DIR := docs/templates/html-template-v0.29.0
# ASYNCAPI_TEMPLATE_GIT := https://github.com/asyncapi/html-template.git
# ASYNCAPI_TEMPLATE_TAG := v0.28.0

# ------------------------------------------------------------
# Основные цели
# ------------------------------------------------------------
.PHONY: all docs redoc puml asyncapi clean serve-docs open

all: docs
docs: redoc asyncapi puml
	@echo "✅ All docs are ready:"
	@echo "   - C4/ERD PNG:      $(OUT_DIR)"
	@echo "   - OpenAPI (HTML):  $(OPENAPI_OUT)"
	@echo "   - AsyncAPI (HTML): $(ASYNCAPI_INDEX)"

# ------------------------------------------------------------
# puml → png
# ------------------------------------------------------------
.PHONY: puml
puml:
	@mkdir -p $(IMG_DIR)
	@plantuml $(PUML_GLOBS) -o $(IMG_DIR)

# ------------------------------------------------------------
# OpenAPI → HTML
# ------------------------------------------------------------
.PHONY: redoc
redoc:
	@mkdir -p $(OUT_DIR)
	@echo "📘 Building OpenAPI docs via Redocly (Docker)..."
	docker run --rm -v "$$(pwd)":/work -w /work redocly/cli:latest \
		build-docs $(OPENAPI_SRC) --output $(OPENAPI_OUT)
	@echo "✅ OpenAPI docs ready at $(OPENAPI_OUT)"

# ------------------------------------------------------------
# AsyncAPI → HTML
# ------------------------------------------------------------
.PHONY: asyncapi
asyncapi:
	@mkdir -p $(OUT_DIR)
	@echo "📡 Building AsyncAPI docs (local npx, html-template@0.28.0)..."
	npx -y -p @asyncapi/cli@1 -p @asyncapi/html-template@0.28.0 \
		asyncapi generate fromTemplate $(ASYNCAPI_SRC) @asyncapi/html-template@0.28.0 \
		--output $(OUT_DIR)/asyncapi --force-write
	@echo "✅ AsyncAPI docs ready at $(OUT_DIR)/asyncapi/index.html"

# ------------------------------------------------------------
# Сервисные цели
# ------------------------------------------------------------
.PHONY: serve-docs open clean

serve-docs:
	@echo "🌐 Serving docs at http://localhost:8080 ..."
	@cd docs/api && python3 -m http.server 8080

open:
	@echo "🖥  Opening docs..."
	@if command -v xdg-open >/dev/null 2>&1; then \
		xdg-open "$(OPENAPI_OUT)"; \
		xdg-open "$(ASYNCAPI_INDEX)"; \
	elif command -v open >/dev/null 2>&1; then \
		open "$(OPENAPI_OUT)"; \
		open "$(ASYNCAPI_INDEX)"; \
	else \
		echo "ℹ️  Please open manually:"; \
		echo "    - $(OPENAPI_OUT)"; \
		echo "    - $(ASYNCAPI_INDEX)"; \
	fi

clean:
	@echo "🧹 Cleaning generated docs..."
	@rm -rf $(OUT_DIR)/openapi $(ASYNCAPI_HTML_DIR)
	@echo "✅ Clean complete"