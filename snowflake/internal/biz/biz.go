package biz

import (
	v1 "snowflake/internal/biz/snowflake/v1"

	"github.com/google/wire"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(NewGreeterUsecase, v1.NewSnowflakeUsecase)
