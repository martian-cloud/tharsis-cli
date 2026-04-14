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
	// humanTimeFormat is used instead of raw RFC3339 timestamps
	// because ISO dates are hard to scan quickly in CLI output.
	humanTimeFormat = "January 2 2006, 3:04:05 PM MST"

	// ellipsis is appended to truncated column values in table output.
	ellipsis = "..."
)

var (
	titleCaser = cases.Title(language.English)

	// transparentFields are wrapper messages (like metadata) whose children
	// should appear as top-level fields rather than nested under a prefix.
	transparentFields = map[string]struct{}{
		"metadata": {},
	}

	// jsonOpts emits zero-values so boolean false and empty strings
	// are distinguishable from absent fields in the output.
	jsonOpts = protojson.MarshalOptions{
		EmitDefaultValues: true,
		UseProtoNames:     true,
	}
)

// field is a flattened proto field with a display name and raw value.
type field struct {
	name  string
	key   string // original snake_case name used for lookups
	value any
}

// ProtoToNamedValues converts a proto message into NamedValues for display,
// sorted by value length (shortest first). Empty values are omitted.
func ProtoToNamedValues(msg proto.Message) []terminal.NamedValue {
	var values []terminal.NamedValue
	for _, f := range flattenProto(msg) {
		// Skip repeated message fields since they don't render well as named values.
		if items, ok := f.value.([]any); ok && len(items) > 0 {
			if _, isMap := items[0].(map[string]any); isMap {
				continue
			}
		}

		s := formatValue(f.value)
		if s == "" {
			continue
		}

		values = append(values, terminal.NamedValue{Name: f.name, Value: s})
	}

	// Shorter values first so the eye can scan key identifiers quickly.
	sort.SliceStable(values, func(i, j int) bool {
		return len(values[i].Value.(string)) < len(values[j].Value.(string))
	})

	return values
}

// ProtoToTable builds a Table from a slice of proto messages.
// fields restricts and orders columns by their snake_case keys
// (e.g. "full_path", "terraform_version"). When omitted, all fields are shown
// as a fallback but callers should always specify the fields they need.
func ProtoToTable(msgs []proto.Message, fields ...string) *terminal.Table {
	if len(msgs) == 0 {
		return nil
	}

	var headers []string
	var headerKeys []string
	for _, f := range protoFieldNames(msgs[0].ProtoReflect()) {
		headers = append(headers, f.name)
		headerKeys = append(headerKeys, f.key)
	}

	if len(fields) > 0 {
		// Commands specify which columns to display so tables only contain
		// relevant information and column order stays predictable regardless
		// of proto field ordering.
		index := make(map[string]int, len(headerKeys))
		for i, k := range headerKeys {
			index[k] = i
		}

		var fh, fk []string
		for _, f := range fields {
			if i, ok := index[f]; ok {
				fh = append(fh, headers[i])
				fk = append(fk, headerKeys[i])
			}
		}

		headers, headerKeys = fh, fk
	}

	rows := make([][]string, len(msgs))
	for i, msg := range msgs {
		values := flattenProtoMap(msg)
		row := make([]string, len(headerKeys))
		for j, key := range headerKeys {
			row[j] = formatValue(values[key])
		}

		rows[i] = row
	}

	// Drop columns where every row is empty since they add visual
	// noise without conveying information (e.g. unset optional fields
	// like team_id on memberships).
	headers, rows = dropEmptyColumns(headers, rows)

	colWidths := columnWidths(headers, rows)

	t := terminal.NewTable(headers...)
	for _, row := range rows {
		for i := range row {
			if colWidths != nil && len(row[i]) > colWidths[i] {
				row[i] = row[i][:colWidths[i]] + ellipsis
			}
		}

		t.Rich(row, nil)
	}

	return t
}

