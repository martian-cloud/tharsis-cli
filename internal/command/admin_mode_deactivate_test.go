package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdminModeDeactivateValidate(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "no arguments is valid",
			args:    nil,
			wantErr: false,
		},
		{
			name:    "unexpected arguments",
			args:    []string{"extra"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &adminModeDeactivateCommand{
				BaseCommand: &BaseCommand{arguments: tt.args},
			}
			err := cmd.validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
