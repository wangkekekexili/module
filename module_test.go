package module_test

import (
	"testing"

	"github.com/wangkekekexili/module"
)

func TestModule(t *testing.T) {
	type config struct{}

	type logger struct {
		Config *config
	}

	type reporter struct {
		Config *config
		Logger *logger
	}

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
