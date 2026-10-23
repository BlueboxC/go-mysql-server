// Copyright 2026 BlueboxC
// Licensed under the Apache License, Version 2.0.
package enginetest_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest/queries"
	"github.com/dolthub/go-mysql-server/enginetest/scriptgen/setup"
)

func TestForkGreatestLeastTypes(t *testing.T) {
	for _, prepared := range []bool{false, true} {
		t.Run(fmt.Sprintf("prepared=%v", prepared), func(t *testing.T) {
			for _, s := range queries.GreatestLeastTypeForkTests {
				h := enginetest.NewDefaultMemoryHarness()
				if prepared {
					enginetest.TestScriptPrepared(t, h, s)
				} else {
					enginetest.TestScript(t, h, s)
				}
			}
		})
	}
}

func TestForkComparisonFunctionQueries(t *testing.T) {
	for _, prepared := range []bool{false, true} {
		t.Run(fmt.Sprintf("prepared=%v", prepared), func(t *testing.T) {
			h := enginetest.NewDefaultMemoryHarness()
			h.Setup(setup.SimpleSetup...)
			e := enginetest.MustNewEngine(t, h)
			defer e.Close()
			for _, q := range queries.FunctionQueryTests {
				upper := strings.ToUpper(q.Query)
				if !strings.Contains(upper, "GREATEST(") && !strings.Contains(upper, "LEAST(") {
					continue
				}
				if q.Skip || (prepared && q.SkipPrepared) {
					t.Run(q.Query, func(t *testing.T) { t.Skip("upstream unsupported case") })
					continue
				}
				if prepared {
					enginetest.TestPreparedQueryWithEngine(t, h, e, q)
				} else {
					enginetest.TestQuery(t, h, e, q)
				}
			}
		})
	}
}