// ExtractListFields extracts the repeated message items and optional PageInfo
// from a proto response. Plain slices are returned as-is with no PageInfo.
func ExtractListFields(data any) ([]proto.Message, *pb.PageInfo) {
	msg, ok := data.(proto.Message)
	if !ok {
		return toProtoSlice(data), nil
	}

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

func dropEmptyColumns(headers []string, rows [][]string) ([]string, [][]string) {
	var keep []int
	for i := range headers {
		for _, row := range rows {
			if row[i] != "" {
				keep = append(keep, i)
				break
			}
		}
	}

	if len(keep) == len(headers) {
		return headers, rows
	}

	fh := make([]string, len(keep))
	for j, i := range keep {
		fh[j] = headers[i]
	}

	for r, row := range rows {
		fr := make([]string, len(keep))
		for j, i := range keep {
			fr[j] = row[i]
		}

		rows[r] = fr
	}

	return fh, rows
}

// columnWidths returns per-column character limits to fit the table within
// the terminal, or nil if no truncation is needed. Columns that naturally
// fit within an equal share keep their width; the surplus is redistributed
// to wider columns so narrow columns aren't needlessly truncated.
func columnWidths(headers []string, rows [][]string) []int {
	n := len(headers)
	if n == 0 {
		return nil
	}

	// Find the natural (max content) width of each column.
	natural := make([]int, n)
	for i, h := range headers {
		natural[i] = len(h)
		for _, row := range rows {
			if len(row[i]) > natural[i] {
				natural[i] = len(row[i])
			}
		}
	}

	separators := n * 3
	totalWidth := separators
	for _, w := range natural {
		totalWidth += w
	}

	termWidth := terminal.GetTerminalWidth()
	if totalWidth <= termWidth {
		return nil
	}

	// Start with an equal share per column, then let narrow columns
	// give back unused space to wider ones.
	available := termWidth - separators
	if available <= 0 {
		return nil
	}

	limits := make([]int, n)
	equalShare := available / n
	if equalShare == 0 {
		equalShare = 1
	}

	surplus := 0

	// Count how many columns need more than their equal share.
	wideCount := 0
	for i, w := range natural {
		if w <= equalShare {
			limits[i] = w
			surplus += equalShare - w
		} else {
			wideCount++
		}
	}

	// Distribute the surplus evenly among wide columns.
	wideShare := equalShare
	if wideCount > 0 {
		wideShare = equalShare + surplus/wideCount
	}

	for i, w := range natural {
		if w > equalShare {
			limits[i] = max(wideShare-len(ellipsis), 0)
		}
	}

	return limits
}

// toProtoSlice converts a typed slice (e.g. []*pb.Group) to []proto.Message
// using reflection, since Go generics can't express "any slice of proto.Message implementors".
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

// flattenProtoMap marshals a proto to JSON then flattens nested objects into
// a single-level map keyed by snake_case paths. This approach avoids manually
// walking every proto field type — protojson handles oneof, wrappers, enums, etc.
func flattenProtoMap(msg proto.Message) map[string]any {
	data, err := jsonOpts.Marshal(msg)
	if err != nil {
		return nil
	}

	var raw map[string]any
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil
	}

	// Proto map fields look like nested JSON objects but shouldn't be flattened —
	// they're user-defined key-value pairs, not structured sub-messages.
	mapFields := make(map[string]struct{})
	collectMapFields(msg.ProtoReflect().Descriptor().Fields(), mapFields)

	out := make(map[string]any)
	flattenJSON(raw, "", mapFields, out)

	return out
}

// flattenJSON recursively flattens a JSON object into out, joining nested keys
// with underscores. Proto map fields and scalar values are stored as-is.
func flattenJSON(m map[string]any, prefix string, mapFields map[string]struct{}, out map[string]any) {
	for k, v := range m {
		nested, isMap := v.(map[string]any)
		if !isMap {
			// Convert RFC3339 timestamps to a human-friendly format
			// since raw timestamps are unreadable in CLI output.
			if s, ok := v.(string); ok {
				if t, err := time.Parse(time.RFC3339Nano, s); err == nil {
					v = t.Local().Format(humanTimeFormat)
				}
			}

			out[prefix+k] = v

			continue
		}

		// Proto map fields are user-defined key-value pairs that should
		// be kept as a single value, not recursively flattened.
		if _, isProtoMap := mapFields[k]; isProtoMap {
			out[prefix+k] = nested
			continue
		}

		childPrefix := prefix + k + "_"
		if _, ok := transparentFields[k]; ok {
			childPrefix = prefix
		}

		flattenJSON(nested, childPrefix, mapFields, out)
	}
}

// collectMapFields finds all map field names in a proto descriptor tree.
// These need special handling during JSON flattening to avoid breaking
// user-defined key-value pairs into separate flattened keys.
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
// flattening transparent wrapper messages (e.g. metadata).
func protoFieldNames(msg protoreflect.Message) []field {
	var out []field
	walkFieldNames(msg.Descriptor().Fields(), msg, "", &out)

	return out
}

// walkFieldNames recursively collects field display names from proto descriptors.
func walkFieldNames(fields protoreflect.FieldDescriptors, m protoreflect.Message, prefix string, out *[]field) {
	for i := range fields.Len() {
		fd := fields.Get(i)
		name := string(fd.Name())
		key := prefix + name

		// Recurse into nested messages to flatten their fields,
		// but treat Timestamps as leaf values since they display as a single string.
		if fd.Kind() == protoreflect.MessageKind && !fd.IsList() && !fd.IsMap() {
			if fd.Message().FullName() == "google.protobuf.Timestamp" {
				*out = append(*out, field{name: displayName(key), key: key})
				continue
			}

			childPrefix := key + "_"
			if _, ok := transparentFields[name]; ok {
				childPrefix = prefix
			}

			walkFieldNames(fd.Message().Fields(), m.Get(fd).Message(), childPrefix, out)

			continue
		}

		*out = append(*out, field{name: displayName(key), key: key})
	}
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

		// Sorted for deterministic output in tests and consistent display.
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

// displayName converts a snake_case proto field name to Title Case for display.
func displayName(s string) string {
	return titleCaser.String(strings.ReplaceAll(s, "_", " "))
}
