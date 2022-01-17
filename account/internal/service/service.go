package service

import (
	service "account/internal/service/account/v1"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(service.NewAccountService)
