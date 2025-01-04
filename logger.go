package main

import (
	"fmt"
)

const (
	reset  = "\033[0m"
	green  = "\033[32m" // Green for INFO
	yellow = "\033[33m" // Yellow for WARN
)

type Logger struct {
	silent bool
}

func (l *Logger) Info(v ...interface{}) {
	if !l.silent {
		fmt.Println(green+"INFO:"+reset, fmt.Sprint(v...))
	}
}

func (l *Logger) Warn(v ...interface{}) {
	if !l.silent {
		fmt.Println(yellow+"WARN:"+reset, fmt.Sprint(v...))
	}
}
