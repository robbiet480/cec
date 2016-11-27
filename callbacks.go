package cec

// #include <libcec/cecc.h>
import "C"

import (
	"time"
	"unsafe"
)

type LogicalAddress struct {
	LogicalAddress int
	Type           string
}

func NewLogicalAddress(address C.cec_logical_address) LogicalAddress {
	return LogicalAddress{LogicalAddress: int(address), Type: GetLogicalNameByAddress(int(address))}
}

type LogMessage struct {
	Message                     string
	Level                       string
	Direction                   string
	MillisecondsSinceConnection int64
	Timestamp                   time.Time
}

//export logMessageCallback
func logMessageCallback(c unsafe.Pointer, msg C.cec_log_message) C.uint8_t {
	var level string
	switch msg.level {
	case C.CEC_LOG_ERROR:
		level = "ERROR"
		break
	case C.CEC_LOG_WARNING:
		level = "WARNING"
		break
	case C.CEC_LOG_NOTICE:
		level = "NOTICE"
		break
	case C.CEC_LOG_TRAFFIC:
		level = "TRAFFIC"
		break
	case C.CEC_LOG_DEBUG:
		level = "DEBUG"
		break
	case C.CEC_LOG_ALL:
		level = "ALL"
		break
	default:
		break
	}
	stringMsg := C.GoString((&msg).message)
	direction := "N/A"
	if len(stringMsg) >= 2 {
		if stringMsg[0:2] == "<<" {
			direction = "Outbound"
		} else if stringMsg[0:2] == ">>" {
			direction = "Inbound"
		}
	}
	message := LogMessage{
		Message:                     stringMsg,
		Level:                       level,
		Direction:                   direction,
		MillisecondsSinceConnection: int64(msg.time),
		Timestamp:                   time.Now(),
	}
	CallbackEvents <- message

	return 1
}

type KeyPress struct {
	KeyCode     int
	KeyCodeName string
	Duration    int
	Timestamp   time.Time
}

//export keyPressCallback
func keyPressCallback(c unsafe.Pointer, keyPress C.cec_keypress) C.uint8_t {
	CallbackEvents <- KeyPress{
		KeyCode:     int(keyPress.keycode),
		KeyCodeName: GetUserControlKeyString(keyPress.keycode),
		Duration:    int(keyPress.duration),
		Timestamp:   time.Now(),
	}
	return 1
}

type DataPacket struct {
	Data interface{}
	Size int
}

type Command struct {
	Initiator       LogicalAddress
	Destination     LogicalAddress
	Acknowledged    bool
	EndOfMessage    bool
	Opcode          int
	OpcodeName      string
	Parameters      DataPacket
	OpcodeSet       bool
	TransmitTimeout int32
	Timestamp       time.Time
}

//export commandCallback
func commandCallback(c unsafe.Pointer, command C.cec_command) C.uint8_t {
	CallbackEvents <- Command{
		Initiator:       NewLogicalAddress(command.initiator),
		Destination:     NewLogicalAddress(command.destination),
		Acknowledged:    (int(command.ack) == 1),
		EndOfMessage:    (int(command.eom) == 1),
		Opcode:          int(command.opcode),
		OpcodeName:      GetOpcodeString(int(command.opcode)),
		Parameters:      DataPacket{Data: command.parameters.data, Size: int(command.parameters.size)},
		OpcodeSet:       (int(command.opcode_set) == 1),
		TransmitTimeout: int32(command.transmit_timeout),
		Timestamp:       time.Now(),
	}
	return 1
}

//export configurationChangedCallback
func configurationChangedCallback(c unsafe.Pointer, configuration C.libcec_configuration) C.uint8_t {
	return 1
}

type Parameter struct {
	Type string
	Data interface{}
}

type Alert struct {
	Type       string
	Parameters Parameter
	Timestamp  time.Time
}

//export alertCallback
func alertCallback(c unsafe.Pointer, alert C.libcec_alert, parameter C.libcec_parameter) C.uint8_t {
	var parameterType string
	switch parameter.paramType {
	case C.CEC_PARAMETER_TYPE_STRING:
		parameterType = "STRING"
	case C.CEC_PARAMETER_TYPE_UNKOWN:
		parameterType = "UNKOWN"
	}

	var alertType string
	switch alert {
	case C.CEC_ALERT_SERVICE_DEVICE:
		alertType = "SERVICE_DEVICE"
	case C.CEC_ALERT_CONNECTION_LOST:
		alertType = "CONNECTION_LOST"
	case C.CEC_ALERT_PERMISSION_ERROR:
		alertType = "PERMISSION_ERROR"
	case C.CEC_ALERT_PORT_BUSY:
		alertType = "PORT_BUSY"
	case C.CEC_ALERT_PHYSICAL_ADDRESS_ERROR:
		alertType = "PHYSICAL_ADDRESS_ERROR"
	case C.CEC_ALERT_TV_POLL_FAILED:
		alertType = "TV_POLL_FAILED"
	}

	CallbackEvents <- Alert{
		Type: alertType,
		Parameters: Parameter{
			Type: parameterType,
			Data: parameter.paramData,
		},
		Timestamp: time.Now(),
	}
	return 1
}

type MenuState struct {
	Activated bool
	Timestamp time.Time
}

// menuState is bool, 0 = activated, 1 = deactivated
//export menuStateChangedCallback
func menuStateChangedCallback(c unsafe.Pointer, state C.cec_menu_state) C.uint8_t {
	CallbackEvents <- MenuState{
		Activated: int(state) == 0,
		Timestamp: time.Now(),
	}
	return 1
}

type SourceActivated struct {
	Source    LogicalAddress
	Active    bool
	Timestamp time.Time
}

//export sourceActivatedCallback
func sourceActivatedCallback(c unsafe.Pointer, logicalAddress C.cec_logical_address, activated int) {
	CallbackEvents <- SourceActivated{
		Source:    NewLogicalAddress(logicalAddress),
		Active:    (activated == 1),
		Timestamp: time.Now(),
	}
}
