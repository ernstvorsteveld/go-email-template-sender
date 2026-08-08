package http

import (
	"encoding/json"
	"net/http"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/in/http/gen"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/in"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

// Server implements the generated gen.ServerInterface contract.
type Server struct {
	contextUC    in.ContextUseCase
	stylesheetUC in.StylesheetUseCase
	templateUC   in.TemplateUseCase
	bindingUC    in.BindingUseCase
	deliveryUC   in.DeliveryUseCase
}

var _ gen.ServerInterface = (*Server)(nil)

func NewServer(
	contextUC in.ContextUseCase,
	stylesheetUC in.StylesheetUseCase,
	templateUC in.TemplateUseCase,
	bindingUC in.BindingUseCase,
	deliveryUC in.DeliveryUseCase,
) *Server {
	return &Server{
		contextUC:    contextUC,
		stylesheetUC: stylesheetUC,
		templateUC:   templateUC,
		bindingUC:    bindingUC,
		deliveryUC:   deliveryUC,
	}
}

// ============================================================================
// Contexts Handlers
// ============================================================================

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

// ============================================================================
// Stylesheets Handlers
// ============================================================================

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

// ============================================================================
// Templates Handlers
// ============================================================================

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

// ============================================================================
// Bindings Handlers
// ============================================================================

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

// ============================================================================
// Deliveries Handlers
// ============================================================================

func (s *Server) CreateDelivery(w http.ResponseWriter, r *http.Request) {
	var req gen.CreateDeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := s.deliveryUC.Dispatch(r.Context(), domain.IDType(req.TemplateId), domain.IDType(req.BindingId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(gen.DeliveryResponse{Status: "dispatched"})
}
