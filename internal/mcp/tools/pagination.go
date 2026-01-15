package tools

import (
	sdktypes "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-sdk-go/pkg/types"
)

const (
	defaultPageLimit = 10
	maxPageLimit     = 50
)

// pageInfo represents pagination information returned in responses.
type pageInfo struct {
	HasNextPage bool   `json:"has_next_page" jsonschema:"Whether there are more results available"`
	TotalCount  int    `json:"total_count" jsonschema:"Total number of items available"`
	Cursor      string `json:"cursor,omitempty" jsonschema:"Cursor for fetching the next page. Use this value in the cursor parameter of the next request."`
}

// buildPageInfo converts SDK PageInfo to MCP pageInfo format.
func buildPageInfo(sdkPageInfo *sdktypes.PageInfo) pageInfo {
	return pageInfo{
		HasNextPage: sdkPageInfo.HasNextPage,
		TotalCount:  sdkPageInfo.TotalCount,
		Cursor:      sdkPageInfo.Cursor,
	}
}

// buildPaginationOptions creates SDK PaginationOptions from limit and cursor pointers.
func buildPaginationOptions(limit *int32, cursor *string) *sdktypes.PaginationOptions {
	finalLimit := int32(defaultPageLimit)
	if limit != nil && *limit > 0 && *limit <= maxPageLimit {
		finalLimit = *limit
	} else if limit != nil && *limit > maxPageLimit {
		finalLimit = maxPageLimit
	}

	return &sdktypes.PaginationOptions{
		Limit:  &finalLimit,
		Cursor: cursor,
	}
}
