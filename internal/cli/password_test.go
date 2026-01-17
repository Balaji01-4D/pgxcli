package cli

import (
	"fmt"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func Test_shouldAskForPassword(t *testing.T) {
	type args struct {
		err         error
		neverPrompt bool
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "not pgx error and never prompt",
			args: args{err: fmt.Errorf("someError"), neverPrompt: true},
			want: false,
		},
		{
			name: "not pgx error and prompt",
			args: args{err: fmt.Errorf("someError"), neverPrompt: false},
			want: false,
		},
		{
			name: "pgx invalid password error and never prompt",
			args: args{err: &pgconn.PgError{Code: "28P01"}, neverPrompt: true},
			want: false,
		},
		{
			name: "pgx invalid password error and prompt",
			args: args{err: &pgconn.PgError{Code: "28P01"}, neverPrompt: false},
			want: true,
		},  

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldAskForPassword(tt.args.err, tt.args.neverPrompt)
			assert.Equal(t, tt.want, got)
		})
	}
}
