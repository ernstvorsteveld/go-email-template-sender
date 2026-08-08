### 1. Create a Context (Data Payload)
POST http://localhost:8080/contexts
Content-Type: application/json

{
    "reference_id": "ORDER-1001",
    "customer_name": "Acme Corp",
    "email_jsonpath": "$.recipient.email",
    "payload": "{\"recipient\": {\"email\": \"test@example.com\", \"name\": \"Alice\"}, \"order_total\": \"120.50\"}"
}

### 2. List Contexts by Customer
GET http://localhost:8080/contexts?customer_name=Acme%20Corp


### 3. Trigger Delivery Orchestrator
# Note: You will need to replace the UUIDs below with the actual IDs
# of a Template and a Binding once they are created in the database.
POST http://localhost:8080/deliveries
Content-Type: application/json

{
    "template_id": "00000000-0000-0000-0000-000000000000",
    "binding_id": "00000000-0000-0000-0000-000000000000"
}


# ====================================================================
# NOTE: The endpoints below are scaffolded in the Application Services 
# but need to be explicitly wired up in `internal/adapter/in/http/router.go` 
# before they will respond to requests.
# ====================================================================

### Create a Stylesheet
POST http://localhost:8080/stylesheets
Content-Type: application/json

{
    "name": "Acme Default Theme",
    "code": "ACME_DEFAULT",
    "css_content": "body { font-family: Arial, sans-serif; } h1 { color: #0055ff; }"
}

### Create a Template
POST http://localhost:8080/templates
Content-Type: application/json

{
    "name": "Order Confirmation",
    "code": "ORDER_CONF_01",
    "html_content": "<h1>Hello {{recipient.name}}</h1><p>Thank you for your order! Your total is ${{order_total}}.</p>",
    "stylesheet_id": null
}

### Create a Binding
POST http://localhost:8080/bindings
Content-Type: application/json

{
    "name": "Acme Order Contexts",
    "query": "SELECT * FROM contexts WHERE customer_name = 'Acme Corp'",
    "template_id": "00000000-0000-0000-0000-000000000000"
}
