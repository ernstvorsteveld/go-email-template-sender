package http

import (
	"encoding/json"
	"net/http"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/in"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
)

type BindingHandler struct {
	useCase in.BindingUseCase
}

func NewBindingHandler(uc in.BindingUseCase) *BindingHandler {
	return &BindingHandler{useCase: uc}
}

type CreateBindingRequest struct {
	Name       string `json:"name"`
	Query      string `json:"query"`
	TemplateID string `json:"template_id"`
}

type BindingResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Query      string `json:"query"`
	TemplateID string `json:"template_id"`
}

func mapToBindingResponse(b domain.Binding) BindingResponse {
	return BindingResponse{
		ID:         uuid.UUID(b.ID).String(),
		Name:       string(b.Name),
		Query:      string(b.Query),
		TemplateID: uuid.UUID(b.Template).String(),
	}
}

func (h *BindingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateBindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		http.Error(w, "invalid template_id", http.StatusBadRequest)
		return
	}

	id, err := h.useCase.CreateBinding(r.Context(), domain.NameType(req.Name), domain.SQLQueryType(req.Query), domain.IDType(tID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": uuid.UUID(id).String()})
}

func (h *BindingHandler) List(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	items, err := h.useCase.GetBindings(r.Context(), domain.NameType(name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []BindingResponse
	for _, item := range items {
		response = append(response, mapToBindingResponse(item))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *BindingHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	item, err := h.useCase.GetBinding(r.Context(), domain.IDType(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapToBindingResponse(item))
}

func (h *BindingHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}

	var req CreateBindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	tID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		http.Error(w, "invalid template_id", http.StatusBadRequest)
		return
	}

	err = h.useCase.UpdateBinding(r.Context(), domain.IDType(id), domain.NameType(req.Name), domain.SQLQueryType(req.Query), domain.IDType(tID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
