// Copyright 2026 BlueboxC
// Licensed under the Apache License, Version 2.0.

package queries

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
)

var FulltextOrdinaryFilterTests = func() []ScriptTest {
	var scripts []ScriptTest
	for _, keyed := range []bool{false, true} {
		for _, ordinary := range []bool{false, true} {
			primary, extra, columns, values := "", "", "a,b,c", "('alpha',1,'echo'),('alpha',1,'echo'),(NULL,2,'echo'),('',3,'omega'),('bravo',4,NULL),('alpha bravo',5,'omega')"
			if keyed {
				primary = "id INT PRIMARY KEY, "
				columns = "id,a,b,c"
				values = "(1,'alpha',1,'echo'),(2,'alpha',1,'echo'),(3,NULL,2,'echo'),(4,'',3,'omega'),(5,'bravo',4,NULL),(6,'alpha bravo',5,'omega')"
			}
			if ordinary {
				extra = ", INDEX ia(a)"
			}
			script := ScriptTest{Name: fmt.Sprintf("ordinary filters beside fulltext/keyed=%v/ordinary=%v", keyed, ordinary), SetUpScript: []string{
				"CREATE TABLE t (" + primary + "a VARCHAR(100), b INT, c VARCHAR(100), FULLTEXT INDEX fa(a), FULLTEXT INDEX fc(c), INDEX ib(b)" + extra + ")",
				"INSERT INTO t (" + columns + ") VALUES " + values,
			}}
			for _, test := range []struct {
				predicate string
				count     int64
			}{
				{"a IS NULL", 1}, {"a IS NOT NULL", 5}, {"a='alpha'", 2}, {"a<=>NULL", 1}, {"a<>'alpha'", 3}, {"a>'alpha'", 2}, {"a LIKE 'alpha%'", 3}, {"a IN ('alpha','bravo')", 3},
				{"a='alpha' AND b=1", 2}, {"a='alpha' OR a IS NULL", 3},
				{"MATCH(a) AGAINST('alpha')", 3}, {"MATCH(c) AGAINST('echo')", 3},
				{"MATCH(a) AGAINST('alpha') AND a='alpha'", 2}, {"a='alpha' AND MATCH(a) AGAINST('alpha')", 2},
				{"MATCH(a) AGAINST('alpha') AND a IS NULL", 0}, {"MATCH(a) AGAINST('alpha') AND b=5", 1},
				{"MATCH(a) AGAINST('alpha') AND MATCH(c) AGAINST('omega')", 1},
				{"a='alpha' OR MATCH(a) AGAINST('bravo')", 4}, {"MATCH(a) AGAINST('alpha') OR a IS NULL", 4},
				{"MATCH(c) AGAINST('echo') AND a IS NULL", 1},
			} {
				script.Assertions = append(script.Assertions, ScriptTestAssertion{Query: "SELECT COUNT(*) FROM t WHERE " + test.predicate, Expected: []sql.Row{{test.count}}})
			}
			scripts = append(scripts, script)
		}
	}
	return scripts
}()
