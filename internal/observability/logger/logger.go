package logger

import "log"

func Info(msg string, v ...interface{})  { log.Printf("INFO: "+msg, v...) }
func Error(msg string, v ...interface{}) { log.Printf("ERROR: "+msg, v...) }
