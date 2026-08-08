package http

import (
	"encoding/json"
	"net/http"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/in"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
)

type TemplateHandler struct {
	useCase in.TemplateUseCase
}

func NewTemplateHandler(uc in.TemplateUseCase) *TemplateHandler {
	return &TemplateHandler{useCase: uc}
}

type CreateTemplateRequest struct {
	Name         string  `json:"name"`
	Code         string  `json:"code"`
	HTMLContent  string  `json:"html_content"`
	StylesheetID *string `json:"stylesheet_id"`
	Subject      string  `json:"subject"`
}

type TemplateResponse struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	Code         string  `json:"code"`
	Version      int     `json:"version"`
	StylesheetID *string `json:"stylesheet_id"`
	HTMLContent  string  `json:"html_content"`
	Subject      string  `json:"subject"`
}

func mapToTemplateResponse(t domain.Template) TemplateResponse {
	var sID *string
	if t.Stylesheet != nil {
		val := uuid.UUID(*t.Stylesheet).String()
		sID = &val
	}
	return TemplateResponse{
		ID:           uuid.UUID(t.ID).String(),
		Name:         string(t.Name),
		Code:         string(t.Code),
		Version:      int(t.Version),
		StylesheetID: sID,
		HTMLContent:  string(t.Content),
		Subject:      string(t.Subject),
	}
}

func (h *TemplateHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var sID *domain.IDType
	if req.StylesheetID != nil {
		parsed, err := uuid.Parse(*req.StylesheetID)
		if err != nil {
			http.Error(w, "invalid stylesheet_id", http.StatusBadRequest)
			return
		}
		id := domain.IDType(parsed)
		sID = &id
	}

	id, err := h.useCase.CreateTemplate(r.Context(), domain.NameType(req.Name), domain.CodeType(req.Code), domain.HTMLType(req.HTMLContent), sID, domain.SubjectType(req.Subject))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"id": uuid.UUID(id).String()})
}

func (h *TemplateHandler) List(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	items, err := h.useCase.GetTemplates(r.Context(), domain.NameType(name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var response []TemplateResponse
	for _, item := range items {
		response = append(response, mapToTemplateResponse(item))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TemplateHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	item, err := h.useCase.GetTemplate(r.Context(), domain.IDType(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mapToTemplateResponse(item))
}

func (h *TemplateHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	var req CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var sID *domain.IDType
	if req.StylesheetID != nil {
		parsed, err := uuid.Parse(*req.StylesheetID)
		if err != nil {
			http.Error(w, "invalid stylesheet_id", http.StatusBadRequest)
			return
		}
		pID := domain.IDType(parsed)
		sID = &pID
	}

	err = h.useCase.UpdateTemplate(r.Context(), domain.IDType(id), domain.NameType(req.Name), domain.CodeType(req.Code), domain.HTMLType(req.HTMLContent), sID, domain.SubjectType(req.Subject))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TemplateHandler) Render(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	html, err := h.useCase.RenderTemplate(r.Context(), domain.IDType(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
