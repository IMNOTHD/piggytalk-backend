package service

import (
	"gateway/internal/service/account/v1"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewTestService, v1.NewAccountService)
