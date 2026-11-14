package config

type IServiceConfig interface {
	Port() int
	AllowedOrigins() []string
	InstanceURL() string
}

type ServiceConfig struct {
	Port_           int      `yaml:"port"`
	AllowedOrigins_ []string `yaml:"allowed_origins"`
	InstanceURL_    string   `yaml:"instance_url"`
}

func (c *ServiceConfig) Port() int {
	return c.Port_
}

func (c *ServiceConfig) AllowedOrigins() []string {
	return c.AllowedOrigins_
}

func (c *ServiceConfig) InstanceURL() string {
	return c.InstanceURL_
}
