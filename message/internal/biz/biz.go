package biz

import (
	v1 "message/internal/biz/message/v1"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewGreeterUsecase, v1.NewMessageUsecase)
