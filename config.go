package gosrv

import (
  "github.com/robfig/config"
)


type Config struct {
  *config.Config
  Env string
}


func NewConfig(env string) *Config {
  return &Config{&config.Config{}, env}
}


func ReadConfig(file, env string) (*Config, error) {
  cfg, err := config.ReadDefault(file)
  if err != nil { return nil, err }

  c := &Config{cfg, env}
  return c, nil
}


func (c Config) String(name string) (string, error) {
  return c.Config.String(c.Env, name)
}


func (c Config) Int(name string) (int, error) {
  return c.Config.Int(c.Env, name)
}


func (c Config) Bool(name string) (bool, error) {
  return c.Config.Bool(c.Env, name)
}
