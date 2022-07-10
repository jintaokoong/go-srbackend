package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
)

var lock = &sync.Mutex{}

var logger *log.Logger

func GetInstance() *log.Logger {
	if logger == nil {
		lock.Lock()
		defer lock.Unlock()
		if logger == nil {
			fmt.Println("Creating single instance now.")
			logger = log.New(os.Stdout,
				"[App] ",
				log.Ldate|log.Ltime|log.Lshortfile)
		}
	}

	return logger
}
