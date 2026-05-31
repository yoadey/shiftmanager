package handler

import "net/http"

// OpenAPIHandler serves the OpenAPI 3.0 document and a minimal Swagger-UI page
// (T-002). Both routes are unauthenticated.
type OpenAPIHandler struct{}

// NewOpenAPIHandler creates a new OpenAPIHandler.
func NewOpenAPIHandler() *OpenAPIHandler { return &OpenAPIHandler{} }

// Spec serves the OpenAPI 3.0 JSON document.
// GET /api/v1/openapi.json
func (h *OpenAPIHandler) Spec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(openAPISpec))
}

// Docs serves a minimal Swagger-UI HTML page that loads the spec.
// GET /api/v1/docs
func (h *OpenAPIHandler) Docs(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(swaggerHTML))
}

const swaggerHTML = `<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>ShiftManager API</title>
  <link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui.css">
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://cdn.jsdelivr.net/npm/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function () {
      window.ui = SwaggerUIBundle({
        url: '/api/v1/openapi.json',
        dom_id: '#swagger-ui'
      });
    };
  </script>
</body>
</html>`

// openAPISpec is a hand-written OpenAPI 3.0 description of the main API surface.
const openAPISpec = `{
  "openapi": "3.0.3",
  "info": {
    "title": "ShiftManager API",
    "version": "1.0.0",
    "description": "REST API for the ShiftManager club volunteer-hours system."
  },
  "servers": [{"url": "/api/v1"}],
  "components": {
    "securitySchemes": {
      "bearerAuth": {"type": "http", "scheme": "bearer", "bearerFormat": "JWT"}
    }
  },
  "security": [{"bearerAuth": []}],
  "paths": {
    "/auth/login": {"get": {"summary": "Begin OIDC login", "security": [], "responses": {"302": {"description": "Redirect to IdP"}}}},
    "/auth/callback": {"get": {"summary": "OIDC callback", "security": [], "responses": {"200": {"description": "Session token"}}}},
    "/auth/me": {"get": {"summary": "Current session", "responses": {"200": {"description": "Authenticated user"}}}},
    "/auth/logout": {"post": {"summary": "Log out", "responses": {"204": {"description": "Logged out"}}}},
    "/members": {
      "get": {"summary": "List members", "responses": {"200": {"description": "Member list"}}},
      "post": {"summary": "Create member (Vorstand+)", "responses": {"201": {"description": "Created"}}}
    },
    "/members/{id}": {
      "get": {"summary": "Get member", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"200": {"description": "Member"}}},
      "put": {"summary": "Update member (Vorstand+)", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"200": {"description": "Updated"}}},
      "delete": {"summary": "Deactivate member (Vorstand+)", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"204": {"description": "Deactivated"}}}
    },
    "/members/me/preferences": {"put": {"summary": "Update own reminder opt-out", "responses": {"200": {"description": "Updated"}}}},
    "/members/{id}/export-data": {"get": {"summary": "GDPR data export (own or Vorstand+)", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"200": {"description": "Export document"}}}},
    "/members/{id}/gdpr-delete": {"post": {"summary": "GDPR anonymize member (Vorstand+)", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"200": {"description": "Anonymized"}}}},
    "/events": {
      "get": {"summary": "List events", "responses": {"200": {"description": "Events"}}},
      "post": {"summary": "Create event (Veranstaltungsleiter+)", "responses": {"201": {"description": "Created"}}}
    },
    "/events/{id}": {"get": {"summary": "Event with timeline", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"200": {"description": "Event"}}}},
    "/shifts/{id}/register": {
      "post": {"summary": "Register for a shift", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"201": {"description": "Registered"}}},
      "delete": {"summary": "Deregister from a shift", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"204": {"description": "Deregistered"}}}
    },
    "/hours/account": {"get": {"summary": "Member hour account", "responses": {"200": {"description": "Account"}}}},
    "/stats": {"get": {"summary": "System statistics (Veranstaltungsleiter+)", "responses": {"200": {"description": "Stats"}}}},
    "/billing/{clubYearId}": {"get": {"summary": "Compute year billing (Vorstand+)", "parameters": [{"name": "clubYearId", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"200": {"description": "Billing report"}}}},
    "/settings": {
      "get": {"summary": "Get application settings", "responses": {"200": {"description": "Settings"}}},
      "put": {"summary": "Update settings (Vorstand+)", "responses": {"200": {"description": "Updated"}}}
    },
    "/settings/branding": {
      "get": {"summary": "Get branding", "responses": {"200": {"description": "Branding"}}},
      "put": {"summary": "Update branding (Vorstand+); returns WCAG contrast warnings", "responses": {"200": {"description": "Branding + warnings"}}}
    },
    "/settings/logo": {"post": {"summary": "Upload logo PNG/SVG (Vorstand+)", "responses": {"200": {"description": "Branding"}}}},
    "/settings/email-templates": {"get": {"summary": "List email templates (Vorstand+)", "responses": {"200": {"description": "Templates"}}}},
    "/settings/email-templates/{name}": {
      "get": {"summary": "Get email template (Vorstand+)", "parameters": [{"name": "name", "in": "path", "required": true, "schema": {"type": "string"}}], "responses": {"200": {"description": "Template"}}},
      "put": {"summary": "Update email template (Vorstand+)", "parameters": [{"name": "name", "in": "path", "required": true, "schema": {"type": "string"}}], "responses": {"200": {"description": "Updated"}}}
    },
    "/settings/email-log": {"get": {"summary": "List email send log (Vorstand+)", "responses": {"200": {"description": "Log entries"}}}},
    "/settings/email-log/{id}/resend": {"post": {"summary": "Resend a logged email (Vorstand+)", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"200": {"description": "Resent"}}}},
    "/settings/members/{id}/fee-tiers": {
      "get": {"summary": "Get per-member fee tier overrides (Vorstand+)", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"200": {"description": "Tiers"}}},
      "put": {"summary": "Set per-member fee tier overrides (Vorstand+)", "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"200": {"description": "Updated"}}}
    },
    "/kiosk/events/{id}": {"get": {"summary": "Public event timeline", "security": [], "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"200": {"description": "Timeline"}, "403": {"description": "Kiosk locked"}}}},
    "/kiosk/shifts/{id}/register": {"post": {"summary": "Public kiosk registration", "security": [], "parameters": [{"name": "id", "in": "path", "required": true, "schema": {"type": "string", "format": "uuid"}}], "responses": {"201": {"description": "Registered"}, "403": {"description": "Kiosk locked"}}}}
  }
}`
