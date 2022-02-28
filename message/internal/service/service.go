package service

import (
	service "message/internal/service/message/v1"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewGreeterService, service.NewMessageService)
