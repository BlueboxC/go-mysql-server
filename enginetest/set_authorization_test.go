// Copyright 2026 Dolthub, Inc.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy at http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package enginetest_test

import (
	"testing"

	"github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/stretchr/testify/require"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest/setup"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/mysql_db"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
)

func TestSetGlobalAuthorization(t *testing.T) {
	h := enginetest.NewDefaultMemoryHarness()
	if h.IsUsingServer() {
		t.Skip("requires direct bound-plan execution and session inspection")
	}
	h.Setup(setup.MydbData)
	e := enginetest.MustNewEngine(t, h).(*sqle.Engine)
	defer e.Close()
	e.Analyzer.Catalog.MySQLDb.AddRootAccount()
	e.Analyzer.Catalog.MySQLDb.SetPersister(&mysql_db.NoopPersister{})
	root := enginetest.NewContextWithClient(h, sql.Client{User: "root", Address: "localhost"})
	run := func(ctx *sql.Context, query string) {
		t.Helper()
		enginetest.RunQueryWithContext(t, e, h, ctx, query)
	}
	run(root, "CREATE USER 'variable_admin'@'localhost'")
	run(root, "GRANT SELECT ON mydb.* TO 'variable_admin'@'localhost'")
	run(root, "GRANT SYSTEM_VARIABLES_ADMIN ON *.* TO 'variable_admin'@'localhost'")
	ctx := enginetest.NewContextWithClient(h, sql.Client{User: "variable_admin", Address: "localhost"})
	persisted := memory.GlobalsMap{}
	ctx.Session.(*memory.Session).SetGlobals(persisted)
	run(root, "SET GLOBAL max_connections = 37")
	run(ctx, "SET PERSIST_ONLY max_connections = 39")
	require.Equal(t, int64(39), persisted["max_connections"])
	enginetest.TestQueryWithContext(t, ctx, e, h, "SELECT @@GLOBAL.max_connections", []sql.Row{{int64(37)}}, nil, nil, nil)
	run(ctx, "SET PERSIST max_connections = 40")
	require.Equal(t, int64(40), persisted["max_connections"])

	// Bind as an administrator, copy the plan as prepared execution does, then revoke.
	query := "SET @set_marker = 9, @@GLOBAL.max_connections = @rhs_marker + 39"
	parsed, err := sqlparser.Parse(query)
	require.NoError(t, err)
	bound, err := e.BoundQueryPlan(ctx, query, parsed, nil)
	require.NoError(t, err)
	bound, err = analyzer.DeepCopyNode(ctx, bound)
	require.NoError(t, err)
	run(ctx, "SET @set_marker = 1, @rhs_marker = 2")
	run(root, "REVOKE SYSTEM_VARIABLES_ADMIN ON *.* FROM 'variable_admin'@'localhost'")
	_, iter, _, err := e.PrepQueryPlanForExecution(ctx, query, bound, &sql.QueryFlags{})
	if iter != nil {
		require.NoError(t, iter.Close(ctx))
	}
	require.True(t, sql.ErrPrivilegeCheckFailed.Is(err), "expected current-privilege rejection, got %v", err)
	enginetest.TestQueryWithContext(t, ctx, e, h, "SELECT @set_marker, @rhs_marker, @@GLOBAL.max_connections", []sql.Row{{int64(1), int64(2), int64(40)}}, nil, nil, nil)
	require.Equal(t, int64(40), persisted["max_connections"])
	for _, denied := range []string{
		"SET @set_marker = 9, @@GLOBAL.max_connections = @rhs_marker + 39",
		"SET PERSIST max_connections = 42",
		"SET PERSIST_ONLY max_connections = 43",
	} {
		enginetest.AssertErrWithCtx(t, e, h, ctx, denied, nil, sql.ErrPrivilegeCheckFailed)
	}
	require.Equal(t, int64(40), persisted["max_connections"])
	enginetest.TestQueryWithContext(t, ctx, e, h, "SELECT @set_marker, @rhs_marker, @@GLOBAL.max_connections", []sql.Row{{int64(1), int64(2), int64(40)}}, nil, nil, nil)

	// A new grant must take effect on the same session and retained plan.
	run(root, "GRANT SYSTEM_VARIABLES_ADMIN ON *.* TO 'variable_admin'@'localhost'")
	_, iter, _, err = e.PrepQueryPlanForExecution(ctx, query, bound, &sql.QueryFlags{})
	require.NoError(t, err)
	_, err = sql.RowIterToRows(ctx, iter)
	require.NoError(t, err)
	enginetest.TestQueryWithContext(t, ctx, e, h, "SELECT @set_marker, @rhs_marker, @@GLOBAL.max_connections", []sql.Row{{int64(9), int64(2), int64(41)}}, nil, nil, nil)
	// SUPER remains a compatible alternative to the dynamic privilege.
	run(root, "REVOKE SYSTEM_VARIABLES_ADMIN ON *.* FROM 'variable_admin'@'localhost'")
	run(root, "GRANT SUPER ON *.* TO 'variable_admin'@'localhost'")
	run(ctx, "SET GLOBAL max_connections = 44")
	run(root, "REVOKE SUPER ON *.* FROM 'variable_admin'@'localhost'")

	// Loading a stored trigger bypasses binding checks, not execution authorization.
	run(root, "CREATE TABLE variable_trigger (i INT)")
	run(root, "CREATE TRIGGER variable_guard BEFORE INSERT ON variable_trigger FOR EACH ROW SET @@GLOBAL.max_connections = 45")
	run(root, "GRANT INSERT ON mydb.* TO 'variable_admin'@'localhost'")
	enginetest.AssertErrWithCtx(t, e, h, ctx, "INSERT INTO variable_trigger VALUES (1)", nil, sql.ErrPrivilegeCheckFailed)
	enginetest.TestQueryWithContext(t, root, e, h, "SELECT @@GLOBAL.max_connections", []sql.Row{{int64(44)}}, nil, nil, nil)

	// Trusted integrators may explicitly disable auth when constructing their own plan.
	b := planbuilder.New(ctx, e.Analyzer.Catalog, nil)
	b.DisableAuth()
	trusted, _, _, flags, err := b.Parse("SET GLOBAL max_connections = 46", nil, false)
	require.NoError(t, err)
	_, iter, _, err = e.PrepQueryPlanForExecution(ctx, "", trusted, flags)
	require.NoError(t, err)
	_, err = sql.RowIterToRows(ctx, iter)
	require.NoError(t, err)
	enginetest.TestQueryWithContext(t, root, e, h, "SELECT @@GLOBAL.max_connections", []sql.Row{{int64(46)}}, nil, nil, nil)

	// Catalog-free builders also remain usable for standalone expression plans.
	b = planbuilder.New(ctx, nil, nil)
	_, _, _, _, err = b.Parse("SET @standalone = 1", nil, false)
	require.NoError(t, err)
}
