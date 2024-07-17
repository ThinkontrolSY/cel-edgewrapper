package celedgewrapper

import (
	"log"
	"reflect"
	"sync"
	"time"

	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

type CacheValue struct {
	Timestamp time.Time
	Var       interface{}
}

// type Cache []CacheValue
type Cache struct {
	data   []CacheValue
	mutex  sync.RWMutex
	locked bool
}

func NewCache(v interface{}) *Cache {
	return &Cache{
		data: []CacheValue{
			{time.Now(), v},
		},
	}
}

func (c *Cache) Add(v interface{}, maxDur time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.locked {
		log.Println("Add operation is locked")
		return
	}

	now := time.Now()
	validTime := now.Add(-maxDur)
	newCache := make([]CacheValue, 0)
	for _, cv := range c.data {
		if cv.Timestamp.After(validTime) {
			newCache = append(newCache, cv)
		}
	}
	c.data = newCache

	if len(c.data) > 0 {
		lastIndex := len(c.data) - 1
		lastValue := c.data[lastIndex].Var

		if reflect.TypeOf(lastValue) == reflect.TypeOf(v) && reflect.DeepEqual(lastValue, v) {
			c.data[lastIndex] = CacheValue{
				Timestamp: now,
				Var:       v,
			}
			return
		}
	}
	c.data = append(c.data, CacheValue{
		Timestamp: now,
		Var:       v,
	})
}

// Clear 方法
func (c *Cache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make([]CacheValue, 0)
}

// Lock 和 Unlock 方法
func (c *Cache) Lock() {
	c.mutex.Lock()
	c.locked = true
	c.mutex.Unlock()
}

func (c *Cache) Unlock() {
	c.mutex.Lock()
	c.locked = false
	c.mutex.Unlock()
}

func (c *Cache) Len() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.data)
}

// Rising 方法
func (c *Cache) Rising() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if len(c.data) < 2 {
		return false
	}

	lastIndex := len(c.data) - 1
	value1 := c.data[lastIndex-1].Var
	value2 := c.data[lastIndex].Var

	if reflect.TypeOf(value1) != reflect.TypeOf(value2) {
		return false
	}

	switch v1 := value1.(type) {
	case string:
		v2 := value2.(string)
		return v1 < v2
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(value1).Int() < reflect.ValueOf(value2).Int()
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(value1).Uint() < reflect.ValueOf(value2).Uint()
	case float32, float64:
		return reflect.ValueOf(value1).Float() < reflect.ValueOf(value2).Float()
	case bool:
		return v1 && !value2.(bool)
	default:
		return false
	}
}

// Falling 方法
func (c *Cache) Falling() bool {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if len(c.data) < 2 {
		return false
	}

	lastIndex := len(c.data) - 1
	value1 := c.data[lastIndex-1].Var
	value2 := c.data[lastIndex].Var

	if reflect.TypeOf(value1) != reflect.TypeOf(value2) {
		return false
	}

	switch v1 := value1.(type) {
	case string:
		v2 := value2.(string)
		return v1 > v2
	case int, int8, int16, int32, int64:
		return reflect.ValueOf(value1).Int() > reflect.ValueOf(value2).Int()
	case uint, uint8, uint16, uint32, uint64:
		return reflect.ValueOf(value1).Uint() > reflect.ValueOf(value2).Uint()
	case float32, float64:
		return reflect.ValueOf(value1).Float() > reflect.ValueOf(value2).Float()
	case bool:
		return !v1 && value2.(bool)
	default:
		return false
	}
}

// the CEL type to represent Test
var CacheType = types.NewTypeValue("CacheType", traits.ReceiverType)

func (t *Cache) ConvertToNative(typeDesc reflect.Type) (interface{}, error) {
	panic("not required")
}

func (t *Cache) ConvertToType(typeVal ref.Type) ref.Val {
	panic("not required")
}

func (c *Cache) Equal(other ref.Val) ref.Val {
	o, ok := other.Value().(*Cache)
	if ok {
		if len(o.data) == len(c.data) {
			return types.Bool(reflect.DeepEqual(o.data, c.data))
		} else {
			return types.Bool(false)
		}
	} else {
		return types.ValOrErr(other, "%v is not of type Test", other)
	}
}

func (t *Cache) Type() ref.Type {
	return CacheType
}

func (t *Cache) Value() interface{} {
	return t
}

func (c *Cache) Receive(function string, overload string, args []ref.Val) ref.Val {
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
			for _, v := range c.data {
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
	val, ok := value.(*Cache)
	if ok {
		return val
	} else {
		//let the default adapter handle other cases
		return types.DefaultTypeAdapter.NativeToValue(value)
	}
}
