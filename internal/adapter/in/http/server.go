package http

import (
	"github.com/ernstvorsteveld/go-email-template-sender/internal/adapter/in/http/gen"
	"github.com/ernstvorsteveld/go-email-template-sender/internal/application/port/in"
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
