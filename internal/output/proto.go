package output

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"

	pb "gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-api/pkg/protos/gen"
	"gitlab.com/infor-cloud/martian-cloud/tharsis/tharsis-cli/internal/terminal"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const (
	// humanTimeFormat is the format used for displaying timestamps.
	humanTimeFormat = "January 2 2006, 3:04:05 PM MST"

	// maxValueLen is the maximum display length for a field value before truncation.
	maxValueLen = 120
)

var (
	titleCaser = cases.Title(language.English)

	// tableSkippedFields are proto fields hidden from table output to reduce clutter.
	tableSkippedFields = map[string]struct{}{
		"created_at": {},
		"created_by": {},
		"updated_at": {},
		"version":    {},
		"trn":        {},
	}

	jsonOpts = protojson.MarshalOptions{
		EmitDefaultValues: true,
		UseProtoNames:     true,
	}
)

// field is a flattened proto field with a display name and raw value.
type field struct {
	name  string
	key   string // original snake_case name
	value any
}

// ProtoToNamedValues converts a proto message into NamedValues for display,
// sorted by value length (shortest first). Empty values are omitted.
func ProtoToNamedValues(msg proto.Message) []terminal.NamedValue {
	var values []terminal.NamedValue
	for _, f := range flattenProto(msg) {
		s := formatValue(f.value)
		if s == "" {
			continue
		}

		if len(s) > maxValueLen {
			s = s[:maxValueLen] + "..."
		}

		values = append(values, terminal.NamedValue{Name: f.name, Value: s})
	}

	sort.SliceStable(values, func(i, j int) bool {
		return len(values[i].Value.(string)) < len(values[j].Value.(string))
	})

	return values
}

// ProtoToTable builds a Table from a slice of proto messages.
// Column headers come from field descriptors; tableSkippedFields are excluded.
func ProtoToTable(msgs []proto.Message) *terminal.Table {
	if len(msgs) == 0 {
		return nil
	}

	// Derive headers from descriptors, filtering skipped fields.
	var headers []string
	var headerKeys []string
	for _, f := range protoFieldNames(msgs[0].ProtoReflect()) {
		if _, skip := tableSkippedFields[f.key]; skip {
			continue
		}

		headers = append(headers, f.name)
		headerKeys = append(headerKeys, f.key)
	}

	// Build rows, keeping only non-skipped fields.
	var rows [][]string
	for _, msg := range msgs {
		values := flattenProtoMap(msg)
		row := make([]string, len(headerKeys))
		for i, key := range headerKeys {
			row[i] = formatValue(values[key])
		}

		rows = append(rows, row)
	}

	// Only truncate if content exceeds terminal width.
	colWidth := 0
	if n := len(headers); n > 0 {
		totalWidth := n * 3
		for i, h := range headers {
			w := len(h)
			for _, row := range rows {
				if len(row[i]) > w {
					w = len(row[i])
				}
			}

			totalWidth += w
		}

		if termWidth := terminal.GetTerminalWidth(); totalWidth > termWidth {
			colWidth = (termWidth - n*3) / n
		}
	}

	t := terminal.NewTable(headers...)
	for _, row := range rows {
		if colWidth > 0 {
			for i := range row {
				if len(row[i]) > colWidth {
					row[i] = row[i][:colWidth] + "..."
				}
			}
		}

		t.Rich(row, nil)
	}

	return t
}

// ExtractListFields extracts the repeated message items and optional PageInfo
// from a proto response. Plain slices are returned as-is with no PageInfo.
func ExtractListFields(data any) ([]proto.Message, *pb.PageInfo) {
	if _, ok := data.(proto.Message); !ok {
		return toProtoSlice(data), nil
	}

	msg := data.(proto.Message)
	var items []proto.Message
	var pageInfo *pb.PageInfo

	fields := msg.ProtoReflect().Descriptor().Fields()
	for i := range fields.Len() {
		fd := fields.Get(i)
		val := msg.ProtoReflect().Get(fd)

		if fd.IsList() && fd.Kind() == protoreflect.MessageKind {
			list := val.List()
			for j := range list.Len() {
				items = append(items, list.Get(j).Message().Interface())
			}

			continue
		}

		if fd.Kind() == protoreflect.MessageKind && fd.Message().Name() == "PageInfo" {
			if pi, ok := val.Message().Interface().(*pb.PageInfo); ok {
				pageInfo = pi
			}
		}
	}

	return items, pageInfo
}

