package psm

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

// const DEBUG_LEVEL = zerolog.TraceLevel

var logout = zerolog.ConsoleWriter{
	Out:        os.Stdout,
	TimeFormat: time.RFC3339,
}

func BuildLogger(level zerolog.Level) zerolog.Logger {
	return zerolog.New(logout).Level(level).
		With().Timestamp().Logger()
}

// Logger is a globally available logger instance.
var Logger = BuildLogger(zerolog.TraceLevel)
