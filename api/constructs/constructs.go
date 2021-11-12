package constructs

import "activity_log/internal/util"

type UserInput struct {
	Text string
}

type UserSchema struct {
	Schema *util.ExpandingMap
}

type UserData struct {
	Data        map[string]interface{}
	TimestampMS int64
}

type UserDataKey string

var (
	Activity     UserDataKey = "ACTIVITY"
	MinutesSpent UserDataKey = "MINUTES"
)