// toProtoSlice converts a typed slice (e.g. []*pb.Group) to []proto.Message.
func toProtoSlice(v any) []proto.Message {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Slice {
		return nil
	}

	msgs := make([]proto.Message, 0, rv.Len())
	for i := range rv.Len() {
		if msg, ok := rv.Index(i).Interface().(proto.Message); ok {
			msgs = append(msgs, msg)
		}
	}

	return msgs
}

// flattenProto returns fields in proto definition order with formatted values.
func flattenProto(msg proto.Message) []field {
	values := flattenProtoMap(msg)
	names := protoFieldNames(msg.ProtoReflect())

	fields := make([]field, 0, len(names))
	for _, n := range names {
		fields = append(fields, field{name: n.name, key: n.key, value: values[n.key]})
	}

	return fields
}

// flattenProtoMap marshals a proto to JSON and returns a flat map of snake_case key → value.
// Nested objects are flattened; timestamps are formatted as human-readable strings.
func flattenProtoMap(msg proto.Message) map[string]any {
	data, err := jsonOpts.Marshal(msg)
	if err != nil {
		return nil
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}

	// Collect map field names so we don't flatten them.
	mapFields := make(map[string]struct{})
	collectMapFields(msg.ProtoReflect().Descriptor().Fields(), mapFields)

	out := make(map[string]any)
	var flatten func(m map[string]any)
	flatten = func(m map[string]any) {
		for k, v := range m {
			nested, isMap := v.(map[string]any)
			if !isMap {
				// Format RFC 3339 timestamps to human-readable.
				if s, ok := v.(string); ok {
					if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
						v = t.Local().Format(humanTimeFormat)
					}
				}

				out[k] = v

				continue
			}

			// Keep proto map fields (e.g. labels) as-is; flatten nested messages.
			if _, isProtoMap := mapFields[k]; isProtoMap {
				out[k] = nested
			} else {
				flatten(nested)
			}
		}
	}

	flatten(raw)

	return out
}

// collectMapFields finds all map field names in a proto descriptor, recursing into nested messages.
func collectMapFields(fields protoreflect.FieldDescriptors, out map[string]struct{}) {
	for i := range fields.Len() {
		fd := fields.Get(i)
		if fd.IsMap() {
			out[string(fd.Name())] = struct{}{}
		} else if fd.Kind() == protoreflect.MessageKind && !fd.IsList() {
			collectMapFields(fd.Message().Fields(), out)
		}
	}
}

// protoFieldNames returns display names in proto definition order,
// flattening nested messages (e.g. metadata).
func protoFieldNames(msg protoreflect.Message) []field {
	var out []field

	var walk func(protoreflect.FieldDescriptors, protoreflect.Message)
	walk = func(fields protoreflect.FieldDescriptors, m protoreflect.Message) {
		for i := range fields.Len() {
			fd := fields.Get(i)
			key := string(fd.Name())

			if fd.Kind() == protoreflect.MessageKind && !fd.IsList() && !fd.IsMap() {
				if fd.Message().FullName() == "google.protobuf.Timestamp" {
					out = append(out, field{name: displayName(key), key: key})
					continue
				}

				walk(fd.Message().Fields(), m.Get(fd).Message())

				continue
			}

			out = append(out, field{name: displayName(key), key: key})
		}
	}

	walk(msg.Descriptor().Fields(), msg)

	return out
}

// formatValue converts a field value to a display string.
func formatValue(v any) string {
	switch val := v.(type) {
	case nil:
		return ""
	case string:
		return val
	case map[string]any:
		pairs := make([]string, 0, len(val))
		for k, v := range val {
			pairs = append(pairs, fmt.Sprintf("%s=%v", k, v))
		}

		sort.Strings(pairs)

		return strings.Join(pairs, ", ")
	case []any:
		items := make([]string, len(val))
		for i, v := range val {
			items[i] = fmt.Sprint(v)
		}

		return strings.Join(items, ", ")
	default:
		return fmt.Sprint(val)
	}
}

// displayName converts a snake_case proto field name to Title Case.
func displayName(s string) string {
	return titleCaser.String(strings.ReplaceAll(s, "_", " "))
}
