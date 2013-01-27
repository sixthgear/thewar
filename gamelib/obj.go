package gamelib

const (
	OBJ_INFANTRY = iota
	OBJ_VEHICLE
	OBJ_BOAT
	OBJ_AIRCRAFT
)

const (
	SUP_NORMAL = iota
	SUP_SUPRESSED
	SUP_PINNED
	SUP_ROUTED
)

const (
	DAM_HEALTHY = iota
	DAM_WOUNDED
	DAM_CRITICAL
	DAM_DESTROYED
)

const (
	MOR_NORMAL = iota
	MOR_HEROIC
	MOR_PANIC
	MOR_BERSERK
)

const (
	SPE_FIT = 1 << iota
	SPE_STUNNED
)

const IMPOSSIBLE = -9999

type ObjStats struct {
	MOV int // movement: hexes the unit may move
	RAN int // range: hexes away the unit can effectively target 
	VIS int // vision: hexes away the unit can see
	STH int // stealth: hexes the unit will negate from enemy vision on visibility check
	ATT int // attack: attack points the unit will contribute when attacking
	CQB int // cqb attack: attack points the unit contribute when attacking when distance == 1
	DEF int // defense: attack points will this unit negate from enemy in head-on attack
	FLK int // flank def: attack points will this unit negate from enemy while flanked
}

var suppressionMods = [...]ObjStats{
	SUP_NORMAL:    ObjStats{},
	SUP_SUPRESSED: ObjStats{ATT: -2},
	SUP_PINNED:    ObjStats{ATT: -2, MOV: IMPOSSIBLE},
	SUP_ROUTED:    ObjStats{MOV: IMPOSSIBLE, ATT: IMPOSSIBLE},
}

var damageMods = [...]ObjStats{
	DAM_HEALTHY:   ObjStats{},
	DAM_WOUNDED:   ObjStats{ATT: -2, DEF: -2, MOV: -2},
	DAM_CRITICAL:  ObjStats{ATT: -4, DEF: -4, MOV: -4},
	DAM_DESTROYED: ObjStats{MOV: IMPOSSIBLE, ATT: IMPOSSIBLE},
}

var moraleMods = [...]ObjStats{
	MOR_NORMAL:  ObjStats{},
	MOR_HEROIC:  ObjStats{ATT: +2},
	MOR_PANIC:   ObjStats{MOV: +1, ATT: IMPOSSIBLE},
	MOR_BERSERK: ObjStats{MOV: +1, ATT: +2, CQB: +4, DEF: -2},
}

var specialMods = [...]ObjStats{
	SPE_FIT:     ObjStats{MOV: +4},
	SPE_STUNNED: ObjStats{MOV: IMPOSSIBLE, ATT: IMPOSSIBLE},
}

type Obj struct {
	Team        int     // who does this unit belong to?
	Type        int     // what kind of unit is this?
	X, Y        int     // current grid position
	Fx, Fy, Fz  float32 // current world position
	AnimCounter int     // animation counter
	AnimTotal   int     // animation counter
	Damage      int     // one of four damage levels
	Supression  int     // one of four supression levels	
	Morale      int     // one of four morale levels	
	Facing      int     // one of six directions, this affects the special flank modifier
	StatsBase   ObjStats
	StatsMod    ObjStats
	OrderQueue  []Order
	Dest        *Hex
}

func (o *Obj) calcObjStats() {
	// base stats
	// + terrain modifiers
	// + damage modifiers
	// + supression modifiers
	// + morale modifiers
	// + special modifiers

	// visibility is determined by checking line of sight
	// then comparing this units stealth stat to the enemies vision stat
	// visible = distance_to_enemy <= en.vis - my.sth

	// if any calculated stat is <= 0, set it to IMPOSSIBLE (-9999)
	// this will allow any modifier to make the action impossible
	// and for the UI to reflect the impossibility of the action
	// suitable stats to make impossible are mv and at

}
