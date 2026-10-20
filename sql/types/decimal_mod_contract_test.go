// Copyright 2022 Dolthub, Inc.
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

package types

import (
	"github.com/cockroachdb/apd/v3"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestDecimalModQuotientPrecision(t *testing.T) {
	for _, tt := range []struct{ a, b, want string }{
		{"100000000", "0.0001", "0"},
		{"1e30", "7", "1"},
		{"476098643", "0.5962", "0.5652"},
		{"-100000000", "0.0001", "0"},
		{"-8.125", "0.5", "-0.125"},
	} {
		t.Run(tt.a+"%"+tt.b, func(t *testing.T) {
			a, _, err := apd.NewFromString(tt.a)
			require.NoError(t, err)
			b, _, err := apd.NewFromString(tt.b)
			require.NoError(t, err)
			want, _, err := apd.NewFromString(tt.want)
			require.NoError(t, err)
			got, err := DecimalMod(a, b)
			require.NoError(t, err)
			require.Zero(t, got.Cmp(want))
			if got.IsZero() {
				require.False(t, got.Negative)
			}
		})
	}
}
