# Project Architecture & Master Implementation Plan: `go-email-template-sender`

> **Overview**: `go-email-template-sender` is a dynamic email templating, CSS styling, arbitrary JSON context data-binding, and email delivery orchestration service written in Golang (Go 1.22+). It is designed adhering strictly to **Hexagonal Architecture (Ports and Adapters)** and **Domain-Driven Design (DDD)** principles, backed by PostgreSQL and Mailpit/SMTP.

---

## 🏛️ 1. Architecture & Design Principles

### A. Hexagonal Architecture (Ports and Adapters)
The codebase enforces strict layer isolation:
* **Domain Layer (`internal/domain`)**: Pure Go domain entities (`Context`, `Stylesheet`, `Template`, `Binding`) free from database tags or HTTP annotations.
* **Inbound Driving Ports (`internal/application/port/in`)**: Go interface contracts defining all application use cases.
* **Outbound Driven Ports (`internal/application/port/out`)**: Go interface contracts defining persistence and email delivery capabilities.
* **Application Services (`internal/application/service`)**: Pure use case orchestrators executing business rules,Handlebars template rendering (`aymerick/raymond`), and CSS link injection (`PuerkitoBio/goquery`).
* **Inbound Driving Adapters (`internal/adapter/in/http`)**: REST HTTP transport powered by Go 1.22 `net/http` standard library router. Implements OpenAPI-generated `gen.ServerInterface`.
* **Outbound Driven Adapters**:
  * **PostgreSQL (`internal/adapter/out/postgres`)**: SQL persistence utilizing `pgx/v5` pool (`pgxpool`), supporting native PostgreSQL `JSONB` querying operators (`payload->>'field'`).
  * **Email Sender (`internal/adapter/out/email`)**: SMTP client connected to Mailpit.

---

## 🗄️ 2. Database Schema (`schema.sql`)

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

## 📄 3. OpenAPI 3.0 Specification (`openapi.yaml`)

