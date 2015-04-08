package cec

// #include <libcec/cecc.h>
import "C"

import (
	"log"
	"unsafe"
)

//export logMessageCallback
func logMessageCallback(c unsafe.Pointer, msg C.cec_log_message) C.int {
	var logPrefix string
	switch msg.level {
	case C.CEC_LOG_ERROR:
		logPrefix = "ERROR:   "
		break
	case C.CEC_LOG_WARNING:
		logPrefix = "WARNING: "
		break
	case C.CEC_LOG_NOTICE:
		logPrefix = "NOTICE:  "
		break
	case C.CEC_LOG_TRAFFIC:
		logPrefix = "TRAFFIC: "
		break
	case C.CEC_LOG_DEBUG:
		logPrefix = "DEBUG:   "
		break
	default:
		break
	}
	log.Println(logPrefix + C.GoString(&msg.message[0]))

	return 0
}
