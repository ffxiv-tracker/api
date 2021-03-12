//+build wireinject

package main

import (
	"github.com/google/wire"

	"ffxiv.anid.dev/internal/config"
	"ffxiv.anid.dev/internal/aws"
	"ffxiv.anid.dev/internal/manager"
	"ffxiv.anid.dev/internal/server"
)

func InitializeServer(c *config.Config) (*server.Server, error) {
	panic(wire.Build(
		server.NewServer,
		wire.FieldsOf(new(*config.Config), "TableName", "MasterPK"),
		wire.Struct(new(manager.TasksManager), "DynSvc", "Table", "MasterPK"),
		aws.NewSession,
		aws.NewDynamoClient,
	))
}
