# Project Architecture & Master Implementation Plan: `go-email-template-sender`

> **Overview**: `go-email-template-sender` is a dynamic email templating, CSS styling, arbitrary JSON context data-binding, and email delivery orchestration service written in Golang (Go 1.22+). It is designed adhering strictly to **Hexagonal Architecture (Ports and Adapters)** and **Domain-Driven Design (DDD)** principles, backed by PostgreSQL and Mailpit/SMTP.

---

## 🏛️ 1. Architecture & Design Principles

### A. Hexagonal Architecture (Ports and Adapters)
The codebase enforces strict layer isolation:
* **Domain Layer (`internal/domain`)**: Pure Go domain entities (`Context`, `Stylesheet`, `Template`, `Binding`) free from database tags or HTTP annotations.
* **Inbound Driving Ports (`internal/application/port/in`)**: Go interface contracts defining all application use cases.
* **Outbound Driven Ports (`internal/application/port/out`)**: Go interface contracts defining persistence and email delivery capabilities.
* **Application Services (`internal/application/service`)**: Pure use case orchestrators executing business rules, Handlebars template rendering (`aymerick/raymond`), and CSS link injection (`PuerkitoBio/goquery`).
* **Inbound Driving Adapters (`internal/adapter/in/http`)**: REST HTTP transport powered by Go 1.22 `net/http` standard library router. Implements OpenAPI-generated `gen.ServerInterface`.
* **Outbound Driven Adapters**:
  * **PostgreSQL (`internal/adapter/out/postgres`)**: SQL persistence utilizing `pgx/v5` pool (`pgxpool`), supporting native PostgreSQL `JSONB` querying operators (`payload->>'field'`).
  * **Email Sender (`internal/adapter/out/email`)**: SMTP client connected to Mailpit.

---

## 🗄️ 2. Database Schema (`schema.sql`)

Defined in [`schema.sql`](file:///Users/ernstvorsteveld/git/go/go-email-template-sender/schema.sql):

```sql
CREATE TABLE IF NOT EXISTS contexts (
    id UUID PRIMARY KEY,
    reference_id VARCHAR(255) NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    email_jsonpath VARCHAR(255) NOT NULL
);

-- GIN index for high-performance JSONB querying over arbitrary payloads
CREATE INDEX IF NOT EXISTS idx_contexts_payload_gin ON contexts USING GIN (payload);

CREATE TABLE IF NOT EXISTS stylesheets (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(255) UNIQUE NOT NULL,
    css_content TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS templates (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(255) UNIQUE NOT NULL,
    version INT NOT NULL,
    stylesheet_id UUID REFERENCES stylesheets(id),
    html_content TEXT NOT NULL,
    subject VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS bindings (
    id UUID PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    query TEXT NOT NULL,
    template_id UUID REFERENCES templates(id) NOT NULL
);
```

---

## 📄 3. OpenAPI 3.0 Specification

The complete API specification is defined in [`openapi.yaml`](file:///Users/ernstvorsteveld/git/go/go-email-template-sender/openapi.yaml).

* **Specification File**: [`openapi.yaml`](file:///Users/ernstvorsteveld/git/go/go-email-template-sender/openapi.yaml)
* **API Endpoints Summary**:
  * `POST /contexts` & `GET /contexts` (Create & List contexts)
  * `GET /contexts/{id}` & `PUT /contexts/{id}` (Get & Update context)
  * `POST /stylesheets` & `GET /stylesheets` (Create & List stylesheets)
  * `GET /stylesheets/{id}` & `PUT /stylesheets/{id}` (Get & Update stylesheet)
  * `POST /templates` & `GET /templates` (Create & List templates)
  * `GET /templates/{id}` & `PUT /templates/{id}` (Get & Update template)
  * `GET /templates/{id}/render` (Render HTML with linked CSS injected)
  * `POST /bindings` & `GET /bindings` (Create & List bindings)
  * `GET /bindings/{id}` & `PUT /bindings/{id}` (Get & Update binding)
  * `POST /deliveries` (Trigger email dispatch workflow)

---

## 🛠️ 4. Code Generation & Subpackage Isolation

### Generator Configuration (`oapi-codegen.yaml`)
Defined in [`oapi-codegen.yaml`](file:///Users/ernstvorsteveld/git/go/go-email-template-sender/oapi-codegen.yaml):

```yaml
package: gen
output: internal/adapter/in/http/gen/api.gen.go
generate:
  models: true
  std-http-server: true
```

### Git Exclusion (`.gitignore`)
Defined in [`.gitignore`](file:///Users/ernstvorsteveld/git/go/go-email-template-sender/.gitignore):

```gitignore
# Binaries
bin/
/server

# Generated OpenAPI code
*.gen.go
internal/adapter/in/http/gen/
```

### Code Generation Target (`Makefile`)
Defined in [`Makefile`](file:///Users/ernstvorsteveld/git/go/go-email-template-sender/Makefile):

```makefile
generate:
	@echo "Generating API server and types from openapi.yaml..."
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest -config oapi-codegen.yaml openapi.yaml
```

---

## 🚀 5. Quickstart & Verification Commands

```bash
# 1. Generate API transport code from openapi.yaml
make generate

# 2. Run unit, integration, and Testcontainers E2E test suite
make test

# 3. Build program binary
make build

# 4. Start local infrastructure (Postgres 16 & Mailpit)
make docker/up
```
