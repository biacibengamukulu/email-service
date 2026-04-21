package logger

import (
	"log"
	"os"
)

type Logger struct {
	*log.Logger
}

func New(service string) *Logger {
	return &Logger{Logger: log.New(os.Stdout, "["+service+"] ", log.LstdFlags|log.Lmicroseconds)}
}
