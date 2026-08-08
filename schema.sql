CREATE TABLE IF NOT EXISTS contexts (
    id UUID PRIMARY KEY,
    reference_id VARCHAR(255) NOT NULL,
    customer_name VARCHAR(255) NOT NULL,
    payload JSONB NOT NULL,
    email_jsonpath VARCHAR(255) NOT NULL
);

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
