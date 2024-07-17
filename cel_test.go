package celedgewrapper

import (
	"testing"
	"time"

	"github.com/google/cel-go/checker/decls"
)

func TestCel(t *testing.T) {
	celRt, err := NewCelRuntime([]*CelRegVarible{
		{"name.tt", decls.String},
		{"group", decls.Double},
		{"age", decls.Uint},
		{"bytedata", decls.Bytes},
		{"b1", decls.Bool},
		{"b2", decls.Bool},
		{"b3", decls.Bool},
	})

	if err != nil {
		t.Fatal(err)
	}

	celRt.UpdateEnv([]*CelRegVarible{
		{"name.mm", decls.String},
	})

	fields := map[string]string{
		"test1":  "name.mm + 'test'",
		"test2":  `size(bytedata)`,
		"test3":  `size([b1, b2, b3].filter(i, i == true))`,
		"test4":  `bytedata.bit(1)`,
		"test5":  `bytedata.to_int()`,
		"test6":  `age + 1u`,
		"test7":  `group * 10.0`,
		"test8":  `age.cache.count(duration("60s"))`,
		"test9":  `age.cache.len()`,
		"test10": `age.cache.rising()`,
		"test11": `age.cache.falling()`,
	}

	bytedata := make([]byte, 2)
	// buf := new(bytes.Buffer)
	// err = binary.Write(buf, binary.BigEndian, int16(25689))
	// if err != nil {
	// 	t.Fatal("binary.Write failed:", err)
	// }
	// copy(bytedata, buf.Bytes())
	ageCache := NewCache(1)
	time.Sleep(1 * time.Second)
	ageCache.Add(3, 5*time.Second)
	time.Sleep(1 * time.Second)
	ageCache.Add(4, 5*time.Second)
	time.Sleep(1 * time.Second)
	ageCache.Add(6, 5*time.Second)
	time.Sleep(1 * time.Second)
	ageCache.Add(1, 5*time.Second)
	time.Sleep(1 * time.Second)
	ageCache.Add(1, 5*time.Second)
	time.Sleep(1 * time.Second)
	ageCache.Add(1, 5*time.Second)

	t.Log(ageCache.Len())

	time.Sleep(10 * time.Millisecond)
	ageCache.Add(2, 50*time.Second)
	t.Log(ageCache.Len())

	time.Sleep(10 * time.Millisecond)
	ageCache.Add(1, 50*time.Second)
	t.Log(ageCache.Len())

	time.Sleep(10 * time.Millisecond)
	ageCache.Add(3, 50*time.Second)
	t.Log(ageCache.Len())

	ageCache.Lock()

	time.Sleep(10 * time.Millisecond)
	ageCache.Add(1, 50*time.Second)
	t.Log(ageCache.Len())
	time.Sleep(10 * time.Millisecond)
	ageCache.Add(2, 50*time.Second)
	t.Log(ageCache.Len())
	time.Sleep(10 * time.Millisecond)
	ageCache.Add(3, 50*time.Second)
	t.Log(ageCache.Len())

	ageCache.Unlock()

	time.Sleep(10 * time.Millisecond)
	ageCache.Add(4, 50*time.Second)
	t.Log(ageCache.Len())
	time.Sleep(10 * time.Millisecond)
	ageCache.Add(4, 50*time.Second)
	t.Log(ageCache.Len())

	values := map[string]interface{}{
		"name.tt":   "s",
		"name.mm":   "test",
		"group":     2.0,
		"bytedata":  bytedata,
		"age":       uint(10),
		"b1":        true,
		"b2":        true,
		"b3":        false,
		"age.cache": ageCache,
	}

	for f, expr := range fields {
		tp, e := celRt.RegProgram(f, expr)
		if e != nil {
			t.Fatalf("key: %s, type: %s, expr: %s, error: %v", f, tp, expr, e)
		} else {
			t.Logf("key: %s, type: %s, expr: %s\n", f, tp, expr)
		}
	}

	t.Log("=======================================")

	for f, expr := range fields {
		v, e := celRt.Eval(f, values)
		if e != nil {
			t.Fatalf("key: %s, expr: %s, error: %v", f, expr, e)
		}
		t.Logf("key: %s, expr: %s, value: %v\n", f, expr, v)
	}

}
