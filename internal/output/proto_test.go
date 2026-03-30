package output

import (
	"strings"
	"testing"

	"github.com/aws/smithy-go/ptr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/descriptorpb"
)

func TestProtoToNamedValues(t *testing.T) {
	tests := []struct {
		name     string
		msg      proto.Message
		expected []string // expected Name fields in order
	}{
		{
			name:     "empty message",
			msg:      &descriptorpb.FileDescriptorProto{},
			expected: nil,
		},
		{
			name: "scalar fields sorted by value length",
			msg: &descriptorpb.FileDescriptorProto{
				Name:   ptr.String("test.proto"),
				Syntax: ptr.String("proto3"),
			},
			expected: []string{"Syntax", "Name"},
		},
		{
			name: "nil fields omitted",
			msg: &descriptorpb.FileDescriptorProto{
				Name:    ptr.String("test.proto"),
				Package: nil,
			},
			expected: []string{"Name"},
		},
		{
			name: "empty string fields omitted",
			msg: &descriptorpb.FileDescriptorProto{
				Name:   ptr.String("test.proto"),
				Syntax: ptr.String(""),
			},
			expected: []string{"Name"},
		},
		{
			name: "nested empty message omitted",
			msg: &descriptorpb.FileDescriptorProto{
				Name:           ptr.String("test.proto"),
				SourceCodeInfo: &descriptorpb.SourceCodeInfo{},
			},
			expected: []string{"Name"},
		},
		{
			name: "multiple fields sorted shortest to longest",
			msg: &descriptorpb.FileDescriptorProto{
				Name:    ptr.String("a_very_long_file_name.proto"),
				Package: ptr.String("pkg"),
				Syntax:  ptr.String("proto3"),
			},
			expected: []string{"Package", "Syntax", "Name"},
		},
		{
			name: "equal length values preserve proto order",
			msg: &descriptorpb.FileDescriptorProto{
				Name:    ptr.String("aaa"),
				Package: ptr.String("bbb"),
				Syntax:  ptr.String("ccc"),
			},
			expected: []string{"Name", "Package", "Syntax"},
		},
		{
			name: "boolean true is included",
			msg: &descriptorpb.FileOptions{
				JavaMultipleFiles: ptr.Bool(true),
			},
			expected: []string{"Java Multiple Files"},
		},
		{
			name: "boolean false is included with EmitDefaultValues",
			msg: &descriptorpb.FileOptions{
				JavaMultipleFiles: ptr.Bool(false),
			},
			expected: []string{"Java Multiple Files"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			values := ProtoToNamedValues(tc.msg)
			var names []string
			for _, v := range values {
				names = append(names, v.Name)
			}

			assert.Equal(t, tc.expected, names)
		})
	}

	t.Run("long values are truncated", func(t *testing.T) {
		long := strings.Repeat("x", 200)
		values := ProtoToNamedValues(&descriptorpb.FileDescriptorProto{Name: &long})

		require.Len(t, values, 1)
		assert.LessOrEqual(t, len(values[0].Value.(string)), maxValueLen+3)
		assert.True(t, strings.HasSuffix(values[0].Value.(string), "..."))
	})

	t.Run("value at exact max length is not truncated", func(t *testing.T) {
		exact := strings.Repeat("x", maxValueLen)
		values := ProtoToNamedValues(&descriptorpb.FileDescriptorProto{Name: &exact})

		require.Len(t, values, 1)
		assert.Equal(t, exact, values[0].Value)
	})

	t.Run("value one char over max is truncated", func(t *testing.T) {
		over := strings.Repeat("x", maxValueLen+1)
		values := ProtoToNamedValues(&descriptorpb.FileDescriptorProto{Name: &over})

		require.Len(t, values, 1)
		assert.Equal(t, maxValueLen+3, len(values[0].Value.(string)))
		assert.True(t, strings.HasSuffix(values[0].Value.(string), "..."))
	})

	t.Run("repeated fields are included as comma-separated", func(t *testing.T) {
		values := ProtoToNamedValues(&descriptorpb.FileDescriptorProto{
			Dependency: []string{"dep1.proto", "dep2.proto"},
		})

		require.Len(t, values, 1)
		assert.Equal(t, "dep1.proto, dep2.proto", values[0].Value)
	})

	t.Run("empty repeated fields are omitted", func(t *testing.T) {
		values := ProtoToNamedValues(&descriptorpb.FileDescriptorProto{
			Name:       ptr.String("test.proto"),
			Dependency: []string{},
		})

		// Only Name should appear; empty dependency list is omitted.
		names := make([]string, len(values))
		for i, v := range values {
			names[i] = v.Name
		}

		assert.NotContains(t, names, "Dependency")
	})
}

