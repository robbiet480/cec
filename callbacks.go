package cec

// #include <libcec/cecc.h>
import "C"

import (
	"fmt"
	"log"
	"unsafe"
)

//export logMessageCallback
func logMessageCallback(c unsafe.Pointer, msg C.cec_log_message) C.uint8_t {
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

//export logSourceChangeCallback
func logSourceChangeCallback(c unsafe.Pointer, logicalAddress C.cec_logical_address, activated int) {
	result := C.cec_get_device_physical_address(C.cec_logical_address(logicalAddress))

	log.Println(fmt.Sprintf("Input changed to %x.%x.%x.%x", (uint(result)>>12)&0xf, (uint(result)>>8)&0xf, (uint(result)>>4)&0xf, uint(result)&0xf))
}
