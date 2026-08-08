package http

import (
	"encoding/json"
	"net/http"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/in/http/gen"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (s *Server) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var req gen.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var sID *domain.IDType
	if req.StylesheetId != nil {
		id := domain.IDType(*req.StylesheetId)
		sID = &id
	}

	id, err := s.templateUC.CreateTemplate(
		r.Context(),
		domain.NameType(req.Name),
		domain.CodeType(req.Code),
		domain.HTMLType(req.HtmlContent),
		sID,
		domain.SubjectType(req.Subject),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(gen.IdResponse{Id: uuid.UUID(id)})
}

func (s *Server) ListTemplates(w http.ResponseWriter, r *http.Request, params gen.ListTemplatesParams) {
	name := ""
	if params.Name != nil {
		name = *params.Name
	}

	items, err := s.templateUC.GetTemplates(r.Context(), domain.NameType(name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]gen.TemplateResponse, 0, len(items))
	for _, item := range items {
		var sID *uuid.UUID
		if item.Stylesheet != nil {
			uid := uuid.UUID(*item.Stylesheet)
			sID = &uid
		}
		response = append(response, gen.TemplateResponse{
			Id:           uuid.UUID(item.ID),
			Name:         string(item.Name),
			Code:         string(item.Code),
			Version:      int(item.Version),
			StylesheetId: sID,
			HtmlContent:  string(item.Content),
			Subject:      string(item.Subject),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) GetTemplate(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	item, err := s.templateUC.GetTemplate(r.Context(), domain.IDType(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	var sID *uuid.UUID
	if item.Stylesheet != nil {
		uid := uuid.UUID(*item.Stylesheet)
		sID = &uid
	}

	resp := gen.TemplateResponse{
		Id:           uuid.UUID(item.ID),
		Name:         string(item.Name),
		Code:         string(item.Code),
		Version:      int(item.Version),
		StylesheetId: sID,
		HtmlContent:  string(item.Content),
		Subject:      string(item.Subject),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) UpdateTemplate(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	var req gen.CreateTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var sID *domain.IDType
	if req.StylesheetId != nil {
		pID := domain.IDType(*req.StylesheetId)
		sID = &pID
	}

	err := s.templateUC.UpdateTemplate(
		r.Context(),
		domain.IDType(id),
		domain.NameType(req.Name),
		domain.CodeType(req.Code),
		domain.HTMLType(req.HtmlContent),
		sID,
		domain.SubjectType(req.Subject),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) RenderTemplate(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	html, err := s.templateUC.RenderTemplate(r.Context(), domain.IDType(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}
