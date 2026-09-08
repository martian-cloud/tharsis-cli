package run

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAnnotations(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		wantLen int
		wantErr bool
	}{
		{
			name:    "empty input returns nil",
			input:   nil,
			wantLen: 0,
			wantErr: false,
		},
		{
			name:    "single valid annotation",
			input:   []string{`{"key":"commit","value":"abc123","link":"https://example.com/commit/abc123"}`},
			wantLen: 1,
			wantErr: false,
		},
		{
			name: "multiple valid annotations",
			input: []string{
				`{"key":"commit","value":"abc123"}`,
				`{"key":"ref","value":"main"}`,
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name:    "invalid JSON returns error",
			input:   []string{"not-json"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAnnotations(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
		})
	}
}

// TestParseAnnotations_PreservesFields confirms key, value, and the optional link survive parsing.
func TestParseAnnotations_PreservesFields(t *testing.T) {
	link := "https://example.com/commit/abc123"
	got, err := parseAnnotations([]string{
		`{"key":"commit","value":"abc123","link":"https://example.com/commit/abc123"}`,
		`{"key":"ref","value":"main"}`,
	})
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "commit", got[0].Key)
	assert.Equal(t, "abc123", got[0].Value)
	require.NotNil(t, got[0].Link)
	assert.Equal(t, link, *got[0].Link)

	assert.Equal(t, "ref", got[1].Key)
	assert.Equal(t, "main", got[1].Value)
	assert.Nil(t, got[1].Link)
}
