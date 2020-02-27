package minivmm

import (
	"github.com/caarlos0/env"
)

// Config is the minivmm configuration structure.
type Config struct {
	Dir               string   `env:"VMM_DIR" envDefault:"/opt/minivmm"`
	Port              int      `env:"VMM_LISTEN_PORT" envDefault:"14151"`
	Origin            string   `env:"VMM_ORIGIN,required"`
	OIDC              string   `env:"VMM_OIDC_URL"`
	Agents            []string `env:"VMM_AGENTS" envSeparator:","`
	CorsOrigins       []string `env:"VMM_CORS_ALLOWED_ORIGINS" envSeparator:","`
	SubnetCIDR        string   `env:"VMM_SUBNET_CIDR"`
    NameServers       []string `env:"VMM_NAME_SERVERS" envDefault:"1.1.1.1,1.0.0.1" envSeparator:","`
	ServerCert        string   `env:"VMM_SERVER_CERT"`
	ServerKey         string   `env:"VMM_SERVER_KEY"`
	NoTLS             bool     `env:"VMM_NO_TLS" envDefault:"false"`
	NoAuth            bool     `env:"VMM_NO_AUTH" envDefault:"false"`
	NoKvm             bool     `env:"VMM_NO_KVM" envDefault:"false"`
	VNCKeyboardLayout string   `env:"VMM_VNC_KEYBOARD_LAYOUT" envDefault:"en-us"`
}

// C is a global configuration object.
var C *Config = nil

// ParseConfig loads configurations from environment variables and sets to the global configuration
func ParseConfig() error {
	c := Config{}
	if err := env.Parse(&c); err != nil {
		return err
	}
	C = &c
	return nil
}

// SetConfig sets the global configuration with the given one. This function is for testing.
func SetConfig(c *Config) {
	C = c
}
