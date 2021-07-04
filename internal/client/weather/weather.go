package weather

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/util/data"
	"time"
)

//nolint
type WeatherSource string

var (
	WeatherStack      WeatherSource = "WeatherStack"
	FakeWeatherSource WeatherSource = "FakeWeatherSource"
)

type TimestampedWeatherCondition struct {
	StartTime   time.Time
	EndTime     time.Time
	Source      WeatherSource
	Lat         *st.Latitude
	Long        *st.Longitude
	WeatherCode int
}

func (twc TimestampedWeatherCondition) String() string {
	return fmt.Sprintf(
		"{ \n\tTimePeriod: { %v - %v }, \n\tSource: %s, \n\tLatLong: (%f, %f), \n\tWeatherCode: %d \n}",
		twc.StartTime,
		twc.EndTime,
		string(twc.Source),
		data.LatitudeToFloat64(twc.Lat),
		data.LongitudeToFloat64(twc.Long),
		twc.WeatherCode,
	)
}

//nolint
type WeatherServiceInterface interface {
	GetHistoricalWeather(timestamp time.Time, lat *st.Latitude, long *st.Longitude) ([]*TimestampedWeatherCondition, error)
}
