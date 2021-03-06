package gamelib

import "math/rand"

const (
	TURN_TICKS = 12
)

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
	Id          uint    // unique identifier
	Team        int     // who does this unit belong to?
	Type        int     // what kind of unit is this?
	AP          int     // remaining action points
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

func (o *Obj) NextDest(world *Map) {
	order := o.OrderQueue[0]
	o.Dest = world.Index(order.Path[len(order.Path)-1])
	cost := 1 - TMOD[o.Type][o.Dest.TerrainType].MOV
	o.AnimCounter = 0
	// if obj.Type == OBJ_AIRCRAFT {
	// 	x := float64(obj.Dest.Index%world.Width - obj.X)
	// 	y := float64(obj.Dest.Index/world.Width - obj.Y)
	// 	h := int(TURN_TICKS*math.Hypot(y, x)) / 2
	// 	obj.AnimTotal = h
	// } else {
	o.AnimTotal = TURN_TICKS * cost
	// }
}

func GenerateObjects(world *Map) {

	// generate random objects
	for i := 0; i < 40; i++ {

		o := new(Obj)
		o.Id = uint(i)
		o.Team = rand.Int() % 4
		o.Type = rand.Int() % 4
		o.Facing = rand.Int() % 6
		o.OrderQueue = make([]Order, 0)

		for {
			x, y := rand.Int()%(world.Width-8)+4, rand.Int()%(world.Width-8)+4
			hex := world.Lookup(x, y)
			t := uint32(0)
			switch o.Type {
			case OBJ_INFANTRY:
				t = T_FOREST
			case OBJ_VEHICLE:
				t = T_OUTDOOR
			case OBJ_BOAT:
				t = T_RIVER
			case OBJ_AIRCRAFT:
				t = hex.TerrainType
			}

			if hex.TerrainType == t && hex.Unit == nil {
				world.Objects = append(world.Objects, o)
				hex.Unit = o
				o.X, o.Y = x, y
				o.Fx, o.Fy, o.Fz = world.HexCenter(hex)
				if o.Type == OBJ_AIRCRAFT {
					o.Fy = 100
				}
				break
			}
		}
	}
}
