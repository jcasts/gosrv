package gosrv

import (
  "github.com/robfig/config"
)


// Environment-specific config for Server configuration and based on the
// excellent config lib by robfig (github.com/robfig/config).
type Config struct {
  *config.Config
  Env string
}


// Create a new config for a given environment.
func NewConfig(env string) *Config {
  return &Config{&config.Config{}, env}
}


// Create a new config by reading from a config file, for a given environment.
func ReadConfig(file, env string) (*Config, error) {
  cfg, err := config.ReadDefault(file)
  if err != nil { return nil, err }

  c := &Config{cfg, env}
  return c, nil
}


// Get config value as String.
func (c Config) String(name string) (string, error) {
  return c.Config.String(c.Env, name)
}


// Get config value as Int.
func (c Config) Int(name string) (int, error) {
  return c.Config.Int(c.Env, name)
}


// Get config value as Bool.
func (c Config) Bool(name string) (bool, error) {
  return c.Config.Bool(c.Env, name)
}
