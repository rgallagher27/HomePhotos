package config

type Config struct {
	Host       string `envconfig:"host" default:"0.0.0.0"`
	Port       string `envconfig:"port" default:"8080"`
	DBPath     string `envconfig:"db_path" default:"./homephotos.db"`
	SourcePath string `envconfig:"source_path" default:"/source"`
	CachePath  string `envconfig:"cache_path" default:"/cache"`
	LogLevel   string `envconfig:"log_level" default:"info"`
	JWTSecret        string `envconfig:"jwt_secret"`
	RegistrationOpen bool   `envconfig:"registration_open" default:"true"`
}
