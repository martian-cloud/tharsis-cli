// Package flag wraps the standard flag package with required flags, defaults,
// deprecation, validation, repeatable flags, aliases, environment variable
// fallback, and shell completion.
//
// Scalar flags use **T pointers and are optional (nil) by default. Use
// [Required] to make a flag mandatory or [Default] to provide a fallback.
// Defaults are applied at registration time; [Set.Parse] only checks
// required flags.
//
// Repeatable flags ([Set.StringSliceVar], [Set.MapVar]) use plain
// pointers and accumulate values across multiple occurrences.
//
// Usage strings support fmt-style arguments followed by [Option] values:
//
//	fs.StringVar(&c.name, "name", "name for %s", "workspace", Required())
//
// Examples:
//
//	var name *string
//	fs.StringVar(&name, "name", "resource name", Required())       // required, must be set
//	fs.StringVar(&name, "format", "output format", Default("json")) // optional, defaults to "json"
//	fs.StringVar(&name, "desc", "description")                      // optional, nil if not set
//
//	// Aliases let short flags work: -n is equivalent to -name.
//	fs.StringVar(&name, "name", "resource name", Aliases("n"))
//
//	// Env var fallback: uses THARSIS_TOKEN if -token is not set.
//	fs.StringVar(&token, "token", "auth token", EnvVar("THARSIS_TOKEN"))
//
//	var tags []string
//	fs.StringSliceVar(&tags, "tag", "tag (repeatable)")  // -tag a -tag b → ["a","b"]
//
//	var labels map[string]string
//	fs.MapVar(&labels, "label", "key=value pair")        // -label env=prod
//	                                                     // -label env=- removes key
package flag
