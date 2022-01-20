package biz

import (
	"account/internal/biz/account/v1"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(v1.NewAccountUsecase)
