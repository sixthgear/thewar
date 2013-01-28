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
