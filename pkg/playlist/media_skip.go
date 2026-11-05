package playlist

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bluenviron/gohlslib/v2/pkg/playlist/primitives"
)

// MediaSkip is a EXT-X-SKIP tag.
type MediaSkip struct {
	// SKIPPED-SEGMENTS
	// required
	SkippedSegments int

	// RECENTLY-REMOVED-DATERANGES
	// quoted string containing the IDs of the removed Date Ranges,
	// separated by tab characters.
	RecentlyRemovedDateRanges []string
}

func (t *MediaSkip) unmarshal(v string) error {
	var attrs primitives.Attributes
	err := attrs.Unmarshal(v)
	if err != nil {
		return err
	}

	skipSegFound := false

	for key, val := range attrs {
		switch key {
		case "SKIPPED-SEGMENTS":
			var tmp uint64
			tmp, err = strconv.ParseUint(val, 10, 31)
			if err != nil {
				return err
			}
			t.SkippedSegments = int(tmp)
			skipSegFound = true

		case "RECENTLY-REMOVED-DATERANGES":
			if val != "" {
				t.RecentlyRemovedDateRanges = strings.Split(val, "\t")
			}
		}
	}

	if !skipSegFound {
		return fmt.Errorf("SKIPPED-SEGMENTS missing")
	}

	return nil
}

func (t MediaSkip) marshal() string {
	ret := "#EXT-X-SKIP:SKIPPED-SEGMENTS=" + strconv.FormatInt(int64(t.SkippedSegments), 10)

	if len(t.RecentlyRemovedDateRanges) != 0 {
		ret += `,RECENTLY-REMOVED-DATERANGES="` +
			strings.Join(t.RecentlyRemovedDateRanges, "\t") + `"`
	}

	ret += "\n"

	return ret
}
