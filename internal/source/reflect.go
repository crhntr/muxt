package source

import (
	"fmt"
	"go/ast"
	"reflect"
	"strconv"
)

func ParseStringWithType(val string, tp ast.Expr) (reflect.Value, error) {
	tpIdent, ok := tp.(*ast.Ident)
	if !ok {
		return reflect.Value{}, fmt.Errorf("type %s is not supported", Format(tp))
	}
	switch tpIdent.Name {
	case reflect.Int.String():
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(int(n)), nil
	case reflect.Int8.String():
		n, err := strconv.ParseInt(val, 10, 8)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(int8(n)), nil
	case reflect.Int16.String():
		n, err := strconv.ParseInt(val, 10, 16)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(int16(n)), nil
	case reflect.Int32.String():
		n, err := strconv.ParseInt(val, 10, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(int32(n)), nil
	case reflect.Int64.String():
		n, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(n), nil
	case reflect.Uint.String():
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(uint(n)), nil
	case reflect.Uint8.String():
		n, err := strconv.ParseUint(val, 10, 8)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(uint8(n)), nil
	case reflect.Uint16.String():
		n, err := strconv.ParseUint(val, 10, 16)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(uint16(n)), nil
	case reflect.Uint32.String():
		n, err := strconv.ParseUint(val, 10, 32)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(uint32(n)), nil
	case reflect.Uint64.String():
		n, err := strconv.ParseUint(val, 10, 64)
		if err != nil {
			return reflect.Value{}, err
		}
		return reflect.ValueOf(n), nil
	default:
		return reflect.Value{}, fmt.Errorf("type %s unknown", Format(tp))
	}
}