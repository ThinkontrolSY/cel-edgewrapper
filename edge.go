package celedgewrapper

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"regexp"
	"sync"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/checker/decls"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter/functions"
	exprpb "google.golang.org/genproto/googleapis/api/expr/v1alpha1"
)

const (
	STRING_DT_REG = `^(W)?String\[(\d+)\]$`
)

type CelRegVarible struct {
	VarName string
	VarType *exprpb.Type
}
type customLib struct {
	envOptions []cel.EnvOption
}

func (c *customLib) CompileOptions() []cel.EnvOption {
	return c.envOptions
}

func (customLib) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{}
}

type CelRuntime struct {
	mu          sync.Mutex
	celEnv      *cel.Env
	celProgOpts cel.ProgramOption
	declsVars   []*exprpb.Decl
	prgs        map[string]cel.Program
}

func NewCelRuntime(regVariables []*CelRegVarible) (*CelRuntime, error) {
	declsVars := []*exprpb.Decl{
		decls.NewConst("Pi", decls.Double, &exprpb.Constant{ConstantKind: &exprpb.Constant_DoubleValue{
			DoubleValue: 3.14159265358979323846264338327950288419716939937510582097494459}}),
		decls.NewFunction("bit", decls.NewInstanceOverload("bit_in_bytes_bytes_int",
			[]*exprpb.Type{decls.Bytes, decls.Int},
			decls.Bool)),
		decls.NewFunction("len", decls.NewInstanceOverload("len_cache",
			[]*exprpb.Type{decls.NewObjectType("CacheType")},
			decls.Int)),
		decls.NewFunction("count", decls.NewInstanceOverload("count_cache",
			[]*exprpb.Type{decls.NewObjectType("CacheType"), decls.Duration},
			decls.Int)),
		decls.NewFunction("to_int", decls.NewInstanceOverload("bytes_to_int_int",
			[]*exprpb.Type{decls.Bytes},
			decls.Int)),
		decls.NewFunction("to_uint", decls.NewInstanceOverload("bytes_to_uint_uint",
			[]*exprpb.Type{decls.Bytes},
			decls.Int)),
	}
	for _, v := range regVariables {
		declsVars = append(declsVars, decls.NewVar(v.VarName, v.VarType), decls.NewVar(v.VarName+".cache", decls.NewObjectType("CacheType")))
	}
	celProgOpts := cel.Functions(
		&functions.Overload{
			Operator: "bit_in_bytes_bytes_int",
			Binary: func(lhs ref.Val, rhs ref.Val) ref.Val {
				bV, ok := lhs.(types.Bytes)
				if !ok {
					return types.NewErr("unexpected type '%v' of bs passed to bit_in_bytes", lhs.Type())
				}
				pV, ok := rhs.(types.Int)
				if !ok {
					return types.NewErr("unexpected type '%v' of position passed to bit_in_bytes", rhs.Type())
				}
				bb, _ := bV.Value().([]byte)
				pos, _ := pV.Value().(int64)
				if len(bb) <= int(pos)/8 {
					return types.NewErr("position exceed of bs")
				}
				return types.Bool(bb[pos/8]&(1<<(pos%8)) != 0)
			}},
		&functions.Overload{
			Operator: "bytes_to_int_int",
			Unary: func(value ref.Val) ref.Val {
				bV, ok := value.(types.Bytes)
				if !ok {
					return types.NewErr("unexpected type '%v' of bs passed to bit_in_bytes", value.Type())
				}
				bb, _ := bV.Value().([]byte)
				bytebuff := bytes.NewBuffer(bb)
				switch len(bb) {
				case 1:
					var d int8
					binary.Read(bytebuff, binary.BigEndian, &d)
					return types.Int(d)
				case 2:
					var d int16
					binary.Read(bytebuff, binary.BigEndian, &d)
					return types.Int(d)
				case 3, 4:
					var d int32
					binary.Read(bytebuff, binary.BigEndian, &d)
					return types.Int(d)
				default:
					var d int64
					binary.Read(bytebuff, binary.BigEndian, &d)
					return types.Int(d)
				}
			}},
		&functions.Overload{
			Operator: "bytes_to_uint_uint",
			Unary: func(value ref.Val) ref.Val {
				bV, ok := value.(types.Bytes)
				if !ok {
					return types.NewErr("unexpected type '%v' of bs passed to bit_in_bytes", value.Type())
				}
				bb, _ := bV.Value().([]byte)
				bytebuff := bytes.NewBuffer(bb)
				switch len(bb) {
				case 1:
					var d uint8
					binary.Read(bytebuff, binary.BigEndian, &d)
					return types.Int(d)
				case 2:
					var d uint16
					binary.Read(bytebuff, binary.BigEndian, &d)
					return types.Int(d)
				case 3, 4:
					var d uint32
					binary.Read(bytebuff, binary.BigEndian, &d)
					return types.Int(d)
				default:
					var d uint64
					binary.Read(bytebuff, binary.BigEndian, &d)
					return types.Int(d)
				}
			}},
		// other functions
	)

	celEnv, err := cel.NewEnv(
		cel.Lib(&customLib{
			envOptions: []cel.EnvOption{
				cel.CustomTypeAdapter(&customTypeAdapter{}),
				cel.Declarations(declsVars...),
			},
		}),
	)
	if err != nil {
		return nil, err
	}
	return &CelRuntime{
		celEnv:      celEnv,
		celProgOpts: celProgOpts,
		declsVars:   declsVars,
		prgs:        make(map[string]cel.Program),
	}, nil
}

