package main

import (
	"mymicro/micro/config"
	"mymicro/micro/registry/consul/api"
	"mymicro/micro/app"
	"mymicro/micro/registry"
	"mymicro/micro/registry/consul"
	"mymicro/pkg/log"
)

func NewApp(basename string) *app.App {
	cfg := config.New()
	appl := app.NewApp("my-service", basename, app.WithOptions(cfg), app.WithRunFunc(run(cfg)))
	return appl
}

func NewRegistrar(registry *options.RegistryOptions) registry.Registrar {
	c := api.DefaultConfig()
	c.Address = registry.Address
	c.Scheme = registry.Scheme
	cli, err := cli.NewClient(c)
	if err != nil {
		panic(err)
	}
	r := consul.New(cli, consul.WithHealthCheck(true))
	return r
}

func NewServiceApp(cfg *config.Config) (*gapp.App, error) {
	log.Init(cfg.Log)
	defer log.Flush()

	register := NewRegistrar(cfg.Registry)
	rpcServer, err := NewServiceHTTPServer(cfg)
	if err != nil {
		return nil, err
	}

	return gapp.New(
		gapp.WithName(cfg.Server.Name),
		gapp.WithRestServer(rpcServer),
		gapp.WithRegistrar(register),
	), nil
}

func run(cfg *config.Config) app.RunFunc {
	return func(baseName string) error {
		serviceApp, err := NewServiceApp(cfg)
		if err != nil {
			return err
		}

		if err := serviceApp.Run(); err != nil {
			log.Errorf("run service app error: %s", err)
			return err
		}
		return nil
	}
}

func main() {
	NewApp("myProject")
}
