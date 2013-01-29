package gamelib

import (
	"encoding/json"
	// "log"
)

const (
	OR_DEFEND  = iota // (dir) face given direction, engage hostiles as noticed, default state
	OR_HOLD           // (dir) face given direction, do not engage
	OR_SUPRESS        // (target) attack the target hex or any other noticed hostiles
	OR_MOVE           // (target) move to the target hex quickly, do not engage
	OR_ATTACK         // (target, dir) move to the target hex methodically while facing dir, engage hostiles as noticed
	OR_REQMAP
)

const (
	CMD_HELLO = iota
	CMD_MAP_REQ
	CMD_MAP_DATA
	CMD_ORD_REQ
	CMD_ORD_ACK
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
	CMD_CHAT_REQ: Command{},
	CMD_CHAT_ACK: Command{},
	CMD_UNIT_ADD: Command{},
	CMD_UNIT_REM: Command{},
	CMD_UNIT_UPD: Command{},
}

type Command struct {
	x int
}

type Order struct {
	Order  uint
	UnitId uint
	Path   []int
}

func (o *Order) Encode() []byte {
	output, _ := json.Marshal(o)
	return append(output, '\n')
}

func (o *Order) Decode(data []byte) (*Order, error) {
	err := json.Unmarshal(data, o)
	return o, err
}
