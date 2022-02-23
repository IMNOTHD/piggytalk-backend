package service

import (
	accountV1 "gateway/internal/service/account/v1"
	eventV1 "gateway/internal/service/event/v1"
	uploadV1 "gateway/internal/service/upload/v1"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(NewTestService, accountV1.NewAccountService, eventV1.NewEventStreamService, uploadV1.NewUploadService)
