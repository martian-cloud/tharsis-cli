package tools

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/mcp/tharsis"
)

func TestGetConnectionInfo(t *testing.T) {
	tharsisURL := "https://test.tharsis.io"
	profileName := "production"

	tests := []struct {
		name string
	}{
		{
			name: "returns connection info when not authenticated",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &ToolContext{
				tharsisURL:  tharsisURL,
				profileName: profileName,
				clientGetter: func() (tharsis.Client, error) {
					return nil, errors.New("not authenticated")
				},
			}

			_, handler := getConnectionInfo(tc)
			_, output, err := handler(t.Context(), nil, getConnectionInfoInput{})

			assert.NoError(t, err)
			assert.Equal(t, tharsisURL, output.TharsisURL)
			assert.Equal(t, profileName, output.ProfileName)
			assert.False(t, output.Authenticated)
		})
	}
}
