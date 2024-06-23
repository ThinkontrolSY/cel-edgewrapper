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

func (c *Cache) Rising() bool {
	if len(*c) < 2 {
		return false
	}
	for i := 0; i < len(*c)-1; i++ {
		currentVar := (*c)[i].Var
		nextVar := (*c)[i+1].Var
		switch v := currentVar.(type) {
		case int:
			if v < nextVar.(int) {
				return true
			}
		case int8:
			if v < nextVar.(int8) {
				return true
			}
		case int16:
			if v < nextVar.(int16) {
				return true
			}
		case int32:
			if v < nextVar.(int32) {
				return true
			}
		case int64:
			if v < nextVar.(int64) {
				return true
			}
		case uint:
			if v < nextVar.(uint) {
				return true
			}
		case uint8:
			if v < nextVar.(uint8) {
				return true
			}
		case uint16:
			if v < nextVar.(uint16) {
				return true
			}
		case uint32:
			if v < nextVar.(uint32) {
				return true
			}
		case uint64:
			if v < nextVar.(uint64) {
				return true
			}
		case float32:
			if v < nextVar.(float32) {
				return true
			}
		case float64:
			if v < nextVar.(float64) {
				return true
			}
		case bool:
			return !v && nextVar.(bool)
		case string:
			if v < nextVar.(string) {
				return true
			}
		default:
			continue
		}
	}
	return false
}

func (c *Cache) Falling() bool {
	if len(*c) < 2 {
		return false
	}
	for i := 0; i < len(*c)-1; i++ {
		currentVar := (*c)[i].Var
		nextVar := (*c)[i+1].Var
		switch v := currentVar.(type) {
		case int:
			if v > nextVar.(int) {
				return true
			}
		case int8:
			if v > nextVar.(int8) {
				return true
			}
		case int16:
			if v > nextVar.(int16) {
				return true
			}
		case int32:
			if v > nextVar.(int32) {
				return true
			}
		case int64:
			if v > nextVar.(int64) {
				return true
			}
		case uint:
			if v > nextVar.(uint) {
				return true
			}
		case uint8:
			if v > nextVar.(uint8) {
				return true
			}
		case uint16:
			if v > nextVar.(uint16) {
				return true
			}
		case uint32:
			if v > nextVar.(uint32) {
				return true
			}
		case uint64:
			if v > nextVar.(uint64) {
				return true
			}
		case float32:
			if v > nextVar.(float32) {
				return true
			}
		case float64:
			if v > nextVar.(float64) {
				return true
			}
		case bool:
			return v && !nextVar.(bool)
		case string:
			if v > nextVar.(string) {
				return true
			}
		default:
			continue
		}
	}
	return false
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
