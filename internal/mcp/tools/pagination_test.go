package tools

import (
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

func TestBuildPaginationOptions(t *testing.T) {
	type testCase struct {
		name           string
		limit          *int32
		cursor         *string
		expectedLimit  int32
		expectedCursor *string
	}

	tests := []testCase{
		{
			name:           "with both limit and cursor",
			limit:          ptr.Int32(50),
			cursor:         ptr.String("abc123"),
			expectedLimit:  50,
			expectedCursor: ptr.String("abc123"),
		},
		{
			name:           "with only limit",
			limit:          ptr.Int32(25),
			cursor:         nil,
			expectedLimit:  25,
			expectedCursor: nil,
		},
		{
			name:           "with only cursor",
			limit:          nil,
			cursor:         ptr.String("xyz789"),
			expectedLimit:  defaultPageLimit,
			expectedCursor: ptr.String("xyz789"),
		},
		{
			name:           "with no parameters - uses defaults",
			limit:          nil,
			cursor:         nil,
			expectedLimit:  defaultPageLimit,
			expectedCursor: nil,
		},
		{
			name:           "with limit exceeding max - capped at maxPageLimit",
			limit:          ptr.Int32(200),
			cursor:         nil,
			expectedLimit:  maxPageLimit,
			expectedCursor: nil,
		},
		{
			name:           "with limit below 1 - uses default",
			limit:          ptr.Int32(0),
			cursor:         nil,
			expectedLimit:  defaultPageLimit,
			expectedCursor: nil,
		},
		{
			name:           "with negative limit - uses default",
			limit:          ptr.Int32(-5),
			cursor:         nil,
			expectedLimit:  defaultPageLimit,
			expectedCursor: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPaginationOptions(tt.limit, tt.cursor)

			assert.NotNil(t, result.Limit)
			assert.Equal(t, tt.expectedLimit, *result.Limit)

			if tt.expectedCursor != nil {
				assert.NotNil(t, result.Cursor)
				assert.Equal(t, *tt.expectedCursor, *result.Cursor)
			} else {
				assert.Nil(t, result.Cursor)
			}
		})
	}
}

func TestBuildPageInfo(t *testing.T) {
	type testCase struct {
		name     string
		input    sdktypes.PageInfo
		expected pageInfo
	}

	tests := []testCase{
		{
			name: "with next page",
			input: sdktypes.PageInfo{
				HasNextPage: true,
				Cursor:      "next-cursor-123",
			},
			expected: pageInfo{
				HasNextPage: true,
				Cursor:      "next-cursor-123",
			},
		},
		{
			name: "without next page",
			input: sdktypes.PageInfo{
				HasNextPage: false,
				Cursor:      "",
			},
			expected: pageInfo{
				HasNextPage: false,
				Cursor:      "",
			},
		},
		{
			name: "last page with cursor",
			input: sdktypes.PageInfo{
				HasNextPage: false,
				Cursor:      "last-cursor",
			},
			expected: pageInfo{
				HasNextPage: false,
				Cursor:      "last-cursor",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPageInfo(&tt.input)
			assert.Equal(t, tt.expected.HasNextPage, result.HasNextPage)
			assert.Equal(t, tt.expected.Cursor, result.Cursor)
		})
	}
}