func (m *CelRuntime) RegProgram(key string, expr string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ast, issues := m.celEnv.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return "", issues.Err()
	} else {
		prg, err := m.celEnv.Program(ast, m.celProgOpts)
		if err != nil {
			return "", err
		}
		m.prgs[key] = prg
		return ast.OutputType().String(), nil
	}
}

func (m *CelRuntime) RegProgramType(key string, expr string) (*exprpb.Type, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	ast, issues := m.celEnv.Compile(expr)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	} else {
		prg, err := m.celEnv.Program(ast, m.celProgOpts)
		if err != nil {
			return nil, err
		}
		m.prgs[key] = prg
		return ast.ResultType(), nil
	}
}

func (m *CelRuntime) UpdateEnv(regVariables []*CelRegVarible) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.declsVars = nil
	for _, r := range regVariables {
		m.declsVars = append(m.declsVars, decls.NewVar(r.VarName, r.VarType))
	}
	newEnv, err := m.celEnv.Extend(cel.Lib(&customLib{
		envOptions: []cel.EnvOption{
			cel.Declarations(m.declsVars...),
		},
	}))
	if err != nil {
		return err
	}
	m.celEnv = newEnv
	return nil
}

func (m *CelRuntime) Eval(key string, vars map[string]interface{}) (ref.Val, error) {
	var prg cel.Program
	var ok bool
	m.mu.Lock()
	prg, ok = m.prgs[key]
	m.mu.Unlock()
	if !ok {
		return nil, fmt.Errorf("program %s not found", key)
	}
	out, _, err := prg.Eval(vars)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func ConvertDataType(dt string) (*exprpb.Type, error) {
	reg, _ := regexp.Compile(STRING_DT_REG)
	match := reg.FindStringSubmatch(dt)
	if match != nil {
		return decls.String, nil
	}
	switch dt {
	case "Bool", "bool":
		return decls.Bool, nil
	case "Byte", "Word", "DWord", "LWord", "bytes":
		return decls.Bytes, nil
	case "Char":
		return decls.String, nil
	case "SInt", "Int", "DInt", "LInt", "int":
		return decls.Int, nil
	case "USInt", "UInt", "UDInt", "ULInt", "uint", "Uint":
		return decls.Uint, nil
	case "Real", "LReal", "double", "float":
		return decls.Double, nil
	case "DTL", "Date", "Date_And_Time", "LDT", "LTime_Of_Day", "Time_Of_Day":
		return decls.Timestamp, nil
	case "S5Time", "Time", "LTime":
		return decls.Duration, nil
	case "string", "String", "WString":
		return decls.String, nil
	case "null_type":
		return decls.Null, nil
	default:
		return nil, fmt.Errorf("unsupported data type %s", dt)
	}
}
