package handler

import (
	_ "embed"
	"net/http"
)

//go:embed spec/openapi.yaml
var openAPISpecYAML []byte

// OpenAPIHandler serves the canonical OpenAPI 3.0 spec (api/openapi.yaml) and
// a Swagger-UI docs page. Both routes are unauthenticated.
type OpenAPIHandler struct{}

// NewOpenAPIHandler creates a new OpenAPIHandler.
func NewOpenAPIHandler() *OpenAPIHandler { return &OpenAPIHandler{} }

// Spec serves the OpenAPI 3.0 YAML document.
// GET /api/v1/openapi.yaml  (also routed from /openapi.json for compat)
func (h *OpenAPIHandler) Spec(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/yaml")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(openAPISpecYAML)
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
