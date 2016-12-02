// Package config includes global configuration variables and defaults
// Config should not import any other subpackages to avoid circular imports
package config

import (
	"time"

	"github.com/Sirupsen/logrus"
)

// Defaults
var (
	// Number of concurrent workers
	Threads = 100

	// HTTP request timeout
	Timeout = 15 * time.Second

	// Location of script files
	ScriptDir = "./scripts"

	Verbosity logrus.Level = logrus.DebugLevel
)
