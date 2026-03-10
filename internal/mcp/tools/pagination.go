package tools

import (
	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
)

const (
	defaultPageLimit int32 = 10
	maxPageLimit     int32 = 50
)

// toSortEnum converts a string sort value to a proto enum pointer.
func toSortEnum[T ~int32](sortStr *string, enumMap map[string]int32) *T {
	if sortStr == nil {
		return nil
	}

	if val, ok := enumMap[*sortStr]; ok {
		enumVal := T(val)
		return &enumVal
	}

	return nil
}

// pageInfo represents pagination information returned in responses.
// Note: Only forward pagination is supported (using cursor for next page).
type pageInfo struct {
	HasNextPage bool    `json:"has_next_page" jsonschema:"Whether there are more results available"`
	TotalCount  int32   `json:"total_count" jsonschema:"Total number of items available"`
	Cursor      *string `json:"cursor,omitempty" jsonschema:"Cursor for fetching the next page. Use this value in the cursor parameter of the next request."`
}

// buildPageInfo converts proto PageInfo to MCP pageInfo format.
func buildPageInfo(pbPageInfo *pb.PageInfo) pageInfo {
	return pageInfo{
		HasNextPage: pbPageInfo.HasNextPage,
		TotalCount:  pbPageInfo.TotalCount,
		Cursor:      pbPageInfo.EndCursor,
	}
}

// buildPaginationOptions creates proto PaginationOptions from limit and cursor pointers.
// Only forward pagination is supported (first/after).
func buildPaginationOptions(limit *int32, cursor *string) *pb.PaginationOptions {
	finalLimit := defaultPageLimit
	if limit != nil && *limit > 0 && *limit <= maxPageLimit {
		finalLimit = *limit
	} else if limit != nil && *limit > maxPageLimit {
		finalLimit = maxPageLimit
	}

	return &pb.PaginationOptions{
		First: &finalLimit,
		After: cursor,
	}
}
