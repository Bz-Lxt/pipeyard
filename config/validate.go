package config

import "fmt"

func CheckDir(dir string) error {
	if dir == "" {
		return fmt.Errorf("empty dir")
	}
	if dir == "/" {
		return fmt.Errorf("refuse root dir")
	}
	return nil
}

func CheckAddr(addr string) error {
	if addr == "" {
		return fmt.Errorf("empty addr")
	}
	return nil
}

func (c Config) Validate() error {
	if err := CheckDir(c.Dir); err != nil {
		return err
	}
	if err := CheckAddr(c.Addr); err != nil {
		return err
	}
	return nil
}
