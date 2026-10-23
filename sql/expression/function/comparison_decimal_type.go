// Copyright 2026 BlueboxC
// Licensed under the Apache License, Version 2.0.
package function

import (
	"fmt"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/types"
)

// compDecimalType combines exact numeric metadata without narrowing evaluated values.
func compDecimalType(args []sql.Expression, argTypes []sql.Type) sql.Type {
	var integerDigits, scale uint8
	for i, t := range argTypes {
		if t == nil || t == types.Null {
			continue
		}
		var p, s uint8
		if dt, ok := t.(sql.DecimalType); ok {
			p, s = dt.Precision(), dt.Scale()
		} else {
			switch t {
			case types.Int8, types.Uint8, types.Boolean:
				p = 3
			case types.Int16, types.Uint16:
				p = 5
			case types.Int24, types.Uint24:
				p = 8
			case types.Int32, types.Uint32:
				p = 10
			case types.Int64:
				p = 19
			case types.Uint64:
				p = 20
			default:
				return types.Float64
			}
		}

		if types.IsInteger(t) {
			p = compIntegerPrecision(args[i], t, p)
		}

		integerDigits = max(integerDigits, p-s)
		scale = max(scale, s)
	}

	scale = min(scale, types.DecimalTypeMaxScale)
	precision := min(uint16(integerDigits)+uint16(scale), types.DecimalTypeMaxPrecision)
	return types.MustCreateDecimalType(uint8(precision), scale)
}

// compIntegerPrecision distinguishes literal widths and CAST metadata from
// integer column widths, which cannot be inferred from the SQL type alone.
func compIntegerPrecision(arg sql.Expression, typ sql.Type, columnPrecision uint8) uint8 {
	switch arg := arg.(type) {
	case *expression.Literal:
		if _, ok := arg.Value().(bool); ok {
			return 1
		}
		var buf [types.DecimalTypeMaxPrecision]byte
		b := fmt.Append(buf[:0], arg.Value())
		if len(b) > 0 && b[0] == '-' {
			b = b[1:]
		}
		return uint8(len(b))
	case *expression.UnaryMinus:
		return compIntegerPrecision(arg.Child, typ, columnPrecision)
	case *expression.Convert:
		if types.IsUnsigned(typ) {
			return 21
		}

		return 20
	default:
		return columnPrecision
	}
}
