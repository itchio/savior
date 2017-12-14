package savior

import (
	"log"
	"os"
)

var outputDebug = os.Getenv("SAVIOR_DEBUG") == "1"

func Debugf(format string, args ...interface{}) {
	log.Printf(format, args...)
}
