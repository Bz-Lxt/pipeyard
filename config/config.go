// Package config 描述作业场打开参数。零值可用。
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/Bz-Lxt/pipeyard/clock"
)

const (
	DefaultAddr     = ":8080"
	DefaultWorkers  = 4
	DefaultLeaseSec = 30
	DefaultMaxJobs  = 256
	SQLiteName      = "yard.sqlite"
	JournalName     = "journal.wal"
	WebDirName      = "web"
)

// Config 打开作业场与 HTTP 服务所需的全部旋钮。
type Config struct {
	Dir      string
	Addr     string
	Workers  int
	LeaseSec int
	MaxJobs  int
	Clock    clock.Clock
	ReadOnly bool
}

func (c Config) withDefaults() Config {
	if c.Dir == "" {
		c.Dir = "data"
	}
	if c.Addr == "" {
		c.Addr = DefaultAddr
	}
	if c.Workers <= 0 {
		c.Workers = DefaultWorkers
	}
	if c.LeaseSec <= 0 {
		c.LeaseSec = DefaultLeaseSec
	}
	if c.MaxJobs <= 0 {
		c.MaxJobs = DefaultMaxJobs
	}
	if c.Clock == nil {
		c.Clock = clock.Beijing{}
	}
	return c
}

// Normalize 校验并填默认值。
func Normalize(c Config) (Config, error) {
	c = c.withDefaults()
	if c.Workers > 64 {
		return c, fmt.Errorf("workers %d too large", c.Workers)
	}
	if c.LeaseSec > 3600 {
		return c, fmt.Errorf("lease %d too large", c.LeaseSec)
	}
	abs, err := filepath.Abs(c.Dir)
	if err != nil {
		return c, err
	}
	c.Dir = abs
	return c, nil
}

func (c Config) SQLitePath() string  { return filepath.Join(c.Dir, SQLiteName) }
func (c Config) JournalPath() string { return filepath.Join(c.Dir, JournalName) }

func (c Config) WebDir() string {
	if _, err := os.Stat(filepath.Join(c.Dir, WebDirName)); err == nil {
		return filepath.Join(c.Dir, WebDirName)
	}
	return WebDirName
}

// FromEnv 读 PIPEYARD_* 环境变量覆盖字段。
func FromEnv(c Config) Config {
	if v := os.Getenv("PIPEYARD_DIR"); v != "" {
		c.Dir = v
	}
	if v := os.Getenv("PIPEYARD_ADDR"); v != "" {
		c.Addr = v
	}
	if v := os.Getenv("PIPEYARD_WORKERS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.Workers = n
		}
	}
	if v := os.Getenv("PIPEYARD_LEASE"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.LeaseSec = n
		}
	}
	if v := os.Getenv("PIPEYARD_MAX_JOBS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			c.MaxJobs = n
		}
	}
	return c
}
