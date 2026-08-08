package http

import (
	"encoding/json"
	"net/http"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/in"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
)

type StylesheetHandler struct {
	useCase in.StylesheetUseCase
}

func NewStylesheetHandler(uc in.StylesheetUseCase) *StylesheetHandler {
	return &StylesheetHandler{useCase: uc}
}

type CreateStylesheetRequest struct {
	Name       string `json:"name"`
	Code       string `json:"code"`
	CSSContent string `json:"css_content"`
}

type StylesheetResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Code       string `json:"code"`
	CSSContent string `json:"css_content"`
}

func mapToStylesheetResponse(s domain.Stylesheet) StylesheetResponse {
	return StylesheetResponse{
		ID:         uuid.UUID(s.ID).String(),
		Name:       string(s.Name),
		Code:       string(s.Code),
		CSSContent: string(s.Content),
	}
}

func (h *StylesheetHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateStylesheetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := h.useCase.CreateStylesheet(r.Context(), domain.NameType(req.Name), domain.CodeType(req.Code), domain.CSSType(req.CSSContent))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": uuid.UUID(id).String()})
}

func (h *StylesheetHandler) List(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	
	items, err := h.useCase.GetStylesheets(r.Context(), domain.NameType(name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []StylesheetResponse
	for _, item := range items {
		response = append(response, mapToStylesheetResponse(item))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *StylesheetHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	item, err := h.useCase.GetStylesheet(r.Context(), domain.IDType(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapToStylesheetResponse(item))
}

func (h *StylesheetHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id format", http.StatusBadRequest)
		return
	}

	var req CreateStylesheetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = h.useCase.UpdateStylesheet(r.Context(), domain.IDType(id), domain.NameType(req.Name), domain.CodeType(req.Code), domain.CSSType(req.CSSContent))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
