package http

import (
	"encoding/json"
	"net/http"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/in/http/gen"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (s *Server) CreateBinding(w http.ResponseWriter, r *http.Request) {
	var req gen.CreateBindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := s.bindingUC.CreateBinding(
		r.Context(),
		domain.NameType(req.Name),
		domain.SQLQueryType(req.Query),
		domain.IDType(req.TemplateId),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(gen.IdResponse{Id: uuid.UUID(id)})
}

func (s *Server) ListBindings(w http.ResponseWriter, r *http.Request, params gen.ListBindingsParams) {
	name := ""
	if params.Name != nil {
		name = *params.Name
	}

	items, err := s.bindingUC.GetBindings(r.Context(), domain.NameType(name))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]gen.BindingResponse, 0, len(items))
	for _, item := range items {
		response = append(response, gen.BindingResponse{
			Id:         uuid.UUID(item.ID),
			Name:       string(item.Name),
			Query:      string(item.Query),
			TemplateId: uuid.UUID(item.Template),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) GetBinding(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	item, err := s.bindingUC.GetBinding(r.Context(), domain.IDType(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := gen.BindingResponse{
		Id:         uuid.UUID(item.ID),
		Name:       string(item.Name),
		Query:      string(item.Query),
		TemplateId: uuid.UUID(item.Template),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) UpdateBinding(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	var req gen.CreateBindingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := s.bindingUC.UpdateBinding(
		r.Context(),
		domain.IDType(id),
		domain.NameType(req.Name),
		domain.SQLQueryType(req.Query),
		domain.IDType(req.TemplateId),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