func TestProtoToTable(t *testing.T) {
	t.Run("nil for nil slice", func(t *testing.T) {
		assert.Nil(t, ProtoToTable(nil))
	})

	t.Run("nil for empty slice", func(t *testing.T) {
		assert.Nil(t, ProtoToTable([]proto.Message{}))
	})

	t.Run("headers and rows from proto messages", func(t *testing.T) {
		msgs := []proto.Message{
			&descriptorpb.FileDescriptorProto{Name: ptr.String("a.proto"), Syntax: ptr.String("proto3")},
			&descriptorpb.FileDescriptorProto{Name: ptr.String("b.proto"), Syntax: ptr.String("proto2")},
		}

		tbl := ProtoToTable(msgs)
		require.NotNil(t, tbl)
		assert.True(t, len(tbl.Headers) > 0)
		assert.Len(t, tbl.Rows, 2)
		assert.Equal(t, "a.proto", tbl.Rows[0][0].Value)
		assert.Equal(t, "b.proto", tbl.Rows[1][0].Value)
	})

	t.Run("empty fields show as empty strings in rows", func(t *testing.T) {
		msgs := []proto.Message{
			&descriptorpb.FileDescriptorProto{Name: ptr.String("a.proto")},
		}

		tbl := ProtoToTable(msgs)
		require.NotNil(t, tbl)
		// All non-name columns should be empty strings, not missing.
		assert.Equal(t, len(tbl.Headers), len(tbl.Rows[0]))
	})

	t.Run("single message produces one row", func(t *testing.T) {
		msgs := []proto.Message{
			&descriptorpb.FileDescriptorProto{Name: ptr.String("only.proto")},
		}

		tbl := ProtoToTable(msgs)
		require.NotNil(t, tbl)
		assert.Len(t, tbl.Rows, 1)
	})

	t.Run("all-empty columns are removed", func(t *testing.T) {
		// Both messages only set Name, so Syntax/Package/etc columns should be stripped.
		msgs := []proto.Message{
			&descriptorpb.FileDescriptorProto{Name: ptr.String("a.proto")},
			&descriptorpb.FileDescriptorProto{Name: ptr.String("b.proto")},
		}

		tbl := ProtoToTable(msgs)
		require.NotNil(t, tbl)

		// Only Name and Edition should remain (Edition has a default enum value).
		for _, h := range tbl.Headers {
			assert.NotEqual(t, "Syntax", h)
			assert.NotEqual(t, "Package", h)
		}
	})

	t.Run("column kept if any row has value", func(t *testing.T) {
		// First message has Syntax, second doesn't — column should still appear.
		msgs := []proto.Message{
			&descriptorpb.FileDescriptorProto{Name: ptr.String("a.proto"), Syntax: ptr.String("proto3")},
			&descriptorpb.FileDescriptorProto{Name: ptr.String("b.proto")},
		}

		tbl := ProtoToTable(msgs)
		require.NotNil(t, tbl)

		hasSyntax := false
		for _, h := range tbl.Headers {
			if h == "Syntax" {
				hasSyntax = true
			}
		}

		assert.True(t, hasSyntax)
	})

	t.Run("uniform non-empty columns are kept", func(t *testing.T) {
		msgs := []proto.Message{
			&descriptorpb.FileDescriptorProto{Name: ptr.String("a.proto"), Syntax: ptr.String("proto3")},
			&descriptorpb.FileDescriptorProto{Name: ptr.String("b.proto"), Syntax: ptr.String("proto3")},
		}

		tbl := ProtoToTable(msgs)
		require.NotNil(t, tbl)

		hasSyntax := false
		for _, h := range tbl.Headers {
			if h == "Syntax" {
				hasSyntax = true
			}
		}

		assert.True(t, hasSyntax)
	})

	t.Run("uniform boolean columns are kept", func(t *testing.T) {
		msgs := []proto.Message{
			&descriptorpb.FileOptions{JavaMultipleFiles: ptr.Bool(false), JavaStringCheckUtf8: ptr.Bool(false)},
			&descriptorpb.FileOptions{JavaMultipleFiles: ptr.Bool(false), JavaStringCheckUtf8: ptr.Bool(false)},
		}

		tbl := ProtoToTable(msgs)
		require.NotNil(t, tbl)

		hasBoolean := false
		for _, h := range tbl.Headers {
			if h == "Java Multiple Files" || h == "Java String Check Utf8" {
				hasBoolean = true
			}
		}

		assert.True(t, hasBoolean)
	})
}

