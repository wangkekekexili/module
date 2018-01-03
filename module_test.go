package module_test

import (
	"errors"
	"testing"

	"github.com/wangkekekexili/module"
)

type config struct {
	name string
}

func (c *config) Load() error {
	c.name = "config"
	return nil
}

type logger struct {
	Config *config

	name string
}

func (g *logger) Load() error {
	if g.Config == nil {
		return errors.New("config module must be loaded")
	}
	g.name = g.Config.name + "logger"
	return nil
}

type reporter struct {
	Config *config
	Logger *logger
}

func TestModule_initialize(t *testing.T) {
	app := &struct {
		Config   *config
		Logger   *logger
		Reporter *reporter
	}{}
	err := module.Load(app)
	if err != nil {
		t.Fatal(err)
	}

	// All modules should be initialized.
	if app.Config == nil || app.Logger == nil || app.Reporter == nil || app.Logger.Config == nil || app.Reporter.Config == nil ||
		app.Reporter.Logger == nil || app.Reporter.Logger.Config == nil {
		t.Fatal("not all modules are initialzied")
	}

	// Modules are singletons.
	if app.Config != app.Logger.Config || app.Config != app.Reporter.Config || app.Config != app.Reporter.Logger.Config {
		t.Fatal("modules are not initialzied as singletons")
	}
}

func TestModule_load(t *testing.T) {
	g := &logger{}
	err := module.Load(g)
	if err != nil {
		t.Fatal(err)
	}
	if g.name != "configlogger" {
		t.Error("logger is not loaded")
	}
	if g.Config.name != "config" {
		t.Error("config is not loaded")
	}
}
