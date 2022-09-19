package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/cel-go/checker/decls"
)

func TestCel(t *testing.T) {
	celRt, err := NewCelRuntime([]*CelRegVarible{
		{"name.tt", decls.String},
		{"group", decls.Int},
		{"age", decls.Int},
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
		"test1": "name.mm == 'test'",
		"test2": `size(bytedata)`,
		"test3": `size([b1, b2, b3].filter(i, i == true))`,
		"test4": `bytedata.bit(1)`,
		"test5": `bytedata.to_int()`,
		"test6": `age.cache.count(duration("4s"))`,
		"test7": `group + age`,
		"test8": `group * age`,
	}

	bytedata := make([]byte, 2)
	bytedata[0] = 23
	values := map[string]interface{}{
		"name.tt":  "s",
		"name.mm":  "test",
		"group":    2,
		"bytedata": bytedata,
		"age":      10,
		"b1":       true,
		"b2":       true,
		"b3":       false,
		"age.cache": Cache([]CacheValue{
			{time.Now().Add(-2 * time.Second), 1},
			{time.Now().Add(-3 * time.Second), 3},
			{time.Now().Add(-4 * time.Second), 4},
			{time.Now().Add(-time.Second), 6},
			{time.Now(), 12},
		}),
	}

	for f, expr := range fields {
		e := celRt.RegProgram(f, expr)
		if e != nil {
			t.Fatalf("key: %s, expr: %s, error: %v", f, expr, e)
		}
	}

	for f, expr := range fields {
		v, e := celRt.Eval(f, values)
		if e != nil {
			t.Fatalf("key: %s, expr: %s, error: %v", f, expr, e)
		}
		fmt.Printf("key: %s, expr: %s, value: %v\n", f, expr, v)
	}

}
