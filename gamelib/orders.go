package gamelib

const (
	OR_DEFEND  = iota // (dir) face given direction, engage hostiles as noticed, default state
	OR_HOLD           // (dir) face given direction, do not engage
	OR_SUPRESS        // (target) attack the target hex or any other noticed hostiles
	OR_MOVE           // (target) move to the target hex quickly, do not engage
	OR_ATTACK         // (target, dir) move to the target hex methodically while facing dir, engage hostiles as noticed
)

type Order struct {
	Order uint8
	Path  []int
}
