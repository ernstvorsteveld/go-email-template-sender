package http

import (
	"encoding/json"
	"net/http"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/in"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
)

type ContextHandler struct {
	useCase in.ContextUseCase
}

func NewContextHandler(uc in.ContextUseCase) *ContextHandler {
	return &ContextHandler{useCase: uc}
}

// Request and Response DTOs isolate HTTP concerns from Domain models
type CreateContextRequest struct {
	ReferenceID   string `json:"reference_id"`
	CustomerName  string `json:"customer_name"`
	Payload       string `json:"payload"`
	EmailJSONPath string `json:"email_jsonpath"`
}

type ContextResponse struct {
	ID            string `json:"id"`
	ReferenceID   string `json:"reference_id"`
	CustomerName  string `json:"customer_name"`
	Payload       string `json:"payload"`
	EmailJSONPath string `json:"email_jsonpath"`
}

func mapToContextResponse(c domain.Context) ContextResponse {
	return ContextResponse{
		ID:            uuid.UUID(c.ID).String(),
		ReferenceID:   string(c.Reference),
		CustomerName:  string(c.Customer),
		Payload:       string(c.Payload),
		EmailJSONPath: string(c.EmailAddress),
	}
}

func (h *ContextHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateContextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.useCase.CreateContext(
		r.Context(),
		domain.ReferenceType(req.ReferenceID),
		domain.CustomerType(req.CustomerName),
		domain.JSONPayloadType(req.Payload),
		domain.JSONPathType(req.EmailJSONPath),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": uuid.UUID(id).String()})
}

func (h *ContextHandler) List(w http.ResponseWriter, r *http.Request) {
	customerName := r.URL.Query().Get("customer_name")
	
	contexts, err := h.useCase.GetContexts(r.Context(), domain.CustomerType(customerName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []ContextResponse
	for _, c := range contexts {
		response = append(response, mapToContextResponse(c))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ContextHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}

	c, err := h.useCase.GetContext(r.Context(), domain.IDType(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapToContextResponse(c))
}

func (h *ContextHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}

	var req CreateContextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = h.useCase.UpdateContext(
		r.Context(),
		domain.IDType(id),
		domain.ReferenceType(req.ReferenceID),
		domain.CustomerType(req.CustomerName),
		domain.JSONPayloadType(req.Payload),
		domain.JSONPathType(req.EmailJSONPath),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
