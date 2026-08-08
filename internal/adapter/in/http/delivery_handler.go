package http

import (
	"encoding/json"
	"net/http"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/in"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
	"github.com/google/uuid"
)

type DeliveryHandler struct {
	useCase in.DeliveryUseCase
}

func NewDeliveryHandler(uc in.DeliveryUseCase) *DeliveryHandler {
	return &DeliveryHandler{useCase: uc}
}

type CreateDeliveryRequest struct {
	TemplateID string `json:"template_id"`
	BindingID  string `json:"binding_id"`
}

func (h *DeliveryHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateDeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	templateUUID, err := uuid.Parse(req.TemplateID)
	if err != nil {
		http.Error(w, "invalid template_id", http.StatusBadRequest)
		return
	}

	bindingUUID, err := uuid.Parse(req.BindingID)
	if err != nil {
		http.Error(w, "invalid binding_id", http.StatusBadRequest)
		return
	}

	// Trigger the bulk orchestration
	err = h.useCase.Dispatch(r.Context(), domain.IDType(templateUUID), domain.IDType(bindingUUID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"status": "dispatched"})
}
