Here is the updated prompt with the collection-level `GET` endpoint for searching Stylesheets added.

---

**"Act as a senior Golang software architect. I need to design a web API using a strict Hexagonal Architecture (Ports and Adapters) and Domain-Driven Design (DDD) principles. The API acts as a dynamic templating, styling, data-binding, and email delivery service, backed by PostgreSQL.**

**Please define the core domain models, inbound/outbound ports (interfaces), application services (use cases), and adapters (HTTP handlers, Postgres repositories) for the following resources:**

**1. Context Data (Arbitrary JSON Store)**

* **`POST /contexts`**: Accepts an external `reference_id` (string), a `customer_name` (string), a `payload` (string) containing arbitrary JSON, and an `email_jsonpath` (a string defining the JSONPath to extract the recipient email address from the payload). Stores this in PostgreSQL. Returns a newly generated internal `id` (UUID).
* **`GET /contexts`**: Retrieves a list of Context entities. Must accept an optional query parameter `customer_name` to filter the results.
* **`GET /contexts/{id}`**: Retrieves the Context entity as JSON, including the `reference_id`, `customer_name`, `email_jsonpath`, `id` and the arbitrary `payload`.
* **`PUT /contexts/{id}`**: Fully replaces the existing JSON payload, reference, `customer_name`, and `email_jsonpath` for the given internal ID.

**2. Stylesheets (CSS Store)**

* **`POST /stylesheets`**: Accepts a `name` (string), a unique `code` (string), and the raw `css_content` string. Stores it in the database and returns an internal `id` (UUID).
* **`GET /stylesheets`**: Retrieves a list of Stylesheet entities. Must accept an optional query parameter `name` to filter the results.
* **`GET /stylesheets/{id}`**: Retrieves the Stylesheet metadata (id, name, code) and the `css_content` as a standard JSON response.
* **`PUT /stylesheets/{id}`**: Fully replaces the CSS content and metadata for the given internal ID.

**3. Document Templates (HTML Store & Rendering)**

* **`POST /templates`**: Accepts a `name` (string), a unique `code` (string), the raw `html_content` string (which may contain Handlebars-style variables), and an optional `stylesheet_id` (foreign key to the Stylesheet). Stores it in the database and returns an internal `id` (UUID). Include a `version` integer that starts at 1.
* **`GET /templates`**: Retrieves a list of Template entities. Must accept an optional query parameter `name` to filter the results.
* **`GET /templates/{id}`**: Retrieves the Template metadata as JSON, including `id`, `name`, `code`, `version`, `stylesheet_id`, and the raw `html_content`.
* **`PUT /templates/{id}`**: Updates the HTML document and/or the linked `stylesheet_id`, and increments its version.
* **`GET /templates/{id}/render`**: Retrieves the HTML template. If a `stylesheet_id` is linked, the application service must use the `[github.com/PuerkitoBio/goquery](https://github.com/PuerkitoBio/goquery)` library to parse the HTML string and safely append a `<link>` tag to the `<head>` section before returning the final HTML string.

**4. Data Source Bindings (Secure Query Definitions)**

* **`POST /bindings`**: Accepts a `name` (string), query` (a select statement that is used to select context objects as string), and a `template_id` to link the data to a specific HTML document. Returns a binding `id`.
* **`GET /bindings`**: Retrieves a list of Binding entities. Must accept an optional query parameter `name` to filter the results.
* **`GET /bindings/{id}`**: Retrieves the Binding details as JSON, including the `id`, `query`, and `template_id`.
* **`PUT /bindings/{id}`**: Updates the complete Binding object.

**5. Delivery & Dispatch (Template Merging & Emailing)**

* **`POST /deliveries`**: Accepts a `template_id` and a `binding_id`. The application service orchestrating this must perform the following workflow:
1. Resolve the Context data using the Binding.
2. Extract the destination email address from the Context's JSON payload using the Context's stored `email_jsonpath`.
3. Retrieve the HTML Template (including the goquery-injected CSS link).
4. Merge the Context JSON payload into the HTML template using a templating engine (e.g., standard `html/template` or a Handlebars library like `aymerick/raymond`).
5. Dispatch the final rendered HTML to the extracted email address via an outbound `EmailSender` port.



**Architectural & Technical Requirements:**

* **Hexagonal Strictness:** Explicitly separate the code into `domain` (entities/aggregates), `application` (use cases), `ports` (interfaces for driving and driven actors), and `adapters` (HTTP transport, Postgres storage, Email sender).
* **Boundary Isolation:** HTTP request/response structs must not bleed into the domain models. SQL queries and database tags (`db:""`) must be contained entirely within the repository adapters, mapping to pure domain entities before returning.
* Use `pgx` for PostgreSQL interactions.
* Use standard Go idioms and a lightweight router (like `net/http` in Go 1.22+ or `chi`).
* Ensure the design prevents SQL injection by utilizing the `query_identifier` approach rather than accepting raw SQL strings."

---
