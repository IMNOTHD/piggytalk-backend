package service

import (
	accountV1 "account/internal/service/account/v1"
	relationV1 "account/internal/service/relation/v1"

	"github.com/google/wire"
)

// ProviderSet is service providers.
var ProviderSet = wire.NewSet(accountV1.NewAccountService, relationV1.NewFriendRelationService)
