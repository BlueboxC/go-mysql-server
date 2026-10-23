// Copyright 2026 BlueboxC
// Licensed under the Apache License, Version 2.0.
package queries

import "github.com/dolthub/go-mysql-server/sql"

var GreatestLeastTypeForkTests = []ScriptTest{{Name: "greatest and least preserve comparison types", Assertions: []ScriptTestAssertion{
	{Query: "SELECT GREATEST(1,2,'9','foo999'), LEAST(1,2,'9','foo999')", Expected: []sql.Row{{"foo999", "1"}}},
	{Query: "SELECT GREATEST('foo999','9',2,1), LEAST('foo999','9',2,1)", Expected: []sql.Row{{"foo999", "1"}}},
	{Query: "SELECT CAST(GREATEST(CAST(1 AS SIGNED),CAST('18446744073709551615' AS UNSIGNED)) AS CHAR)", Expected: []sql.Row{{"18446744073709551615"}}},
	{Query: "SELECT CAST(GREATEST(CAST('9007199254740993' AS UNSIGNED),CAST('9007199254740992' AS UNSIGNED)) AS CHAR)", Expected: []sql.Row{{"9007199254740993"}}},
	{Query: "SELECT GREATEST('2020-01-01 00:00:00',CAST('2021-01-01 00:00:00' AS DATETIME))", Expected: []sql.Row{{"2021-01-01 00:00:00"}}},
	{Query: "SELECT LEAST(CAST('2021-01-01 00:00:00' AS DATETIME),'2020-01-01 00:00:00')", Expected: []sql.Row{{"2020-01-01 00:00:00"}}},
	{Query: "SELECT GREATEST(1,NULL,'2'), LEAST(NULL,2,1)", Expected: []sql.Row{{nil, nil}}},
}}}
