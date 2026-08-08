package http

import (
	"encoding/json"
	"net/http"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/in/http/gen"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (s *Server) CreateStylesheet(w http.ResponseWriter, r *http.Request) {
	var req gen.CreateStylesheetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.stylesheetUC.CreateStylesheet(
		r.Context(),
		domain.NameType(req.Name),
		domain.CodeType(req.Code),
		domain.CSSType(req.CssContent),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(gen.IdResponse{Id: uuid.UUID(id)})
}

func (s *Server) ListStylesheets(w http.ResponseWriter, r *http.Request, params gen.ListStylesheetsParams) {
	name := ""
	if params.Name != nil {
		name = *params.Name
	}

	items, err := s.stylesheetUC.GetStylesheets(r.Context(), domain.NameType(name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]gen.StylesheetResponse, 0, len(items))
	for _, item := range items {
		response = append(response, gen.StylesheetResponse{
			Id:         uuid.UUID(item.ID),
			Name:       string(item.Name),
			Code:       string(item.Code),
			CssContent: string(item.Content),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) GetStylesheet(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	item, err := s.stylesheetUC.GetStylesheet(r.Context(), domain.IDType(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := gen.StylesheetResponse{
		Id:         uuid.UUID(item.ID),
		Name:       string(item.Name),
		Code:       string(item.Code),
		CssContent: string(item.Content),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) UpdateStylesheet(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	var req gen.CreateStylesheetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := s.stylesheetUC.UpdateStylesheet(
		r.Context(),
		domain.IDType(id),
		domain.NameType(req.Name),
		domain.CodeType(req.Code),
		domain.CSSType(req.CssContent),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
