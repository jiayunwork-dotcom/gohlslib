package playlist

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/bluenviron/gohlslib/v2/pkg/playlist/primitives"
)

// MediaDateRange is a EXT-X-DATERANGE tag (RFC 8216, section 4.3.2.7).
type MediaDateRange struct {
	// ID
	// required
	ID string

	// CLASS
	Class string

	// START-DATE
	// required
	StartDate time.Time

	// END-DATE
	EndDate *time.Time

	// DURATION
	Duration *time.Duration

	// PLANNED-DURATION
	PlannedDuration *time.Duration

	// END-ON-NEXT
	EndOnNext bool

	// SCTE35-CMD, stored with its original 0x prefix.
	SCTE35Cmd *string

	// SCTE35-OUT, stored with its original 0x prefix.
	SCTE35Out *string

	// SCTE35-IN, stored with its original 0x prefix.
	SCTE35In *string

	// client-defined attributes.
	// Keys include the X- prefix, values do not include quotes.
	XAttrs map[string]string
}

func (d *MediaDateRange) unmarshal(v string) error {
	var attrs primitives.Attributes
	err := attrs.Unmarshal(v)
	if err != nil {
		return err
	}

	startDateFound := false

	for key, val := range attrs {
		switch {
		case key == "ID":
			d.ID = val

		case key == "CLASS":
			d.Class = val

		case key == "START-DATE":
			var tmp time.Time
			tmp, err = parseTime(val)
			if err != nil {
				return err
			}
			d.StartDate = tmp
			startDateFound = true

		case key == "END-DATE":
			var tmp time.Time
			tmp, err = parseTime(val)
			if err != nil {
				return err
			}
			d.EndDate = &tmp

		case key == "DURATION":
			var tmp primitives.Duration
			err = tmp.Unmarshal(val)
			if err != nil {
				return err
			}
			dv := time.Duration(tmp)
			d.Duration = &dv

		case key == "PLANNED-DURATION":
			var tmp primitives.Duration
			err = tmp.Unmarshal(val)
			if err != nil {
				return err
			}
			dv := time.Duration(tmp)
			d.PlannedDuration = &dv

		case key == "END-ON-NEXT":
			d.EndOnNext = (val == "YES")

		case key == "SCTE35-CMD":
			d.SCTE35Cmd = &val

		case key == "SCTE35-OUT":
			d.SCTE35Out = &val

		case key == "SCTE35-IN":
			d.SCTE35In = &val

		case strings.HasPrefix(key, "X-"):
			if d.XAttrs == nil {
				d.XAttrs = make(map[string]string)
			}
			d.XAttrs[key] = val
		}
	}

	if d.ID == "" {
		return fmt.Errorf("DATERANGE ID missing")
	}

	if !startDateFound {
		return fmt.Errorf("DATERANGE START-DATE missing")
	}

	if d.EndDate != nil && d.EndDate.Before(d.StartDate) {
		return fmt.Errorf("DATERANGE END-DATE is earlier than START-DATE")
	}

	if d.Duration != nil && *d.Duration < 0 {
		return fmt.Errorf("DATERANGE DURATION is negative")
	}

	if d.PlannedDuration != nil && *d.PlannedDuration < 0 {
		return fmt.Errorf("DATERANGE PLANNED-DURATION is negative")
	}

	if d.EndDate != nil && d.Duration != nil &&
		!d.EndDate.Equal(d.StartDate.Add(*d.Duration)) {
		return fmt.Errorf("DATERANGE END-DATE is not equal to START-DATE plus DURATION")
	}

	if d.EndOnNext {
		if d.Class == "" {
			return fmt.Errorf("DATERANGE END-ON-NEXT requires CLASS")
		}

		if d.Duration != nil || d.EndDate != nil {
			return fmt.Errorf("DATERANGE END-ON-NEXT must not be used with DURATION or END-DATE")
		}
	}

	return nil
}

