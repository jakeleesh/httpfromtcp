package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string

var rn = []byte("\r\n")

func NewHeaders() Headers {
	return map[string]string{}
}

// fieldLine, so return: key, value or error
func parseHeader(fieldLine []byte) (string, string, error) {
	// Split into 2 subslices
	// Because field-value can have colon in it
	// If field-value have colon, don't want to have multiple splits
	parts := bytes.SplitN(fieldLine, []byte(":"), 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("malformed field line")
	}

	name := parts[0]
	value := bytes.TrimSpace(parts[1])

	// Optional space that should not be there
	if bytes.HasSuffix(name, []byte(" ")) {
		return "", "", fmt.Errorf("malformed field name")
	}

	return string(name), string(value), nil
}

// Parsing field lines
// Field lines look like - field-name ":" OWS field-value OWS
// Have a field-name, a colon, optional space, field-value, optional white space, and then \r\n
// Done parsing headers when encounter empty field-line, or a \r\n without any content
func (h Headers) Parse(data []byte) (int, bool, error) {
	read := 0
	done := false

	for {
		idx := bytes.Index(data[read:], rn)
		if idx == -1 {
			break
		}

		// EMPTY HEADER
		// Hit empty header, this is the ending
		if idx == 0 {
			done = true
			// Move it forward
			read += len(rn)
			break
		}

		name, value, err := parseHeader(data[read : read+idx])
		if err != nil {
			return 0, false, err
		}

		read += idx + len(rn)
		// map is already a pointer
		h[name] = value
	}

	return read, done, nil
}