```yaml
openapi: 3.0.3
info:
  title: Go Email Template Sender API
  version: 1.0.0
  description: A dynamic templating, styling, data-binding, and email delivery service built with Hexagonal Architecture.
servers:
  - url: http://localhost:8080
    description: Local development server

tags:
  - name: Contexts
    description: Arbitrary JSON data context store
  - name: Stylesheets
    description: CSS stylesheet management
  - name: Templates
    description: Dynamic HTML document templates
  - name: Bindings
    description: Data source query bindings
  - name: Deliveries
    description: Email delivery and dispatch orchestrator

paths:
  /contexts:
    post:
      tags: [Contexts]
      summary: Create a new Context
      operationId: createContext
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateContextRequest'
      responses:
        '201':
          description: Context created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/IdResponse'
    get:
      tags: [Contexts]
      summary: List Contexts
      operationId: listContexts
      parameters:
        - name: customer_name
          in: query
          required: false
          schema:
            type: string
      responses:
        '200':
          description: A list of Contexts
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/ContextResponse'

  /contexts/{id}:
    get:
      tags: [Contexts]
      summary: Get a Context by ID
      operationId: getContext
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Context found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/ContextResponse'
        '404':
          description: Context not found
    put:
      tags: [Contexts]
      summary: Fully replace an existing Context
      operationId: updateContext
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateContextRequest'
      responses:
        '204':
          description: Context updated successfully

  /stylesheets:
    post:
      tags: [Stylesheets]
      summary: Create a new Stylesheet
      operationId: createStylesheet
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateStylesheetRequest'
      responses:
        '201':
          description: Stylesheet created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/IdResponse'
    get:
      tags: [Stylesheets]
      summary: List Stylesheets
      operationId: listStylesheets
      parameters:
        - name: name
          in: query
          required: false
          schema:
            type: string
      responses:
        '200':
          description: A list of Stylesheets
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/StylesheetResponse'

  /stylesheets/{id}:
    get:
      tags: [Stylesheets]
      summary: Get Stylesheet by ID
      operationId: getStylesheet
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Stylesheet found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/StylesheetResponse'
    put:
      tags: [Stylesheets]
      summary: Fully replace an existing Stylesheet
      operationId: updateStylesheet
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateStylesheetRequest'
      responses:
        '204':
          description: Stylesheet updated

  /templates:
    post:
      tags: [Templates]
      summary: Create a new Template
      operationId: createTemplate
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateTemplateRequest'
      responses:
        '201':
          description: Template created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/IdResponse'
    get:
      tags: [Templates]
      summary: List Templates
      operationId: listTemplates
      parameters:
        - name: name
          in: query
          required: false
          schema:
            type: string
      responses:
        '200':
          description: A list of Templates
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/TemplateResponse'

  /templates/{id}:
    get:
      tags: [Templates]
      summary: Get Template by ID
      operationId: getTemplate
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Template found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/TemplateResponse'
    put:
      tags: [Templates]
      summary: Update Template document
      operationId: updateTemplate
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateTemplateRequest'
      responses:
        '204':
          description: Template updated

  /templates/{id}/render:
    get:
      tags: [Templates]
      summary: Render a Template with linked CSS injected
      operationId: renderTemplate
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Rendered HTML document
          content:
            text/html:
              schema:
                type: string

  /bindings:
    post:
      tags: [Bindings]
      summary: Create a new Binding
      operationId: createBinding
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateBindingRequest'
      responses:
        '201':
          description: Binding created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/IdResponse'
    get:
      tags: [Bindings]
      summary: List Bindings
      operationId: listBindings
      parameters:
        - name: name
          in: query
          required: false
          schema:
            type: string
      responses:
        '200':
          description: A list of Bindings
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/BindingResponse'

  /bindings/{id}:
    get:
      tags: [Bindings]
      summary: Get Binding by ID
      operationId: getBinding
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      responses:
        '200':
          description: Binding found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/BindingResponse'
    put:
      tags: [Bindings]
      summary: Fully replace an existing Binding
      operationId: updateBinding
      parameters:
        - name: id
          in: path
          required: true
          schema:
            type: string
            format: uuid
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateBindingRequest'
      responses:
        '204':
          description: Binding updated

  /deliveries:
    post:
      tags: [Deliveries]
      summary: Trigger email delivery dispatch
      operationId: createDelivery
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateDeliveryRequest'
      responses:
        '202':
          description: Delivery dispatch accepted
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/DeliveryResponse'

components:
  schemas:
    IdResponse:
      type: object
      required: [id]
      properties:
        id:
          type: string
          format: uuid

    CreateContextRequest:
      type: object
      required: [reference_id, customer_name, payload, email_jsonpath]
      properties:
        reference_id:
          type: string
        customer_name:
          type: string
        payload:
          type: string
        email_jsonpath:
          type: string

    ContextResponse:
      type: object
      required: [id, reference_id, customer_name, payload, email_jsonpath]
      properties:
        id:
          type: string
          format: uuid
        reference_id:
          type: string
        customer_name:
          type: string
        payload:
          type: string
        email_jsonpath:
          type: string

    CreateStylesheetRequest:
      type: object
      required: [name, code, css_content]
      properties:
        name:
          type: string
        code:
          type: string
        css_content:
          type: string

    StylesheetResponse:
      type: object
      required: [id, name, code, css_content]
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        code:
          type: string
        css_content:
          type: string

    CreateTemplateRequest:
      type: object
      required: [name, code, html_content, subject]
      properties:
        name:
          type: string
        code:
          type: string
        html_content:
          type: string
        subject:
          type: string
        stylesheet_id:
          type: string
          format: uuid
          nullable: true

    TemplateResponse:
      type: object
      required: [id, name, code, version, html_content, subject, stylesheet_id]
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        code:
          type: string
        version:
          type: integer
        html_content:
          type: string
        subject:
          type: string
        stylesheet_id:
          type: string
          format: uuid
          nullable: true

    CreateBindingRequest:
      type: object
      required: [name, query, template_id]
      properties:
        name:
          type: string
        query:
          type: string
        template_id:
          type: string
          format: uuid

    BindingResponse:
      type: object
      required: [id, name, query, template_id]
      properties:
        id:
          type: string
          format: uuid
        name:
          type: string
        query:
          type: string
        template_id:
          type: string
          format: uuid

    CreateDeliveryRequest:
      type: object
      required: [template_id, binding_id]
      properties:
        template_id:
          type: string
          format: uuid
        binding_id:
          type: string
          format: uuid

    DeliveryResponse:
      type: object
      required: [status]
      properties:
        status:
          type: string
          example: "dispatched"
```

---

## 🛠️ 4. Code Generation & Subpackage Isolation

### Generator Configuration (`oapi-codegen.yaml`)
The generated transport code is isolated into a dedicated `gen/` subpackage:

```yaml
package: gen
output: internal/adapter/in/http/gen/api.gen.go
generate:
  models: true
  std-http-server: true
```

### Git Exclusion (`.gitignore`)
The generated code is excluded from version control so it can be generated on demand:

```gitignore
# Binaries
bin/
/server

# Generated OpenAPI code
*.gen.go
internal/adapter/in/http/gen/
```

### Code Generation Target (`Makefile`)
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
