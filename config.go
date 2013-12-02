package srv

import (
  "github.com/robfig/config"
)


type Config struct {
  *config.Config
  Env string
}


func ReadConfig(file, env string) (*Config, error) {
  cfg, err := config.ReadDefault(config_file)
  if err != nil { return nil, err }

  c := &Config{cfg, env}
  return c, nil
}


func (c Config) String(name string) string {
  return c.Config.String(c.Env, name)
}


func (c Config) Int(name string) int {
  return c.Config.Int(c.Env, name)
}


func (c Config) Bool(name string) bool {
  return c.Config.Bool(c.Env, name)
}
