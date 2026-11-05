package playlist

import (
	"strconv"
	"time"

	"github.com/bluenviron/gohlslib/v2/pkg/playlist/primitives"
)

// MediaServerControl is a EXT-X-SERVER-CONTROL tag.
type MediaServerControl struct {
	// CAN-BLOCK-RELOAD
	CanBlockReload bool

	// PART-HOLD-BACK
	// The value is a decimal-floating-point number of seconds that
	// indicates the server-recommended minimum distance from the end of
	// the Playlist at which clients should begin to play or to which
	// they should seek when playing in Low-Latency Mode.  Its value MUST
	// be at least twice the Part Target Duration.  Its value SHOULD be
	// at least three times the Part Target Duration.
	PartHoldBack *time.Duration

	// CAN-SKIP-UNTIL
	// Indicates that the Server can produce Playlist Delta Updates in
	// response to the _HLS_skip Delivery Directive.  Its value is the
	// Skip Boundary, a decimal-floating-point number of seconds.  The
	// Skip Boundary MUST be at least six times the Target Duration.
	CanSkipUntil *time.Duration

	// CAN-SKIP-DATERANGES
	// Indicates that the Server can produce Playlist Delta Updates that
	// remove EXT-X-DATERANGE tags, in response to the _HLS_skip=v2
	// Delivery Directive.
	CanSkipDateRanges bool
}

func (t *MediaServerControl) unmarshal(v string) error {
	var attrs primitives.Attributes
	err := attrs.Unmarshal(v)
	if err != nil {
		return err
	}

	for key, val := range attrs {
		switch key {
		case "CAN-BLOCK-RELOAD":
			t.CanBlockReload = (val == "YES")

		case "PART-HOLD-BACK":
			var d primitives.Duration
			err = d.Unmarshal(val)
			if err != nil {
				return err
			}
			tmp := time.Duration(d)
			t.PartHoldBack = &tmp

		case "CAN-SKIP-UNTIL":
			var d primitives.Duration
			err = d.Unmarshal(val)
			if err != nil {
				return err
			}
			tmp := time.Duration(d)
			t.CanSkipUntil = &tmp

		case "CAN-SKIP-DATERANGES":
			t.CanSkipDateRanges = (val == "YES")
		}
	}

	return nil
}

func (t MediaServerControl) marshal() string {
	ret := "#EXT-X-SERVER-CONTROL:"

	if t.CanBlockReload {
		ret += "CAN-BLOCK-RELOAD=YES"
	}

	if t.PartHoldBack != nil {
		ret += ",PART-HOLD-BACK=" + strconv.FormatFloat(t.PartHoldBack.Seconds(), 'f', 5, 64)
	}

	if t.CanSkipUntil != nil {
		ret += ",CAN-SKIP-UNTIL=" + strconv.FormatFloat(t.CanSkipUntil.Seconds(), 'f', 5, 64)
	}

	if t.CanSkipDateRanges {
		ret += ",CAN-SKIP-DATERANGES=YES"
	}

	ret += "\n"

	return ret
}
