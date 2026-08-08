package http

import (
	"log/slog"
	"net/http"
	"time"

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

// NewRouter wires up all HTTP endpoints to their respective UseCase driving ports.
func NewRouter(
	contextUC in.ContextUseCase,
	stylesheetUC in.StylesheetUseCase,
	templateUC in.TemplateUseCase,
	bindingUC in.BindingUseCase,
	deliveryUC in.DeliveryUseCase,
) http.Handler {
	mux := http.NewServeMux()

	// Contexts
	if contextUC != nil {
		ctxHandler := NewContextHandler(contextUC)
		mux.HandleFunc("POST /contexts", ctxHandler.Create)
		mux.HandleFunc("GET /contexts", ctxHandler.List)
		mux.HandleFunc("GET /contexts/{id}", ctxHandler.Get)
		mux.HandleFunc("PUT /contexts/{id}", ctxHandler.Update)
	}

	// Deliveries
	if deliveryUC != nil {
		deliveryHandler := NewDeliveryHandler(deliveryUC)
		mux.HandleFunc("POST /deliveries", deliveryHandler.Create)
	}

	// Stylesheets
	if stylesheetUC != nil {
		ssHandler := NewStylesheetHandler(stylesheetUC)
		mux.HandleFunc("POST /stylesheets", ssHandler.Create)
		mux.HandleFunc("GET /stylesheets", ssHandler.List)
		mux.HandleFunc("GET /stylesheets/{id}", ssHandler.Get)
		mux.HandleFunc("PUT /stylesheets/{id}", ssHandler.Update)
	}

	// Templates
	if templateUC != nil {
		tmplHandler := NewTemplateHandler(templateUC)
		mux.HandleFunc("POST /templates", tmplHandler.Create)
		mux.HandleFunc("GET /templates", tmplHandler.List)
		mux.HandleFunc("GET /templates/{id}", tmplHandler.Get)
		mux.HandleFunc("PUT /templates/{id}", tmplHandler.Update)
		mux.HandleFunc("GET /templates/{id}/render", tmplHandler.Render)
	}

	// Bindings
	if bindingUC != nil {
		bHandler := NewBindingHandler(bindingUC)
		mux.HandleFunc("POST /bindings", bHandler.Create)
		mux.HandleFunc("GET /bindings", bHandler.List)
		mux.HandleFunc("GET /bindings/{id}", bHandler.Get)
		mux.HandleFunc("PUT /bindings/{id}", bHandler.Update)
	}

	return loggingMiddleware(mux)
}
