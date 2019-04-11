package livesync

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

func parseCursor(cursor string) (float64, error) {
	c, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, errors.Wrapf(err, "%s: bad cursor", cursor)
	}
	var cursorSlice []interface{}
	err = json.Unmarshal(c, &cursorSlice)
	if err != nil || len(cursorSlice) != 2 {
		return 0, errors.Errorf("%s: cursor is not an array", string(c))
	}
	index, ok := cursorSlice[1].(float64)
	if !ok {
		return 0, errors.Errorf("%s: cursor does not have an index", string(c))
	}
	return index, nil
}

// CompareCursors returns the difference between relay cursors.
// Relay cursors are base64 encoded strings that look like this: ["natural",109].
// The second element of the slice is the incrementing index, this is what we want to compare.
func CompareCursors(cursor1, cursor2 string) (int, error) {
	if cursor1 == "" || cursor2 == "" {
		return strings.Compare(cursor1, cursor2), nil
	}

	index1, err := parseCursor(cursor1)
	if err != nil {
		return 0, err
	}
	index2, err := parseCursor(cursor2)
	if err != nil {
		return 0, err
	}

	return int(index1 - index2), nil
}
