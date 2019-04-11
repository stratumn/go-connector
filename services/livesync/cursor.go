package livesync

import (
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/pkg/errors"
)

// CompareCursors returns the difference between relay cursors.
// Relay cursors are base64 encoded strings that look like this: ["natural",109].
// The second element of the slice is the incrementing index, this is what we want to compare.
func CompareCursors(cursor1, cursor2 string) (int, error) {
	if cursor1 == "" || cursor2 == "" {
		return strings.Compare(cursor1, cursor2), nil
	}

	c1, err := base64.StdEncoding.DecodeString(cursor1)
	if err != nil {
		return 0, errors.Wrapf(err, "%s: bad cursor", cursor1)
	}
	c2, err := base64.StdEncoding.DecodeString(cursor2)
	if err != nil {
		return 0, errors.Wrapf(err, "%s: bad cursor", cursor2)
	}

	var cursorSlice []interface{}
	err = json.Unmarshal(c1, &cursorSlice)
	if err != nil || len(cursorSlice) != 2 {
		return 0, errors.Errorf("%s: cursor is not an array", string(c1))
	}
	id1, ok := cursorSlice[1].(float64)
	if !ok {
		return 0, errors.Errorf("%s: cursor does not have an index", string(c1))
	}
	err = json.Unmarshal(c2, &cursorSlice)
	if err != nil || len(cursorSlice) != 2 {
		return 0, errors.Errorf("%s: cursor is not an array", string(c2))
	}
	id2, ok := cursorSlice[1].(float64)
	if !ok {
		return 0, errors.Errorf("%s: cursor does not have an index", string(c2))
	}

	return int(id1 - id2), nil
}
