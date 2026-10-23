// Copyright 2026 BlueboxC
// Licensed under the Apache License, Version 2.0.
package function

import (
	"context"
	"errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/stretchr/testify/require"
	"testing"
)

type comparisonBaseType interface{ sql.Type }
type comparisonErrorType struct {
	comparisonBaseType
	err error
}

func (t comparisonErrorType) Compare(context.Context, any, any) (int, error) { return 0, t.err }
func TestGreatestLeastPropagatesComparisonErrors(t *testing.T) {
	ctx := sql.NewEmptyContext()
	expected := errors.New("comparison refused")
	args := []sql.Expression{expression.NewLiteral(int64(1), types.Int64), expression.NewLiteral(int64(2), types.Int64)}
	for _, direction := range []cmpDir{cmpGreater, cmpLess} {
		_, err := cmpEval(ctx, args, nil, comparisonErrorType{comparisonBaseType: types.Int64, err: expected}, types.Int64, direction)
		require.ErrorIs(t, err, expected)
	}
}
