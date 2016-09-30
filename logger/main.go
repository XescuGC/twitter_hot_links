package logger

import (
	"log"
	"os"
)

type logFile struct {
	file *os.File
}

func NewLogFile(fileName string) *logFile {
	file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file", fileName, ":", err)
	}

	return &logFile{file}
}

func (l *logFile) WithNamespace(namespace string) *log.Logger {
	return log.New(l.file, "["+namespace+"] ", log.Ldate|log.Ltime|log.Lshortfile)
}
