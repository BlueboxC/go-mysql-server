// Copyright 2026 BlueboxC
// Licensed under the Apache License, Version 2.0.
package enginetest_test

import (
	"fmt"
	"github.com/dolthub/go-mysql-server/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest/queries"
	"testing"
)

func TestForkJoinPredicatePreservation(t *testing.T) {
	for _, prepared := range []bool{false, true} {
		t.Run(fmt.Sprintf("prepared=%v", prepared), func(t *testing.T) {
			for _, script := range queries.JoinScriptTests {
				if script.Name != "join reordering does not drop filters of large cross join" {
					continue
				}
				h := enginetest.NewDefaultMemoryHarness()
				if prepared {
					enginetest.TestScriptPrepared(t, h, script)
				} else {
					enginetest.TestScript(t, h, script)
				}
			}
		})
	}
}
