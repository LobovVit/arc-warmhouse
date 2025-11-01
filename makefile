# ------------------------------------------------------------
# Warmhouse / SmartHome — Makefile
# Генерация диаграмм (PlantUML), REST-доков (OpenAPI/Redoc),
# и AsyncAPI-доков (MQTT).
# ------------------------------------------------------------

# --- Исходники / каталоги ---
C4_DIR         := docs/c4
ERD_DIR        := docs/erd
PUML_GLOBS     := $(C4_DIR)/*.puml $(ERD_DIR)/*.puml
IMG_DIR        := ../images

OUT_DIR        := docs/api
DIAGS_OUT      := $(OUT_DIR)

### ===================== OpenAPI (Redocly) =====================

# исходники
OPENAPI_FACADE_SRC    := schemas/openapi-smarthome.yml
OPENAPI_DEVICE_SRC    := schemas/openapi-device.yml
OPENAPI_TELEM_SRC     := schemas/openapi-telemetry.yml

# артефакты
OPENAPI_DIR_BUNDLES   := schemas/_bundles
OPENAPI_DIR_DOCS      := docs/api/openapi

OPENAPI_FACADE_BUND   := $(OPENAPI_DIR_BUNDLES)/openapi-smarthome.bundled.yml
OPENAPI_FACADE_HTML   := $(OPENAPI_DIR_DOCS)/openapi-smarthome.html
OPENAPI_DEVICE_HTML   := $(OPENAPI_DIR_DOCS)/openapi-device.html
OPENAPI_TELEM_HTML    := $(OPENAPI_DIR_DOCS)/openapi-telemetry.html

# обёртка для вызова redocly из Docker
REDOCLY := docker run --rm -v "$$(pwd)":/work -w /work redocly/cli:latest

ASYNCAPI_SRC         := schemas/asyncapi-smarthome.yml
ASYNCAPI_HTML_DIR    := $(OUT_DIR)/asyncapi
ASYNCAPI_INDEX       := $(ASYNCAPI_HTML_DIR)/index.html

# ------------------------------------------------------------
# Основные цели
# ------------------------------------------------------------
.PHONY: all docs redoc diagrams openapi-docs asyncapi

all: docs
docs: openapi-docs asyncapi diagrams
	@echo "✅ All docs are ready:"
	@echo "   - C4/ERD PNG:      $(OUT_DIR)"
	@echo "   - OpenAPI (HTML):  $(OPENAPI_OUT)"
	@echo "   - AsyncAPI (HTML): $(ASYNCAPI_INDEX)"

# ------------------------------------------------------------
# puml → png
# ------------------------------------------------------------
.PHONY: diagrams
diagrams:
	@mkdir -p $(IMG_DIR)
	@plantuml $(PUML_GLOBS) -o $(IMG_DIR)

# ------------------------------------------------------------
# OpenAPI → HTML
# ------------------------------------------------------------

openapi-docs: openapi-lint redoc-facade redoc-device redoc-telemetry
	@echo "📄 OpenAPI docs ready:"
	@echo "   - $(OPENAPI_FACADE_HTML)"
	@echo "   - $(OPENAPI_DEVICE_HTML)"
	@echo "   - $(OPENAPI_TELEM_HTML)"

.PHONY: openapi-lint
openapi-lint:
	@echo "🔍 Lint OpenAPI specs..."
	$(REDOCLY) lint $(OPENAPI_FACADE_SRC)
	$(REDOCLY) lint $(OPENAPI_DEVICE_SRC)
	$(REDOCLY) lint $(OPENAPI_TELEM_SRC)
	@echo "✅ Lint OK"

.PHONY: openapi-bundle-facade
openapi-bundle-facade: $(OPENAPI_FACADE_BUND)

$(OPENAPI_FACADE_BUND): $(OPENAPI_FACADE_SRC) $(OPENAPI_DEVICE_SRC) $(OPENAPI_TELEM_SRC)
	@mkdir -p $(OPENAPI_DIR_BUNDLES)
	@echo "🧩 Bundling facade (resolving $${ref})..."
	$(REDOCLY) bundle $(OPENAPI_FACADE_SRC) --output $(OPENAPI_FACADE_BUND)
	@echo "✅ Bundled → $(OPENAPI_FACADE_BUND)"

.PHONY: redoc-facade redoc-device redoc-telemetry openapi-docs
redoc-facade: openapi-bundle-facade
	@mkdir -p $(OPENAPI_DIR_DOCS)
	@echo "📘 Building Facade docs..."
	$(REDOCLY) build-docs $(OPENAPI_FACADE_BUND) --output $(OPENAPI_FACADE_HTML)
	@echo "✅ Facade docs → $(OPENAPI_FACADE_HTML)"

redoc-device:
	@mkdir -p $(OPENAPI_DIR_DOCS)
	@echo "📗 Building Device docs..."
	$(REDOCLY) build-docs $(OPENAPI_DEVICE_SRC) --output $(OPENAPI_DEVICE_HTML)
	@echo "✅ Device docs → $(OPENAPI_DEVICE_HTML)"

redoc-telemetry:
	@mkdir -p $(OPENAPI_DIR_DOCS)
	@echo "📙 Building Telemetry docs..."
	$(REDOCLY) build-docs $(OPENAPI_TELEM_SRC) --output $(OPENAPI_TELEM_HTML)
	@echo "✅ Telemetry docs → $(OPENAPI_TELEM_HTML)"

openapi-docs: openapi-lint redoc-facade redoc-device redoc-telemetry
	@echo "📄 OpenAPI docs ready:"
	@echo "   - $(OPENAPI_FACADE_HTML)"
	@echo "   - $(OPENAPI_DEVICE_HTML)"
	@echo "   - $(OPENAPI_TELEM_HTML)"

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