package biz

import (
	accountV1 "account/internal/biz/account/v1"
	relationV1 "account/internal/biz/relation/v1"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(accountV1.NewAccountUsecase, relationV1.NewFriendRelationUsecase)
