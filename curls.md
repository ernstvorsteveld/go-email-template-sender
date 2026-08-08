## Curl statements

### Context

**Create Context**
```bash
curl --request POST \
  --url http://localhost:8180/contexts \
  --header 'Content-Type: application/json' \
  --data '{
    "reference_id": "ORDER-1001",
    "customer_name": "Acme Corp",
    "email_jsonpath": "$.recipient.email",
    "payload": "{\"recipient\": {\"email\": \"test@example.com\", \"name\": \"Alice\"}, \"order_total\": \"120.50\"}"
}'
```

**Get Context**
```bash
curl --request GET \
  --url http://localhost:8180/contexts/217cef25-1c26-4ece-ae65-665dcaf09251
```

### Stylesheet

**Create Stylesheet**
```bash
curl --request POST \
  --url http://localhost:8180/stylesheets/a46cc6e0-605c-41b8-bf0e-6b42341f27f3 \
  --header 'Content-Type: application/json' \
  --data '{
    "name": "Acme Default Theme",
    "code": "ACME_DEFAULT",
    "css_content": "body { font-family: Arial, sans-serif; } h1 { color: #0055ff; }"
}'
```

**Get Stylesheet**
```bash
curl --request GET \
  --url http://localhost:8180/stylesheets/a46cc6e0-605c-41b8-bf0e-6b42341f27f3
```
