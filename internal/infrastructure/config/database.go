package config

type IDatabaseConfig interface {
	DBHost() string
	DBPort() int
	DBUser() string
	DBPassword() string
	DBName() string
	DBSSLMode() string
}

type DatabaseConfig struct {
	DBHost_     string `yaml:"host"`
	DBPort_     int    `yaml:"port"`
	DBUser_     string `yaml:"user"`
	DBPassword_ string `yaml:"password"`
	DBName_     string `yaml:"dbname"`
	DBSSLMode_  string `yaml:"sslmode"`
}

func (d *DatabaseConfig) DBHost() string {
	return d.DBHost_
}

func (d *DatabaseConfig) DBPort() int {
	return d.DBPort_
}

func (d *DatabaseConfig) DBUser() string {
	return d.DBUser_
}

func (d *DatabaseConfig) DBPassword() string {
	return d.DBPassword_
}

func (d *DatabaseConfig) DBName() string {
	return d.DBName_
}

func (d *DatabaseConfig) DBSSLMode() string {
	return d.DBSSLMode_
}
