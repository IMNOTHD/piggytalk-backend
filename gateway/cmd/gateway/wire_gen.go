// Code generated by Wire. DO NOT EDIT.

//go:generate go run github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"gateway/internal/biz/account/v1"
	v1_3 "gateway/internal/biz/event/v1"
	"gateway/internal/conf"
	"gateway/internal/data"
	"gateway/internal/server"
	v1_2 "gateway/internal/service/account/v1"
	v1_4 "gateway/internal/service/event/v1"
	v1_5 "gateway/internal/service/upload/v1"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
)

// Injectors from wire.go:

// initApp init kratos application.
func initApp(confServer *conf.Server, confData *conf.Data, logger log.Logger) (*kratos.App, func(), error) {
	accountUsecase := v1.NewAccountUsecase(logger)
	accountService := v1_2.NewAccountService(accountUsecase, logger)
	httpServer := server.NewHTTPServer(confServer, accountService, logger)
	dataData, cleanup, err := data.NewData(confData, logger)
	if err != nil {
		return nil, nil, err
	}
	eventRepo := data.NewEventRepo(dataData, logger)
	eventUsecase := v1_3.NewEventUsecase(eventRepo, logger)
	eventStreamService := v1_4.NewEventStreamService(eventUsecase, logger)
	uploadService := v1_5.NewUploadService(logger)
	grpcServer := server.NewGRPCServer(confServer, eventStreamService, uploadService, logger)
	app := newApp(logger, httpServer, grpcServer)
	return app, func() {
		cleanup()
	}, nil
}
