package cec

/*
#cgo pkg-config: libcec
#include <stdio.h>
#include <errno.h>
#include <libcec/cecc.h>

ICECCallbacks g_callbacks;
// callbacks.go exports
void logMessageCallback(void *, const cec_log_message*);
void keyPressCallback(void *, const cec_keypress*);
void commandCallback(void *, const cec_command*);
void configurationChangedCallback(void *, const libcec_configuration*);
void alertCallback(void *, const libcec_alert, const libcec_parameter);
int menuStateChangedCallback(void *, const cec_menu_state);
void sourceActivatedCallback(void *, const cec_logical_address, uint8_t activated);

void setupCallbacks(libcec_configuration *conf)
{
	g_callbacks.logMessage = &logMessageCallback;
	g_callbacks.keyPress = &keyPressCallback;
	g_callbacks.commandReceived = &commandCallback;
	g_callbacks.configurationChanged = &configurationChangedCallback;
	g_callbacks.alert = &alertCallback;
	g_callbacks.menuStateChanged = &menuStateChangedCallback;
	g_callbacks.sourceActivated = &sourceActivatedCallback;
	(*conf).callbacks = &g_callbacks;
}

void setName(libcec_configuration *conf, char *name)
{
	snprintf((*conf).strDeviceName, 13, "%s", name);
}

static void clearLogicalAddresses(cec_logical_addresses* addresses)
{
	int i;

	addresses->primary = CECDEVICE_UNREGISTERED;
	for (i = 0; i < 16; i++)
		addresses->addresses[i] = 0;
}

void setLogicalAddress(cec_logical_addresses* addresses, cec_logical_address address)
{
	if (addresses->primary == CECDEVICE_UNREGISTERED)
		addresses->primary = address;

	addresses->addresses[(int) address] = 1;
}
*/
import "C"

import (
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"
)

// Connection class
type Connection struct {
	connection C.libcec_connection_t
}

type cecAdapter struct {
	Path string
	Comm string
}

var CallbackEvents chan interface{}

func cecInit(deviceName, deviceType string) (C.libcec_connection_t, error) {
	var connection C.libcec_connection_t
	var conf C.libcec_configuration

	conf.clientVersion = C.uint32_t(C.LIBCEC_VERSION_CURRENT)

	for i := 0; i < 5; i++ {
		conf.deviceTypes.types[i] = C.CEC_DEVICE_TYPE_RESERVED
	}
	if deviceType == "tv" {
		conf.deviceTypes.types[0] = C.CEC_DEVICE_TYPE_TV
	} else if deviceType == "recording" {
		conf.deviceTypes.types[0] = C.CEC_DEVICE_TYPE_RECORDING_DEVICE
	} else if deviceType == "reserved" {
		conf.deviceTypes.types[0] = C.CEC_DEVICE_TYPE_RESERVED
	} else if deviceType == "tuner" {
		conf.deviceTypes.types[0] = C.CEC_DEVICE_TYPE_TUNER
	} else if deviceType == "playback" {
		conf.deviceTypes.types[0] = C.CEC_DEVICE_TYPE_PLAYBACK_DEVICE
	} else if deviceType == "audio" {
		conf.deviceTypes.types[0] = C.CEC_DEVICE_TYPE_AUDIO_SYSTEM
	} else {
		conf.deviceTypes.types[0] = C.CEC_DEVICE_TYPE_RECORDING_DEVICE
	}

	C.setName(&conf, C.CString(deviceName))

	CallbackEvents = make(chan interface{})
	C.setupCallbacks(&conf)

	connection = C.libcec_initialise(&conf)
	if connection == C.libcec_connection_t(nil) {
		return connection, errors.New("Failed to init CEC")
	}
	return connection, nil
}

func getAdapter(connection C.libcec_connection_t, name string) (cecAdapter, error) {
	var adapter cecAdapter

	var deviceList [10]C.cec_adapter
	devicesFound := int(C.libcec_find_adapters(connection, &deviceList[0], 10, nil))

	for i := 0; i < devicesFound; i++ {
		device := deviceList[i]
		adapter.Path = C.GoStringN(&device.path[0], 1024)
		adapter.Comm = C.GoStringN(&device.comm[0], 1024)

		if strings.Contains(adapter.Path, name) || strings.Contains(adapter.Comm, name) {
			return adapter, nil
		}
	}

	return adapter, errors.New("No Device Found")
}