// validateConsistency checks the tag against a previous tag with the same ID.
func (d MediaDateRange) validateConsistency(prev *MediaDateRange) error {
	if !d.StartDate.Equal(prev.StartDate) {
		return fmt.Errorf("DATERANGE with ID %s has an inconsistent START-DATE", d.ID)
	}

	if d.Class != "" && prev.Class != "" && d.Class != prev.Class {
		return fmt.Errorf("DATERANGE with ID %s has an inconsistent CLASS", d.ID)
	}

	if d.EndDate != nil && prev.EndDate != nil && !d.EndDate.Equal(*prev.EndDate) {
		return fmt.Errorf("DATERANGE with ID %s has an inconsistent END-DATE", d.ID)
	}

	if d.Duration != nil && prev.Duration != nil && *d.Duration != *prev.Duration {
		return fmt.Errorf("DATERANGE with ID %s has an inconsistent DURATION", d.ID)
	}

	if d.PlannedDuration != nil && prev.PlannedDuration != nil &&
		*d.PlannedDuration != *prev.PlannedDuration {
		return fmt.Errorf("DATERANGE with ID %s has an inconsistent PLANNED-DURATION", d.ID)
	}

	if d.SCTE35Cmd != nil && prev.SCTE35Cmd != nil && *d.SCTE35Cmd != *prev.SCTE35Cmd {
		return fmt.Errorf("DATERANGE with ID %s has an inconsistent SCTE35-CMD", d.ID)
	}

	if d.SCTE35Out != nil && prev.SCTE35Out != nil && *d.SCTE35Out != *prev.SCTE35Out {
		return fmt.Errorf("DATERANGE with ID %s has an inconsistent SCTE35-OUT", d.ID)
	}

	if d.SCTE35In != nil && prev.SCTE35In != nil && *d.SCTE35In != *prev.SCTE35In {
		return fmt.Errorf("DATERANGE with ID %s has an inconsistent SCTE35-IN", d.ID)
	}

	for key, val := range d.XAttrs {
		if prevVal, ok := prev.XAttrs[key]; ok && prevVal != val {
			return fmt.Errorf("DATERANGE with ID %s has an inconsistent %s", d.ID, key)
		}
	}

	return nil
}

func (d MediaDateRange) marshal() string {
	var ret strings.Builder

	ret.WriteString(`#EXT-X-DATERANGE:ID="` + d.ID + `"`)

	if d.Class != "" {
		ret.WriteString(`,CLASS="` + d.Class + `"`)
	}

	ret.WriteString(",START-DATE=" + d.StartDate.Format(timeRFC3339Millis))

	if d.EndDate != nil {
		ret.WriteString(",END-DATE=" + d.EndDate.Format(timeRFC3339Millis))
	}

	if d.Duration != nil {
		ret.WriteString(",DURATION=" + strconv.FormatFloat(d.Duration.Seconds(), 'f', 5, 64))
	}

	if d.PlannedDuration != nil {
		ret.WriteString(",PLANNED-DURATION=" +
			strconv.FormatFloat(d.PlannedDuration.Seconds(), 'f', 5, 64))
	}

	xKeys := make([]string, 0, len(d.XAttrs))
	for key := range d.XAttrs {
		xKeys = append(xKeys, key)
	}
	sort.Strings(xKeys)

	for _, key := range xKeys {
		val := d.XAttrs[key]
		ret.WriteString("," + key + "=")
		if strings.ContainsAny(val, ", ") {
			ret.WriteString(`"` + val + `"`)
		} else {
			ret.WriteString(val)
		}
	}

	if d.EndOnNext {
		ret.WriteString(",END-ON-NEXT=YES")
	}

	if d.SCTE35Cmd != nil {
		ret.WriteString(",SCTE35-CMD=" + *d.SCTE35Cmd)
	}

	if d.SCTE35Out != nil {
		ret.WriteString(",SCTE35-OUT=" + *d.SCTE35Out)
	}

	if d.SCTE35In != nil {
		ret.WriteString(",SCTE35-IN=" + *d.SCTE35In)
	}

	ret.WriteString("\n")

	return ret.String()
}
