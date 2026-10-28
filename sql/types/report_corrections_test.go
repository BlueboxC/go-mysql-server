package types

import (
	"github.com/dolthub/go-mysql-server/sql"
	"reflect"
	"testing"
	"time"
)

func TestReportBooleanYearEnumSet(t *testing.T) {
	ctx := sql.NewEmptyContext()
	for _, typ := range []sql.Type{Year, MustCreateEnumType([]string{"x", "y"}, sql.Collation_Default), MustCreateSetType([]string{"x", "y"}, sql.Collation_Default)} {
		for _, v := range []bool{false, true} {
			n := 0
			if v {
				n = 1
			}
			got, gr, ge := typ.Convert(ctx, v)
			want, wr, we := typ.Convert(ctx, n)
			if !reflect.DeepEqual(got, want) || gr != wr || (ge == nil) != (we == nil) {
				t.Fatalf("%s %t: %v/%v vs integer %v/%v", typ, v, got, ge, want, we)
			}
		}
		if _, _, err := typ.Convert(ctx, "true"); err == nil {
			t.Fatalf("%s accepted text true", typ)
		}
	}
}

func TestReportZeroDateNumericConversions(t *testing.T) {
	for _, typ := range []sql.DatetimeType{Date, Datetime, Timestamp} {
		f, err := typ.ToFloat64(ZeroTime)
		if err != nil || f != 0 {
			t.Fatalf("%s zero float: %v %v", typ, f, err)
		}
		d, err := typ.ToDecimal(ZeroTime)
		if err != nil || !d.IsZero() {
			t.Fatalf("%s zero decimal: %v %v", typ, d, err)
		}
	}
	ordinary := time.Date(2020, 1, 2, 3, 4, 5, 0, time.UTC)
	f, err := Datetime.ToFloat64(ordinary)
	if err != nil || f != 20200102030405 {
		t.Fatalf("ordinary datetime: %v %v", f, err)
	}
	if DateToNumber(ZeroTime) != 0 || DateTimeToNumber(ZeroTime) != 0 || DateToNumber(ordinary) != 20200102 {
		t.Fatal("date numeric control")
	}
}
