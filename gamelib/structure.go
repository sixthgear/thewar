package gamelib

type StrStats struct {
}

type Struct struct {
	team       uint8  // who does this unit belong to?
	structType uint8  // what kind of unit is this?
	x, y       uint32 // current grid position	
	damage     uint8  // one of four damage levels
	facing     uint8  // one of six directions, this affects the special flank modifier
	statsBase  StrStats
	statsMod   StrStats
	orderQueue []Order
}
