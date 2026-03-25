package tools

import (
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

func TestToSortEnum(t *testing.T) {
	enumMap := map[string]int32{
		"name_asc":  1,
		"name_desc": 2,
		"date_asc":  3,
	}

	type testCase struct {
		name     string
		sortStr  *string
		expected *int32
	}

	testCases := []testCase{
		{
			name:     "nil input",
			sortStr:  nil,
			expected: nil,
		},
		{
			name:     "valid enum value",
			sortStr:  ptr.String("name_asc"),
			expected: ptr.Int32(1),
		},
		{
			name:     "invalid enum value",
			sortStr:  ptr.String("invalid"),
			expected: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := toSortEnum[int32](tc.sortStr, enumMap)
			if tc.expected == nil {
				assert.Nil(t, result)
			} else {
				assert.Equal(t, *tc.expected, *result)
			}
		})
	}
}

func TestBuildPageInfo(t *testing.T) {
	type testCase struct {
		name     string
		input    *pb.PageInfo
		expected pageInfo
	}

	testCases := []testCase{
		{
			name: "with cursor",
			input: &pb.PageInfo{
				HasNextPage: true,
				TotalCount:  100,
				EndCursor:   ptr.String("cursor123"),
			},
			expected: pageInfo{
				HasNextPage: true,
				TotalCount:  100,
				Cursor:      ptr.String("cursor123"),
			},
		},
		{
			name: "without cursor",
			input: &pb.PageInfo{
				HasNextPage: false,
				TotalCount:  10,
			},
			expected: pageInfo{
				HasNextPage: false,
				TotalCount:  10,
				Cursor:      nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := buildPageInfo(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestBuildPaginationOptions(t *testing.T) {
	type testCase struct {
		name     string
		limit    *int32
		cursor   *string
		expected *pb.PaginationOptions
	}

	testCases := []testCase{
		{
			name:   "nil limit uses default",
			limit:  nil,
			cursor: nil,
			expected: &pb.PaginationOptions{
				First: ptr.Int32(10),
				After: nil,
			},
		},
		{
			name:   "valid limit",
			limit:  ptr.Int32(25),
			cursor: ptr.String("cursor123"),
			expected: &pb.PaginationOptions{
				First: ptr.Int32(25),
				After: ptr.String("cursor123"),
			},
		},
		{
			name:   "limit exceeds max",
			limit:  ptr.Int32(100),
			cursor: nil,
			expected: &pb.PaginationOptions{
				First: ptr.Int32(50),
				After: nil,
			},
		},
		{
			name:   "zero limit uses default",
			limit:  ptr.Int32(0),
			cursor: nil,
			expected: &pb.PaginationOptions{
				First: ptr.Int32(10),
				After: nil,
			},
		},
		{
			name:   "negative limit uses default",
			limit:  ptr.Int32(-5),
			cursor: nil,
			expected: &pb.PaginationOptions{
				First: ptr.Int32(10),
				After: nil,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := buildPaginationOptions(tc.limit, tc.cursor)
			assert.Equal(t, tc.expected, result)
		})
	}
}
