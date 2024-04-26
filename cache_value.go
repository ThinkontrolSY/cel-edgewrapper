package celedgewrapper

import (
	"reflect"
	"time"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

type CacheValue struct {
	Timestamp time.Time
	Var       interface{}
}

type Cache []CacheValue

func (c *Cache) Count(du time.Duration) int {
	count := 0
	now := time.Now()
	for _, v := range *c {
		if v.Timestamp.After(now.Add(-du)) {
			count++
		}
	}
	return count
}

func (c *Cache) Len() int {
	return len(*c)
}

func (c *Cache) Diff() float64 {
	if len(*c) < 2 {
		return 0
	}
	switch v := (*c)[len(*c)-1].Var.(type) {
	case int:
		return float64(v - (*c)[len(*c)-2].Var.(int))
	case int8:
		return float64(v - (*c)[len(*c)-2].Var.(int8))
	case int16:
		return float64(v - (*c)[len(*c)-2].Var.(int16))
	case int32:
		return float64(v - (*c)[len(*c)-2].Var.(int32))
	case int64:
		return float64(v - (*c)[len(*c)-2].Var.(int64))
	case uint:
		return float64(v - (*c)[len(*c)-2].Var.(uint))
	case uint8:
		return float64(v - (*c)[len(*c)-2].Var.(uint8))
	case uint16:
		return float64(v - (*c)[len(*c)-2].Var.(uint16))
	case uint32:
		return float64(v - (*c)[len(*c)-2].Var.(uint32))
	case uint64:
		return float64(v - (*c)[len(*c)-2].Var.(uint64))
	case float32:
		return float64(v - (*c)[len(*c)-2].Var.(float32))
	case float64:
		return v - (*c)[len(*c)-2].Var.(float64)
	case bool:
		if v && !(*c)[len(*c)-2].Var.(bool) {
			return 1
		}
		if !v && (*c)[len(*c)-2].Var.(bool) {
			return -1
		}
		return 0
	default:
		return 0
	}
}

func (c *Cache) Rising() bool {
	if len(*c) < 2 {
		return false
	}
	switch v := (*c)[len(*c)-1].Var.(type) {
	case int:
		return v > (*c)[len(*c)-2].Var.(int)
	case int8:
		return v > (*c)[len(*c)-2].Var.(int8)
	case int16:
		return v > (*c)[len(*c)-2].Var.(int16)
	case int32:
		return v > (*c)[len(*c)-2].Var.(int32)
	case int64:
		return v > (*c)[len(*c)-2].Var.(int64)
	case uint:
		return v > (*c)[len(*c)-2].Var.(uint)
	case uint8:
		return v > (*c)[len(*c)-2].Var.(uint8)
	case uint16:
		return v > (*c)[len(*c)-2].Var.(uint16)
	case uint32:
		return v > (*c)[len(*c)-2].Var.(uint32)
	case uint64:
		return v > (*c)[len(*c)-2].Var.(uint64)
	case float32:
		return v > (*c)[len(*c)-2].Var.(float32)
	case float64:
		return v > (*c)[len(*c)-2].Var.(float64)
	case bool:
		return v && !(*c)[len(*c)-2].Var.(bool)
	case string:
		return v > (*c)[len(*c)-2].Var.(string)
	default:
		return false
	}
}

func (c *Cache) Falling() bool {
	if len(*c) < 2 {
		return false
	}
	switch v := (*c)[len(*c)-1].Var.(type) {
	case int:
		return v < (*c)[len(*c)-2].Var.(int)
	case int8:
		return v < (*c)[len(*c)-2].Var.(int8)
	case int16:
		return v < (*c)[len(*c)-2].Var.(int16)
	case int32:
		return v < (*c)[len(*c)-2].Var.(int32)
	case int64:
		return v < (*c)[len(*c)-2].Var.(int64)
	case uint:
		return v < (*c)[len(*c)-2].Var.(uint)
	case uint8:
		return v < (*c)[len(*c)-2].Var.(uint8)
	case uint16:
		return v < (*c)[len(*c)-2].Var.(uint16)
	case uint32:
		return v < (*c)[len(*c)-2].Var.(uint32)
	case uint64:
		return v < (*c)[len(*c)-2].Var.(uint64)
	case float32:
		return v < (*c)[len(*c)-2].Var.(float32)
	case float64:
		return v < (*c)[len(*c)-2].Var.(float64)
	case bool:
		return !v && (*c)[len(*c)-2].Var.(bool)
	case string:
		return v < (*c)[len(*c)-2].Var.(string)
	default:
		return false
	}
}

// the CEL type to represent Test
var CacheType = types.NewTypeValue("CacheType", traits.ReceiverType)

func (t Cache) ConvertToNative(typeDesc reflect.Type) (interface{}, error) {
	panic("not required")
}

func (t Cache) ConvertToType(typeVal ref.Type) ref.Val {
	panic("not required")
}

func (c Cache) Equal(other ref.Val) ref.Val {
	o, ok := other.Value().(Cache)
	if ok {
		if len(o) == len(c) {
			return types.Bool(true)
		} else {
			return types.Bool(false)
		}
	} else {
		return types.ValOrErr(other, "%v is not of type Test", other)
	}
}

func (t Cache) Type() ref.Type {
	return CacheType
}

func (t Cache) Value() interface{} {
	return t
}

func (c Cache) Receive(function string, overload string, args []ref.Val) ref.Val {
	if function == "len" {
		return types.Int(c.Len())
	} else if function == "diff" {
		return types.Double(c.Diff())
	} else if function == "rising" {
		return types.Bool(c.Rising())
	} else if function == "falling" {
		return types.Bool(c.Falling())
	} else if function == "count" && len(args) == 1 {
		if du, dok := args[0].(types.Duration); dok {
			count := 0
			now := time.Now()
			for _, v := range c {
				if v.Timestamp.After(now.Add(-du.Duration)) {
					count++
				}
			}
			return types.Int(count)
		}
		return types.ValOrErr(CacheType, "count arg should be a duration")
	}
	return types.ValOrErr(CacheType, "no such function - %s", function)
}

func (t *Cache) HasTrait(trait int) bool {
	return trait == traits.ReceiverType
}

func (t *Cache) TypeName() string {
	return CacheType.TypeName()
}

type customTypeAdapter struct {
}

func (customTypeAdapter) NativeToValue(value interface{}) ref.Val {
	val, ok := value.(Cache)
	if ok {
		return val
	} else {
		//let the default adapter handle other cases
		return types.DefaultTypeAdapter.NativeToValue(value)
	}
}
