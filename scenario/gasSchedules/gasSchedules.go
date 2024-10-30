package gasschedules

// import _ "embed"

// //go:embed gasScheduleV1.toml
// var gasScheduleV1 string

// //go:embed gasScheduleV2.toml
// var gasScheduleV2 string

// //go:embed gasScheduleV3.toml
// var gasScheduleV3 string

//go:generate go run scripts/includetoml.go

// GetV3 yields the schedule V3
func GetV3() string {
	return gasScheduleV3
}

// GetV4 yields the schedule V4
func GetV4() string {
	return gasScheduleV4
}
