package stringsext

type KVDictString = KVDict[string, string]

// ParseKVDictString parses a string into a key-value dictionary.
//
// The input string `s` is expected to be in a format where keys and values are separated by `=` and pairs are separated by `,`.
// Values can optionally be quoted with double quotes, allowing for special characters such as commas, equals signs, and quotes.
// Special characters can be escaped with a backslash `\` to be interpreted literally.
//
// The function returns:
//   - A [KVDictString] containing the parsed key-value pairs.
//   - A slice of strings (`keyord`) where each element corresponds to the starting key of each pair.
//   - An error if the parsing fails due to invalid syntax.
//
// Examples:
//
//	dict, keyord, err := ParseKVDictString(`key1=value1,key2="value, with comma",key3=value3`)
//	if err != nil {
//	    log.Fatalf("Failed to parse KVDict: %v", err)
//	}
//	fmt.Println(dict)    // Output: map[key1:value1 key2:value, with comma key3:value3]
//	fmt.Println(keyord) // Output: [key1 key2 key3]
//
// Notes:
//   - The function supports escape sequences for special characters within keys and values.
//   - If a key or value is not terminated by a comma, it is considered the final pair.
//   - The function validates the syntax and returns an error if it encounters unexpected symbols or malformed input.
//
//nolint:gocognit,exhaustive,cyclop
func ParseKVDictString(s string) (dict KVDictString, keyord []string, err error) {
	return ParseKVDict[string, string](s)
}
