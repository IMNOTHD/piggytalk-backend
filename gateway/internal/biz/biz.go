package biz

import (
	accountV1 "gateway/internal/biz/account/v1"
	eventV1 "gateway/internal/biz/event/v1"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(accountV1.NewAccountUsecase, eventV1.NewEventUsecase)
