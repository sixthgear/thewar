package gamelib

import (
// "encoding/json"
)

const (
	CMD_HELLO = iota
	CMD_MAP_REQ
	CMD_MAP_DATA
	CMD_ORD_REQ
	CMD_ORD_ACK
	CMD_MSG_ACK
	CMD_CHAT_REQ
	CMD_CHAT_ACK
	CMD_UNIT_ADD
	CMD_UNIT_REM
	CMD_UNIT_UPD
)

var COMM = [...]Command{
	CMD_HELLO:    Command{},
	CMD_MAP_REQ:  Command{},
	CMD_MAP_DATA: Command{},
	CMD_ORD_REQ:  Command{},
	CMD_ORD_ACK:  Command{},
	CMD_MSG_ACK:  Command{},
	CMD_CHAT_REQ: Command{},
	CMD_CHAT_ACK: Command{},
	CMD_UNIT_ADD: Command{},
	CMD_UNIT_REM: Command{},
	CMD_UNIT_UPD: Command{},
}

type Command struct {
	x int
}
