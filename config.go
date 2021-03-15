package gormjqdt

import "gorm.io/gorm"

// Config defines the config for middleware.
type Config struct {
	// CaseSensitiveFilter config to determine current gormdtt using case sensitive filter.
	//
	// Optional. Default: false
	CaseSensitiveFilter bool

	// Model defines a model that using for retrive pagination resource.
	//
	// Required. Default: nil
	Model interface{}

	// Engine defines a DB engine (in this case, we using gorm) that using for build the query.
	//
	// Required. Default: nil
	Engine *gorm.DB

	// Dialect defines DB dialect that using in your app.
	//
	// Optional. Default: get dialect from connection
	Dialect string
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	CaseSensitiveFilter: false,
	Model:               nil,
	Engine:              nil,
	Dialect:             "",
}

// Helper function to set default values
func configDefault(config ...Config) Config {
	// Return default config if nothing provided
	if len(config) < 1 {
		if ConfigDefault.Engine != nil {
			ConfigDefault.Dialect = ConfigDefault.Engine.Name()
		}

		return ConfigDefault
	}

	// Override default config
	cfg := config[0]

	// Set default values
	if cfg.Model == nil {
		cfg.Model = ConfigDefault.Model
	}
	if cfg.Engine == nil {
		cfg.Engine = ConfigDefault.Engine
	}
	if cfg.Dialect == "" {
		cfg.Dialect = cfg.Engine.Name()
	}

	return cfg
}
