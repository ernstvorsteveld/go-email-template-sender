package http

import (
	"encoding/json"
	"net/http"

	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/in/http/gen"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/domain"
)

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
