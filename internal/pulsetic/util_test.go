package pulsetic

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIntBool_UnmarshalJSON(t *testing.T) {
	type args struct {
		bytes []byte
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr require.ErrorAssertionFunc
	}{
		{"false int", args{[]byte("0")}, false, require.NoError},
		{"true int", args{[]byte("1")}, true, require.NoError},
		{"false", args{[]byte("false")}, false, require.NoError},
		{"true", args{[]byte("true")}, true, require.NoError},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var b IntBool
			tt.wantErr(t, b.UnmarshalJSON(tt.args.bytes))
			assert.Equal(t, tt.want, bool(b))
		})
	}
}
