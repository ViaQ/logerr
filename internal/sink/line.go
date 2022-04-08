package sink

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// Keys used to log specific builtin fields
const (
	TimeStampKey = "_ts"
	FileLineKey  = "_file:line"
	LevelKey     = "_level"
	ComponentKey = "_component"
	MessageKey   = "_message"
	ErrorKey     = "_error"
)

// Line orders log line fields
type Line struct {
	Timestamp string
	FileLine  string
	Verbosity string
	Component string
	Message   string
	Context   map[string]interface{}
}

// LineJSON add json tags to Line struct (production logs)
type LineJSON struct {
	Timestamp string                 `json:"_ts"`
	FileLine  string                 `json:"-"`
	Verbosity string                 `json:"_level"`
	Component string                 `json:"_component"`
	Message   string                 `json:"_message"`
	Context   map[string]interface{} `json:"-"`
}

// LineJSONDev add json tags to Line struct (developer logs, enable using environment variable LOG_DEV)
type LineJSONDev struct {
	Timestamp string                 `json:"_ts"`
	FileLine  string                 `json:"_file:line"`
	Verbosity string                 `json:"_level"`
	Component string                 `json:"_component"`
	Message   string                 `json:"_message"`
	Context   map[string]interface{} `json:"-"`
}

// MarshalJSON implements custom marshaling for log line: (1) flattening context (2) support for developer mode
func (l Line) MarshalJSON() ([]byte, error) {
	lineTemp := LineJSON(l)

	lineValue, err := json.Marshal(lineTemp)
	if err != nil {
		return nil, err
	}
	verbosity, errConvert := strconv.Atoi(l.Verbosity)
	if verbosity > 1 && errConvert == nil {
		lineTempDev := LineJSONDev(l)
		lineValue, err = json.Marshal(lineTempDev)
		if err != nil {
			return nil, err
		}
	}
	lineValue = lineValue[1 : len(lineValue)-1]

	contextValue, err := json.Marshal(lineTemp.Context)
	if err != nil {
		return nil, err
	}
	contextValue = contextValue[1 : len(contextValue)-1]

	sep := ""
	if len(contextValue) > 0 {
		sep = ","
	}
	return []byte(fmt.Sprintf("{%s%s%s}", lineValue, sep, contextValue)), nil
}