func openAdapter(connection C.libcec_connection_t, adapter cecAdapter) error {
	C.libcec_init_video_standalone(connection)

	result := C.libcec_open(connection, C.CString(adapter.Comm), C.CEC_DEFAULT_CONNECT_TIMEOUT)
	if result < 1 {
		return errors.New("Failed to open adapter")
	}

	return nil
}

// Transmit CEC command - command is encoded as a hex string with
// colons (e.g. "40:04")
func (c *Connection) Transmit(command string) error {
	var cecCommand C.cec_command

	cmd, err := hex.DecodeString(removeSeparators(command))
	if err != nil {
		log.Fatal(err)
	}
	cmdLen := len(cmd)

	if cmdLen > 0 {
		cecCommand.initiator = C.cec_logical_address((cmd[0] >> 4) & 0xF)
		cecCommand.destination = C.cec_logical_address(cmd[0] & 0xF)
		if cmdLen > 1 {
			cecCommand.opcode_set = 1
			cecCommand.opcode = C.cec_opcode(cmd[1])
		} else {
			cecCommand.opcode_set = 0
		}
		if cmdLen > 2 {
			cecCommand.parameters.size = C.uint8_t(cmdLen - 2)
			for i := 0; i < cmdLen-2; i++ {
				cecCommand.parameters.data[i] = C.uint8_t(cmd[i+2])
			}
		} else {
			cecCommand.parameters.size = 0
		}
	}

	result := C.libcec_transmit(c.connection, (*C.cec_command)(&cecCommand))
	if result < 1 {
		return errors.New("Failed to transmit!")
	}
	return nil
}

// Destroy - destroy the cec connection
func (c *Connection) Destroy() {
	C.libcec_destroy(c.connection)
}

// PowerOn - power on the device with the given logical address
func (c *Connection) PowerOn(address int) error {
	if C.libcec_power_on_devices(c.connection, C.cec_logical_address(address)) != 1 {
		return errors.New("Error in cec_power_on_devices")
	}
	return nil
}

// Standby - put the device with the given address in standby mode
func (c *Connection) Standby(address int) error {
	if C.libcec_standby_devices(c.connection, C.cec_logical_address(address)) != 1 {
		return errors.New("Error in cec_standby_devices")
	}
	return nil
}

// VolumeUp - send a volume up command to the amp if present
func (c *Connection) VolumeUp() error {
	if C.libcec_volume_up(c.connection, 1) != 0 {
		return errors.New("Error in cec_volume_up")
	}
	return nil
}

// VolumeDown - send a volume down command to the amp if present
func (c *Connection) VolumeDown() error {
	if C.libcec_volume_down(c.connection, 1) != 0 {
		return errors.New("Error in cec_volume_down")
	}
	return nil
}

// Mute - send a mute/unmute command to the amp if present
func (c *Connection) Mute() error {
	if C.libcec_mute_audio(c.connection, 1) != 0 {
		return errors.New("Error in cec_mute_audio")
	}
	return nil
}

// KeyPress - send a key press (down) command code to the given address
func (c *Connection) KeyPress(address int, key int) error {
	if C.libcec_send_keypress(c.connection, C.cec_logical_address(address), C.cec_user_control_code(key), 1) != 1 {
		return errors.New("Error in cec_send_keypress")
	}
	return nil
}

// KeyRelease - send a key releas command to the given address
func (c *Connection) KeyRelease(address int) error {
	if C.libcec_send_key_release(c.connection, C.cec_logical_address(address), 1) != 1 {
		return errors.New("Error in cec_send_key_release")
	}
	return nil
}

// GetActiveDevices - returns an array of active devices
func (c *Connection) GetActiveDevices() [16]bool {
	var devices [16]bool
	result := C.libcec_get_active_devices(c.connection)

	for i := 0; i < 16; i++ {
		if int(result.addresses[i]) > 0 {
			devices[i] = true
		}
	}

	return devices
}

// GetActiveSource - returns the logical address of the currently active source
func (c *Connection) GetActiveSource() int {
	return int(C.libcec_get_active_source(c.connection))
}