func TestToProtoSlice(t *testing.T) {
	t.Run("converts typed slice", func(t *testing.T) {
		input := []*descriptorpb.FileDescriptorProto{
			{Name: ptr.String("a.proto")},
			{Name: ptr.String("b.proto")},
		}

		assert.Len(t, toProtoSlice(input), 2)
	})

	t.Run("nil for non-slice", func(t *testing.T) {
		assert.Nil(t, toProtoSlice("not a slice"))
	})

	t.Run("nil for nil input", func(t *testing.T) {
		assert.Nil(t, toProtoSlice(nil))
	})

	t.Run("empty for empty slice", func(t *testing.T) {
		assert.Empty(t, toProtoSlice([]*descriptorpb.FileDescriptorProto{}))
	})

	t.Run("nil for int slice", func(t *testing.T) {
		// Slice of non-proto types should return empty.
		assert.Empty(t, toProtoSlice([]int{1, 2, 3}))
	})

	t.Run("single element slice", func(t *testing.T) {
		input := []*descriptorpb.FileDescriptorProto{
			{Name: ptr.String("only.proto")},
		}

		result := toProtoSlice(input)
		require.Len(t, result, 1)
	})

	t.Run("pointer to slice returns nil", func(t *testing.T) {
		s := []*descriptorpb.FileDescriptorProto{}
		assert.Nil(t, toProtoSlice(&s))
	})
}

func TestExtractListFields(t *testing.T) {
	t.Run("plain slice returns items with no page info", func(t *testing.T) {
		input := []*descriptorpb.FileDescriptorProto{
			{Name: ptr.String("a.proto")},
		}

		items, pageInfo := ExtractListFields(input)
		assert.Len(t, items, 1)
		assert.Nil(t, pageInfo)
	})

	t.Run("nil for non-proto non-slice", func(t *testing.T) {
		items, pageInfo := ExtractListFields("not a proto")
		assert.Nil(t, items)
		assert.Nil(t, pageInfo)
	})

	t.Run("nil for nil input", func(t *testing.T) {
		items, pageInfo := ExtractListFields(nil)
		assert.Nil(t, items)
		assert.Nil(t, pageInfo)
	})

	t.Run("proto with repeated field extracts items", func(t *testing.T) {
		msg := &descriptorpb.FileDescriptorProto{
			MessageType: []*descriptorpb.DescriptorProto{
				{Name: ptr.String("Foo")},
				{Name: ptr.String("Bar")},
			},
		}

		items, pageInfo := ExtractListFields(msg)
		assert.Len(t, items, 2)
		assert.Nil(t, pageInfo)
	})

	t.Run("proto with empty repeated field returns nil items", func(t *testing.T) {
		msg := &descriptorpb.FileDescriptorProto{}

		items, pageInfo := ExtractListFields(msg)
		assert.Nil(t, items)
		assert.Nil(t, pageInfo)
	})

	t.Run("proto with only scalar fields returns nil items", func(t *testing.T) {
		msg := &descriptorpb.FileDescriptorProto{
			Name:   ptr.String("test.proto"),
			Syntax: ptr.String("proto3"),
		}

		items, pageInfo := ExtractListFields(msg)
		assert.Nil(t, items)
		assert.Nil(t, pageInfo)
	})

	t.Run("int input returns nil", func(t *testing.T) {
		items, pageInfo := ExtractListFields(42)
		assert.Nil(t, items)
		assert.Nil(t, pageInfo)
	})
}

