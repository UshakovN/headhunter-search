package redis

type Config struct {
	Addr     string `yaml:"addr" required:"true"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}
