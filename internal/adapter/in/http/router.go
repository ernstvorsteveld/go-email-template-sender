package http

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/in/http/gen"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/in"
)

// loggingMiddleware wraps an http.Handler to log the request details
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		next.ServeHTTP(w, r)
		
		slog.Info("HTTP Request",
			slog.String("method", r.Method),
			slog.String("path", r.URL.Path),
			slog.String("duration", time.Since(start).String()),
		)
	})
}

// NewRouter wires up all generated HTTP endpoints to their respective UseCase driving ports.
func NewRouter(
	contextUC in.ContextUseCase,
	stylesheetUC in.StylesheetUseCase,
	templateUC in.TemplateUseCase,
	bindingUC in.BindingUseCase,
	deliveryUC in.DeliveryUseCase,
) http.Handler {
	server := NewServer(contextUC, stylesheetUC, templateUC, bindingUC, deliveryUC)
	handler := gen.Handler(server)
	return loggingMiddleware(handler)
}
