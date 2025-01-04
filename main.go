package main

func main() {
	logger := Logger{silent: false}
	logger.Info("This is an info message")
	logger.Warn("This is a warning message")
}
