// Copyright 2020-2021 Dolthub, Inc.
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

package expression

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestModIntegerAndFloatingContract(t *testing.T) {
	ctx := sql.NewEmptyContext()
	for _, tt := range []struct {
		name             string
		l, r             any
		lt, rt, wantType sql.Type
		want             any
	}{
		{"signed", int64(-8), int64(3), types.Int64, types.Int64, types.Int64, int64(-2)},
		{"minimum", int64(math.MinInt64), int64(-1), types.Int64, types.Int64, types.Int64, int64(0)},
		{"unsigned", uint64(math.MaxUint64), int64(-7), types.Uint64, types.Int64, types.Uint64, uint64(1)},
		{"mixed sign", int64(-8), uint64(math.MaxUint64), types.Int64, types.Uint64, types.Int64, int64(-8)},
		{"zero divisor", int64(8), int64(0), types.Int64, types.Int64, types.Int64, nil},
		{"null", nil, int64(7), types.Int64, types.Int64, types.Int64, nil},
		{"floating", float64(1e30), int64(7), types.Float64, types.Int64, types.Float64, math.Mod(1e30, 7)},
	} {
		t.Run(tt.name, func(t *testing.T) {
			e := NewMod(NewLiteral(tt.l, tt.lt), NewLiteral(tt.r, tt.rt))
			require.True(t, e.Type(ctx).Equals(tt.wantType), "type %s", e.Type(ctx))
			got, err := e.Eval(ctx, nil)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}
