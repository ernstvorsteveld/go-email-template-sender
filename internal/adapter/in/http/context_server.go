package http

import (
	"encoding/json"
	"net/http"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/in/http/gen"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

func (s *Server) CreateContext(w http.ResponseWriter, r *http.Request) {
	var req gen.CreateContextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	id, err := s.contextUC.CreateContext(
		r.Context(),
		domain.ReferenceType(req.ReferenceId),
		domain.CustomerType(req.CustomerName),
		domain.JSONPayloadType(req.Payload),
		domain.JSONPathType(req.EmailJsonpath),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(gen.IdResponse{Id: uuid.UUID(id)})
}

func (s *Server) ListContexts(w http.ResponseWriter, r *http.Request, params gen.ListContextsParams) {
	custName := ""
	if params.CustomerName != nil {
		custName = *params.CustomerName
	}

	contexts, err := s.contextUC.GetContexts(r.Context(), domain.CustomerType(custName))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := make([]gen.ContextResponse, 0, len(contexts))
	for _, c := range contexts {
		response = append(response, gen.ContextResponse{
			Id:            uuid.UUID(c.ID),
			ReferenceId:   string(c.Reference),
			CustomerName:  string(c.Customer),
			Payload:       string(c.Payload),
			EmailJsonpath: string(c.EmailAddress),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) GetContext(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	c, err := s.contextUC.GetContext(r.Context(), domain.IDType(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	resp := gen.ContextResponse{
		Id:            uuid.UUID(c.ID),
		ReferenceId:   string(c.Reference),
		CustomerName:  string(c.Customer),
		Payload:       string(c.Payload),
		EmailJsonpath: string(c.EmailAddress),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) UpdateContext(w http.ResponseWriter, r *http.Request, id openapi_types.UUID) {
	var req gen.CreateContextRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := s.contextUC.UpdateContext(
		r.Context(),
		domain.IDType(id),
		domain.ReferenceType(req.ReferenceId),
		domain.CustomerType(req.CustomerName),
		domain.JSONPayloadType(req.Payload),
		domain.JSONPathType(req.EmailJsonpath),
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
