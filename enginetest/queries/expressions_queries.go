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

package queries

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
)

// ExpressionsScriptTests contains self-contained expressions script tests.
var ExpressionsScriptTests = []ScriptTest{
	{
		Name:        "fork subquery unnesting preserves other negations",
		SetUpScript: []string{"CREATE TABLE t(id INT PRIMARY KEY,kind VARCHAR(20))", "INSERT INTO t VALUES(1,'task'),(2,'template')", "CREATE TABLE labels(id INT)", "INSERT INTO labels VALUES(1),(2)", "CREATE TABLE empty_scalar_rows(id INT)"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT id FROM t WHERE kind<>'template' AND id IN(SELECT id FROM labels)", Expected: []sql.Row{{1}}},
			{Query: "SELECT id FROM t WHERE NOT(kind='template') AND id IN(SELECT id FROM labels)", Expected: []sql.Row{{1}}},
			{Query: "SELECT id FROM t WHERE id<>(SELECT id FROM empty_scalar_rows) AND id IN(SELECT id FROM labels)", Expected: []sql.Row{}},
		},
	},

	{
		Name:        "fork join predicates require every referenced relation",
		SetUpScript: []string{"CREATE TABLE t0(x INT)", "INSERT INTO t0 VALUES(1),(2)", "CREATE TABLE t1(y INT)", "INSERT INTO t1 VALUES(1),(NULL)", "CREATE TABLE t2(z INT)", "INSERT INTO t2 VALUES(1),(0)", "CREATE VIEW v2 AS SELECT z FROM t2", "CREATE TABLE a(c0 BOOLEAN PRIMARY KEY,c1 INT)", "CREATE TABLE b(c0 INT,c1 BOOLEAN,c2 BOOLEAN,PRIMARY KEY(c0,c1))", "INSERT INTO b VALUES(1,false,false),(2,true,true),(3,false,NULL),(4,true,NULL)", "INSERT INTO a VALUES(true,1),(false,NULL)", "CREATE VIEW va(c0) AS SELECT a.c1 FROM a"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT COUNT(*) FROM t0,t1,v2 WHERE (x=y) OR (v2.z<>1)", Expected: []sql.Row{{5}}},
			{Query: "SELECT * FROM va,b,a WHERE (va.c0=b.c0)<=(CASE WHEN a.c0 THEN b.c2 END)", Expected: []sql.Row{{1, 2, 1, 1, 1, 1}}},
		},
	},
	{
		Name:        "fork NULL extended relation remains in predicate eligibility",
		SetUpScript: []string{"CREATE TABLE a(x BOOLEAN)", "INSERT INTO a VALUES(0)", "CREATE TABLE b(y BOOLEAN)", "INSERT INTO b VALUES(1)", "CREATE TABLE e(z BOOLEAN)"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT * FROM a,b LEFT JOIN e ON(1=0) WHERE (a.x<b.y) AND (a.x<e.z)", Expected: []sql.Row{}},
			{Query: "SELECT * FROM a,b LEFT JOIN e ON(1=0) WHERE a.x<e.z", Expected: []sql.Row{}},
			{Query: "SELECT * FROM a,b LEFT JOIN e ON(1=0) WHERE ((a.x<b.y) AND (a.x<e.z)) IS NULL", Expected: []sql.Row{{0, 1, nil}}},
		},
	},

	{
		Name:        "fork indexed comparisons retain the original value",
		SetUpScript: []string{"CREATE TABLE t0(c0 INT PRIMARY KEY)", "INSERT INTO t0 VALUES(1),(2)", "CREATE TABLE t1(c0 FLOAT PRIMARY KEY)", "INSERT INTO t1 VALUES(0.89451053545548)", "CREATE TABLE t2(c0 TIMESTAMP PRIMARY KEY)", "INSERT INTO t2 VALUES('1970-01-23 10:07:51')", "CREATE TABLE t3(c0 TIME PRIMARY KEY)", "INSERT INTO t3 VALUES('12:00:00')", "CREATE TABLE t4(c0 DECIMAL(10,2) PRIMARY KEY)", "INSERT INTO t4 VALUES(1.00)"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT c0 FROM t0 WHERE c0<>'1.5' ORDER BY c0", Expected: []sql.Row{{1}, {2}}},
			{Query: "SELECT COUNT(*) FROM t1 WHERE c0<>CAST(0.89451053545548 AS FLOAT)", Expected: []sql.Row{{0}}},
			{Query: "SELECT COUNT(*) FROM t2 WHERE c0>'1969-12-18 02:32:34'", Expected: []sql.Row{{1}}},
			{Query: "SELECT COUNT(*) FROM t3 WHERE c0<>2411793", Expected: []sql.Row{{1}}},
			{Query: "SELECT COUNT(*) FROM t4 WHERE c0=1.001", Expected: []sql.Row{{0}}},
			{Query: "SELECT * FROM t4 WHERE 1.001=c0", Expected: []sql.Row{}},
			{Query: "SELECT * FROM t4 WHERE c0=1.00", Expected: []sql.Row{{"1.00"}}},
			{Query: "SELECT * FROM t4 WHERE c0<1.001", Expected: []sql.Row{{"1.00"}}},
			{Query: "SELECT c0 FROM t0 WHERE c0='1'", Expected: []sql.Row{{1}}},
		},
	},
	{
		Name:        "fork explicit collation cannot reuse incompatible index ordering",
		SetUpScript: []string{"CREATE TABLE t0(s VARCHAR(20) COLLATE utf8mb4_bin PRIMARY KEY)", "INSERT INTO t0 VALUES('a'),('B')"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT s FROM t0 WHERE s=('A' COLLATE utf8mb4_general_ci)", Expected: []sql.Row{{"a"}}},
			{Query: "SELECT s FROM t0 WHERE ('A' COLLATE utf8mb4_general_ci)=s", Expected: []sql.Row{{"a"}}},
			{Query: "SELECT s FROM t0 WHERE s='a'", Expected: []sql.Row{{"a"}}},
			{Query: "SELECT s FROM t0 WHERE s='A'", Expected: []sql.Row{}},
		},
	},

	{
		Name:        "fork unique NULL join keys obey comparison semantics",
		SetUpScript: []string{"CREATE TABLE t0(c0 INT)", "CREATE TABLE t1(c0 INT UNIQUE)", "INSERT INTO t0 VALUES(NULL),(1),(2)", "INSERT INTO t1 VALUES(NULL),(1),(2)"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT * FROM t0 JOIN t1 ON t0.c0=t1.c0 ORDER BY t0.c0", Expected: []sql.Row{{1, 1}, {2, 2}}},
			{Query: "SELECT /*+ LOOKUP_JOIN(t0,t1) */ * FROM t0 JOIN t1 ON t0.c0=t1.c0 ORDER BY t0.c0", Expected: []sql.Row{{1, 1}, {2, 2}}},
			{Query: "SELECT /*+ LOOKUP_JOIN(t0,t1) */ * FROM t0 JOIN t1 ON t0.c0<=>t1.c0 ORDER BY t0.c0", Expected: []sql.Row{{nil, nil}, {1, 1}, {2, 2}}},
			{Query: "SELECT /*+ LOOKUP_JOIN(t0,t1) */ * FROM t0 LEFT JOIN t1 ON t0.c0=t1.c0 ORDER BY t0.c0", Expected: []sql.Row{{nil, nil}, {1, 1}, {2, 2}}},
		},
	},
	{
		Name:        "fork mixed IN subquery uses SQL comparison",
		SetUpScript: []string{"CREATE TABLE t0(c0 INT)", "INSERT INTO t0 VALUES(2),(NULL)"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT false IN(SELECT 'W1q'),false NOT IN(SELECT 'W1q')", Expected: []sql.Row{{true, false}}},
			{Query: "SELECT 'a' IN(SELECT c0 FROM t0),'a' NOT IN(SELECT c0 FROM t0)", Expected: []sql.Row{{nil, nil}}},
			{Query: "SELECT NULL IN(SELECT c0 FROM t0 WHERE false),NULL NOT IN(SELECT c0 FROM t0 WHERE false)", Expected: []sql.Row{{false, true}}},
			{Query: "SELECT 2 IN(SELECT c0 FROM t0),2 NOT IN(SELECT c0 FROM t0)", Expected: []sql.Row{{true, false}}},
		},
	},
	{
		Name:        "fork negated scalar subquery retains cardinality and NULL",
		SetUpScript: []string{"CREATE TABLE t0(c0 INT)", "CREATE TABLE t1(c0 INT)", "INSERT INTO t1 VALUES(1)", "CREATE TABLE t2(c0 INT)", "INSERT INTO t2 VALUES(2)", "CREATE TABLE t3(c0 INT)", "INSERT INTO t3 VALUES(2),(3)"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT * FROM t1 WHERE c0<>(SELECT c0 FROM t0)", Expected: []sql.Row{}},
			{Query: "SELECT * FROM t1 WHERE c0<>(SELECT c0 FROM t2)", Expected: []sql.Row{{1}}},
			{Query: "SELECT * FROM t1 WHERE c0<>(SELECT c0 FROM t3)", ExpectedErr: sql.ErrExpectedSingleRow},
		},
	},

	{
		Name:        "fork range heap preserves input filters",
		SetUpScript: []string{"CREATE TABLE t0(a INT PRIMARY KEY, w INT)", "INSERT INTO t0 VALUES (1,0),(2,1)", "CREATE TABLE t1(lo INT, hi INT)", "INSERT INTO t1 VALUES (0,10)"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT t0.a,t0.w FROM t0 JOIN t1 ON t0.a BETWEEN t1.lo AND t1.hi WHERE t0.w=1", Expected: []sql.Row{{2, 1}}},
			{Query: "SELECT t0.a,t0.w FROM t0 LEFT JOIN t1 ON t0.a BETWEEN t1.lo AND t1.hi WHERE t0.w=1", Expected: []sql.Row{{2, 1}}},
		},
	},
	{
		Name:        "fork fulltext returns each keyed document once",
		SetUpScript: []string{"CREATE TABLE t0(c0 INT PRIMARY KEY,c1 TEXT,c2 TEXT,FULLTEXT ft(c1,c2))", "INSERT INTO t0 VALUES (1,'abc def','ghi jkl'),(2,'abc','mno'),(3,'xyz','pqr'),(4,'abc def','ghi jkl')", "CREATE TABLE t1(c1 TEXT,FULLTEXT ft(c1))", "INSERT INTO t1 VALUES ('abc def'),('abc def'),('xyz')"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT c0 FROM t0 WHERE MATCH(c1,c2) AGAINST ('abc def') ORDER BY c0", Expected: []sql.Row{{1}, {2}, {4}}},
			{Query: "SELECT c0 FROM t0 WHERE MATCH(c1,c2) AGAINST ('abc') ORDER BY c0", Expected: []sql.Row{{1}, {2}, {4}}},
			{Query: "SELECT c1 FROM t1 WHERE MATCH(c1) AGAINST ('abc def')", Expected: []sql.Row{{"abc def"}, {"abc def"}}},
		},
	},
	{
		Name:        "fork fulltext is not an ordered column index",
		SetUpScript: []string{"CREATE TABLE t0(c0 INT,c1 VARCHAR(100))", "CREATE FULLTEXT INDEX ft ON t0(c1)", "INSERT INTO t0 VALUES (1,'b'),(2,'a')"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT * FROM t0 ORDER BY c1", Expected: []sql.Row{{2, "a"}, {1, "b"}}},
			{Query: "SELECT * FROM t0 ORDER BY c1 DESC", Expected: []sql.Row{{1, "b"}, {2, "a"}}},
		},
	},
	{
		Name:        "fork decimal comparisons retain both scales",
		SetUpScript: []string{"CREATE TABLE t0(c0 DECIMAL(10,2))", "INSERT INTO t0 VALUES (1.00)"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT c0=1.001,1.001=c0,c0<1.001,1.001>c0 FROM t0", Expected: []sql.Row{{false, false, true, true}}},
			{Query: "SELECT * FROM t0 WHERE c0=1.001", Expected: []sql.Row{}},
			{Query: "SELECT CAST(2 AS DECIMAL(20,0))>1.5,1.5<CAST(2 AS DECIMAL(20,0))", Expected: []sql.Row{{true, true}}},
		},
	},
	{
		Name:        "fork LIKE escape rejects row-dependent expressions",
		SetUpScript: []string{"CREATE TABLE t0(c0 INT,c1 INT)", "INSERT INTO t0 VALUES (1,5),(2,0),(3,NULL),(NULL,7)"},
		Assertions: []ScriptTestAssertion{
			{Query: "SELECT * FROM t0 WHERE c1 LIKE c0 ESCAPE c0", ExpectedErr: sql.ErrInvalidArgument},
			{Query: "SELECT * FROM t0 WHERE c1 LIKE c0 ESCAPE c0 UNION ALL SELECT * FROM t0 WHERE NOT(c1 LIKE c0 ESCAPE c0)", ExpectedErr: sql.ErrInvalidArgument},
			{Query: "SELECT 'a_b' LIKE 'a!_b' ESCAPE '!', 'a_b' LIKE 'a!_b' ESCAPE CONCAT('','!')", Expected: []sql.Row{{true, true}}},
		},
	},

	{
		Name:    "coalesce tests",
		Dialect: "mysql",
		SetUpScript: []string{
			"create table c select coalesce(NULL, 1);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "select * from c;",
				Expected: []sql.Row{
					{1},
				},
			},
			{
				Query: "select COLUMN_NAME, DATA_TYPE from INFORMATION_SCHEMA.COLUMNS where TABLE_NAME='c';",
				Expected: []sql.Row{
					{"coalesce(NULL, 1)", "int"},
				},
			},
		},
	},
	{
		Name: "different cases of function name should result in the same outcome",
		SetUpScript: []string{
			"create table t (b binary(2) primary key);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:       "select hex(*) from t;",
				ExpectedErr: sql.ErrStarUnsupported,
			},
			{
				Query:       "select HEX(*) from t;",
				ExpectedErr: sql.ErrStarUnsupported,
			},
			{
				Query:       "select HeX(*) from t;",
				ExpectedErr: sql.ErrStarUnsupported,
			},
		},
	},
	{
		Name:    "name_const queries",
		Dialect: "mysql",
		SetUpScript: []string{
			"create table t (i int primary key);",
			"insert into t values (1), (2), (3);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "select name_const(123, 123)",
				ExpectedColumns: sql.Schema{
					{Name: "123", Type: types.Int8, Nullable: false},
				},
				Expected: []sql.Row{
					{123},
				},
			},
			{
				Query: "select name_const('abc', 123)",
				ExpectedColumns: sql.Schema{
					{Name: "abc", Type: types.Int8, Nullable: false},
				},
				Expected: []sql.Row{
					{123},
				},
			},
			{
				Query: "select name_const('abc', 'abc')",
				ExpectedColumns: sql.Schema{
					{Name: "abc", Type: types.Text, Nullable: false},
				},
				Expected: []sql.Row{
					{"abc"},
				},
			},
			{
				Query: "select name_const(date '2000-01-02', 123)",
				ExpectedColumns: sql.Schema{
					{Name: "2000-01-02", Type: types.Int8, Nullable: false},
				},
				Expected: []sql.Row{
					{123},
				},
			},
			{
				Query: "select name_const('abc', date '2001-02-03')",
				ExpectedColumns: sql.Schema{
					{Name: "abc", Type: types.Text, Nullable: false},
				},
				Expected: []sql.Row{
					{"2001-02-03"},
				},
			},

			{
				Query:          "select name_const('abc', 1+1)",
				ExpectedErrStr: "incorrect arguments to: NAME_CONST",
			},
			{
				Query:          "select name_const(1+1, 123)",
				ExpectedErrStr: "incorrect arguments to: NAME_CONST",
			},
			{
				Query:          "select name_const(i, 123) from t",
				ExpectedErrStr: "incorrect arguments to: NAME_CONST",
			},
			{
				Query:          "select name_const(123, i) from t",
				ExpectedErrStr: "incorrect arguments to: NAME_CONST",
			},
			{
				Query:          "select name_const()",
				ExpectedErrStr: "incorrect parameter count in the call to native function NAME_CONST",
			},
			{
				Query:          "select name_const(1)",
				ExpectedErrStr: "incorrect parameter count in the call to native function NAME_CONST",
			},
			{
				Query:          "select name_const(1, 2, 3)",
				ExpectedErrStr: "incorrect parameter count in the call to native function NAME_CONST",
			},
		},
	},
	{
		Name: "coalesce with system types",
		SetUpScript: []string{
			"create table t as select @@admin_port as port1, @@port as port2, COALESCE(@@admin_port, @@port) as\n port3;",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "describe t;",
				Expected: []sql.Row{
					{"port1", "bigint", "YES", "", nil, ""},
					{"port2", "bigint", "YES", "", nil, ""},
					{"port3", "bigint", "YES", "", nil, ""},
				},
			},
		},
	},

	{
		Name:    "not expression optimization",
		Dialect: "mysql",
		SetUpScript: []string{
			"create table t (i int);",
			"insert into t values (123);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "select * from t where 1 = (not(not(i)))",
				Expected: []sql.Row{
					{123},
				},
			},
			{
				Query: "select * from t where true = (not(not(i)))",
				Expected: []sql.Row{
					{123},
				},
			},
			{
				Query: "select * from t where true = (not(not(i = 123)))",
				Expected: []sql.Row{
					{123},
				},
			},
			{
				Query: "select * from t where false = (not(not(i != 123)))",
				Expected: []sql.Row{
					{123},
				},
			},
			{
				Query: "select * from t where i != (false or i);",
				Expected: []sql.Row{
					{123},
				},
			},
			{
				Query: "select * from t where ((true and -1) >= 0);",
				Expected: []sql.Row{
					{123},
				},
			},
		},
	},
	{
		Name: "tinyint column does not restrict IF or IFNULL output",
		// https://github.com/dolthub/dolt/issues/9321
		SetUpScript: []string{
			"create table t0 (c0 tinyint);",
			"insert into t0 values (null);",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: "select ifnull(t0.c0, 128) as ref0 from t0",
				Expected: []sql.Row{
					{128},
				},
			},
			{
				Query:    "select if(t0.c0 = 1, t0.c0, 128) as ref0 from t0",
				Expected: []sql.Row{{128}},
			},
		},
	},
	{
		// https://github.com/dolthub/dolt/issues/10243
		Dialect: "mysql",
		Name:    "OR filters are simplified to correct type",
		SetUpScript: []string{
			"create table t0(c1 boolean)",
			"insert into t0 values (true)",
			"create table t1(c1 int)",
			"insert into t1 values (2)",
		},
		Assertions: []ScriptTestAssertion{
			{
				Query:    "select 1 from t0 where (19 or 's') != (7 != t0.c1);",
				Expected: []sql.Row{},
			},
			{
				Query:    "SELECT * FROM t1 WHERE NOT EXISTS (SELECT 1 FROM t0 WHERE ((((19)OR('s')))!=(((7)!=(t0.c1)))));",
				Expected: []sql.Row{{2}},
			},
		},
	},
	{
		Name: "Between filter",
		SetUpScript: []string{
			"create table test(x int, y int, z int);",
			`insert into test values
                     (null, 0, 0),
                     (1, null, 1),
                     (2, 2, null),
                     (3, 2, 4),
                     (4, 2, 3),
                     (5, 6, 7),
                     (6, 6, 5),
                     (7, 8, 7),
                     (8, 9, 7)`,
		},
		Assertions: []ScriptTestAssertion{
			{
				Query: `select x, y, z,
       						(x between y and z), (x between x and z), (x between y and x), (x between y and y),
       						(x between x and x) from test order by x`,
				Expected: []sql.Row{
					{nil, 0, 0, nil, nil, nil, nil, nil},
					{1, nil, 1, nil, true, nil, nil, true},
					{2, 2, nil, nil, nil, true, true, true},
					{3, 2, 4, true, true, true, false, true},
					{4, 2, 3, false, false, true, false, true},
					{5, 6, 7, false, true, false, false, true},
					{6, 6, 5, false, false, true, true, true},
					{7, 8, 7, false, true, false, false, true},
					{8, 9, 7, false, false, false, false, true},
				},
			},
			{
				Query:    "select x from test where (x between y and z) order by x",
				Expected: []sql.Row{{3}},
			},
			{
				Query:    "select x from test where (x between x and z) order by x",
				Expected: []sql.Row{{1}, {3}, {5}, {7}},
			},
			{
				Query:    "select x from test where (x between y and x) order by x",
				Expected: []sql.Row{{2}, {3}, {4}, {6}},
			},
			{
				Query:    "select x from test where (x between y and y) order by x",
				Expected: []sql.Row{{2}, {6}},
			},
			{
				Query:    "select x from test where (x between x and x) order by x",
				Expected: []sql.Row{{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}},
			},
		},
	},
}
