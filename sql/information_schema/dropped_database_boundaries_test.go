// Copyright 2026 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package information_schema

import (
	"context"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/mysql_db"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/stretchr/testify/require"
)

type disappearingRoutineCatalog struct {
	sql.Catalog
	failure error
}

func (c disappearingRoutineCatalog) Database(_ *sql.Context, name string) (sql.Database, error) {
	if name == "gone" {
		return nil, c.failure
	}
	return sql.UnresolvedDatabase(name), nil
}

func (c disappearingRoutineCatalog) AuthorizationHandler() sql.AuthorizationHandler {
	return sql.NoopAuthorizationHandlerFactory{}.CreateHandler(c)
}

func (c disappearingRoutineCatalog) Overrides() sql.EngineOverrides {
	return sql.EngineOverrides{}
}

type tableMetadataFailure struct {
	sql.Database
	failure error
}

func (d tableMetadataFailure) GetTableNames(*sql.Context) ([]string, error) { return nil, d.failure }

type viewMetadataFailure struct {
	sql.Database
	sql.ViewDatabase
	failure error
}

func (d viewMetadataFailure) Name() string { return "healthy" }

func (d viewMetadataFailure) AllViews(*sql.Context) ([]sql.ViewDefinition, error) {
	return nil, d.failure
}

type triggerMetadataFailure struct {
	sql.TriggerDatabase
	failure error
}

func (d triggerMetadataFailure) GetTriggers(*sql.Context) ([]sql.TriggerDefinition, error) {
	return nil, d.failure
}

type eventMetadataFailure struct {
	sql.EventDatabase
	failure error
}

func (d eventMetadataFailure) GetEvents(*sql.Context) ([]sql.EventDefinition, interface{}, error) {
	return nil, nil, d.failure
}

func TestCatalogEnumerationPreservesUnexpectedErrors(t *testing.T) {
	ctx := sql.NewContext(context.Background())
	failure := sql.ErrInvalidType.New("metadata sentinel")
	tableDB := tableMetadataFailure{failure: failure}
	checks := map[string]func() error{
		"table_iteration":  func() error { return iterDBTables(ctx, tableDB, func(sql.Table) (bool, error) { return true, nil }) },
		"table_resolution": func() error { return forEachDBTable(ctx, nil, tableDB, func(sql.Table) error { return nil }) },
		"views":            func() error { _, err := dbViews(ctx, viewMetadataFailure{failure: failure}); return err },
		"triggers":         func() error { _, err := dbTriggers(ctx, triggerMetadataFailure{failure: failure}); return err },
		"events":           func() error { _, err := dbEvents(ctx, eventMetadataFailure{failure: failure}); return err },
	}
	for name, check := range checks {
		t.Run(name, func(t *testing.T) { err := check(); require.Error(t, err); require.True(t, sql.ErrInvalidType.Is(err)) })
	}
}

func TestRoutinesCatalogDroppedDatabaseBoundary(t *testing.T) {
	for _, gone := range []bool{true, false} {
		t.Run(map[bool]string{true: "dropped", false: "real_error"}[gone], func(t *testing.T) {
			session := sql.NewBaseSession()
			privileges := mysql_db.NewPrivilegeSet()
			privileges.AddGlobalStatic(sql.PrivilegeType_CreateRoutine)
			session.SetPrivilegeSet(privileges, 0)
			ctx := sql.NewContext(context.Background(), sql.WithSession(session))
			ctx.SetCurrentDatabase("healthy")
			failure := sql.ErrInvalidType.New("metadata sentinel")
			if gone {
				failure = sql.ErrDatabaseNotFound.New("gone")
			}
			catalog := disappearingRoutineCatalog{failure: failure}
			procedures := map[string][]*plan.Procedure{
				"gone":    {},
				"healthy": {{Name: "kept_p", CreateProcedureString: "CREATE PROCEDURE kept_p() SELECT 1"}},
			}
			iter, err := routinesRowIter(ctx, catalog, procedures)
			if !gone {
				require.Error(t, err)
				require.True(t, sql.ErrInvalidType.Is(err))
				return
			}
			require.NoError(t, err)
			rows, err := sql.RowIterToRows(ctx, iter)
			require.NoError(t, err)
			require.Len(t, rows, 1)
			require.Equal(t, "kept_p", rows[0][0])
		})
	}
}