// GetDeviceOSDName - get the OSD name of the specified device
func (c *Connection) GetDeviceOSDName(address int) string {
        var name *C.char = C.CString("")
 	C.libcec_get_device_osd_name(c.connection, C.cec_logical_address(address), name)
	return C.GoString(name)
}

// IsActiveSource - check if the device at the given address is the active source
func (c *Connection) IsActiveSource(address int) bool {
	result := C.libcec_is_active_source(c.connection, C.cec_logical_address(address))

	if int(result) != 0 {
		return true
	}

	return false
}

// GetDeviceVendorID - Get the Vendor-ID of the device at the given address
func (c *Connection) GetDeviceVendorID(address int) uint64 {
	result := C.libcec_get_device_vendor_id(c.connection, C.cec_logical_address(address))

	return uint64(result)
}

// GetDevicePhysicalAddress - Get the physical address of the device at
// the given logical address
func (c *Connection) GetDevicePhysicalAddress(address int) string {
	result := C.libcec_get_device_physical_address(c.connection, C.cec_logical_address(address))

	return fmt.Sprintf("%x.%x.%x.%x", (uint(result)>>12)&0xf, (uint(result)>>8)&0xf, (uint(result)>>4)&0xf, uint(result)&0xf)
}

// GetDevicePowerStatus - Get the power status of the device at the
// given address
func (c *Connection) GetDevicePowerStatus(address int) string {
	result := C.libcec_get_device_power_status(c.connection, C.cec_logical_address(address))

	// C.CEC_POWER_STATUS_UNKNOWN == error

	if int(result) == C.CEC_POWER_STATUS_ON {
		return "on"
	} else if int(result) == C.CEC_POWER_STATUS_STANDBY {
		return "standby"
	} else if int(result) == C.CEC_POWER_STATUS_IN_TRANSITION_STANDBY_TO_ON {
		return "starting"
	} else if int(result) == C.CEC_POWER_STATUS_IN_TRANSITION_ON_TO_STANDBY {
		return "shutting down"
	} else {
		return ""
	}
}

func (c *Connection) GetAudioStatus() string {
	result := C.libcec_audio_get_status(c.connection)

	if int(result) == C.CEC_AUDIO_MUTE_STATUS_MASK {
		return "MUTE"
	} else if int(result) == C.CEC_AUDIO_VOLUME_STATUS_MASK {
		return "MASK"
	} else if int(result) == C.CEC_AUDIO_VOLUME_MIN {
		return "0"
	} else if int(result) == C.CEC_AUDIO_VOLUME_MAX {
		return "100"
	} else if int(result) == C.CEC_AUDIO_VOLUME_STATUS_UNKNOWN {
		return "Unknown"
	} else {
		return "OTHER"
	}

}

func (c *Connection) PollDevice(address int) bool {
	result := C.libcec_poll_device(c.connection, C.cec_logical_address(address))

	return (result != 0)
}

// GetVendorString - Get vendor string by ID
func GetVendorString(id uint64) string {
	var vendorName *C.char = C.CString("")
	_, err := C.libcec_vendor_id_to_string(C.cec_vendor_id(id), vendorName, 10)
	if err != nil {
		return "Unknown"
	}
	return C.GoString(vendorName)
}

// GetOpcodeString - Get opcode string by hex
func GetOpcodeString(opcode int) string {
	var opcodeString *C.char = C.CString("")
	_, err := C.libcec_opcode_to_string(C.cec_opcode(opcode), opcodeString, 10)
	if err != nil {
		return "Unknown"
	}
	return C.GoString(opcodeString)
}

// GetUserControlKeyString - Get user control key string by int
func GetUserControlKeyString(key C.cec_user_control_code) string {
	var keyString *C.char = C.CString("")
	_, err := C.libcec_user_control_key_to_string(key, keyString, 10)
	if err != nil {
		return "Unknown"
	}
	return C.GoString(keyString)
}

// GetLogicalNameByAddress - get logical name by address
func GetLogicalNameByAddress(addr int) string {
	var logicalName *C.char = C.CString("")
	 _, err := C.libcec_logical_address_to_string(C.cec_logical_address(addr), logicalName, 10)
	if err != nil {
		return "Unknown"
	}
	return C.GoString(logicalName)
}
