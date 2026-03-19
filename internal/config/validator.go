package config

import "errors"

// Validate ensures config is sane.
func (c *Config) Validate() error {
    if c.ServiceName == "" {
        return errors.New("service name required")
    }
    return nil
}
