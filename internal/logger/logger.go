package logger

import (
	"os"

	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func InitLog() {
	Log = zerolog.New(os.Stdout)
}