func TestFormatValue(t *testing.T) {
	assert.Equal(t, "", formatValue(nil))
	assert.Equal(t, "hello", formatValue("hello"))
	assert.Equal(t, "42", formatValue(float64(42)))
	assert.Equal(t, "true", formatValue(true))
	assert.Equal(t, "a=1, b=2", formatValue(map[string]any{"b": 2, "a": 1}))
	assert.Equal(t, "x, y, z", formatValue([]any{"x", "y", "z"}))
	assert.Equal(t, "", formatValue([]any{}))

	// Edge cases.
	assert.Equal(t, "", formatValue(""))
	assert.Equal(t, "k=v", formatValue(map[string]any{"k": "v"}))
	assert.Equal(t, "only", formatValue([]any{"only"}))
	assert.Equal(t, "k=[nested]", formatValue(map[string]any{"k": []any{"nested"}}))
	assert.Equal(t, "0", formatValue(0))
	assert.Equal(t, "false", formatValue(false))
	assert.Equal(t, "3.14", formatValue(3.14))
	assert.Equal(t, "", formatValue(map[string]any{}))
}

func TestFlattenProtoMap(t *testing.T) {
	t.Run("nested message fields are prefixed", func(t *testing.T) {
		msg := &descriptorpb.DescriptorProto{
			Name: ptr.String("MyMessage"),
			Options: &descriptorpb.MessageOptions{
				Deprecated: ptr.Bool(true),
			},
		}

		result := flattenProtoMap(msg)
		assert.Equal(t, "MyMessage", result["name"])
		assert.Equal(t, true, result["options_deprecated"])
	})

	t.Run("enum fields are included as strings", func(t *testing.T) {
		optLevel := descriptorpb.FileOptions_SPEED
		msg := &descriptorpb.FileDescriptorProto{
			Name:    ptr.String("test.proto"),
			Options: &descriptorpb.FileOptions{OptimizeFor: &optLevel},
		}

		result := flattenProtoMap(msg)
		assert.Contains(t, result, "options_optimize_for")
	})

	t.Run("empty message returns map with defaults", func(t *testing.T) {
		result := flattenProtoMap(&descriptorpb.FileDescriptorProto{})
		assert.NotNil(t, result)
	})
}

func TestDisplayName(t *testing.T) {
	assert.Equal(t, "Full Path", displayName("full_path"))
	assert.Equal(t, "Created At", displayName("created_at"))
	assert.Equal(t, "Name", displayName("name"))
	assert.Equal(t, "Current State Version Id", displayName("current_state_version_id"))
	assert.Equal(t, "", displayName(""))

	// Edge cases.
	assert.Equal(t, "A  B", displayName("a__b"))
	assert.Equal(t, " Leading", displayName("_leading"))
	assert.Equal(t, "Trailing ", displayName("trailing_"))
	assert.Equal(t, "Allcaps", displayName("ALLCAPS"))
}
