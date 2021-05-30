package service

import (
	"fmt"
	st "seneca/api/type"
	"seneca/internal/client/weather"
	"seneca/internal/util/data"
	"time"
)

type MockWeatherService struct {
	// key: timestamp/lat/long
	resultsMap map[string][]*weather.TimestampedWeatherCondition
	getCalls   int
}

func NewMock() *MockWeatherService {
	return &MockWeatherService{
		resultsMap: map[string][]*weather.TimestampedWeatherCondition{},
		getCalls:   0,
	}
}

func constructKey(lat *st.Latitude, long *st.Longitude) string {
	return fmt.Sprintf("%f/%f", data.LatitudeToFloat64(lat), data.LongitudeToFloat64(long))
}

func (fws *MockWeatherService) GetHistoricalWeather(timestamp time.Time, lat *st.Latitude, long *st.Longitude) ([]*weather.TimestampedWeatherCondition, error) {
	fws.getCalls++
	key := constructKey(lat, long)
	results, ok := fws.resultsMap[key]
	if !ok {
		return nil, fmt.Errorf("no results found for %q", key)
	}
	return results, nil
}

func (fws *MockWeatherService) InsertHistoricalWeather(results []*weather.TimestampedWeatherCondition, lat *st.Latitude, long *st.Longitude) {
	if _, ok := fws.resultsMap[constructKey(lat, long)]; !ok {
		fws.resultsMap[constructKey(lat, long)] = []*weather.TimestampedWeatherCondition{}
	}
	existing := fws.resultsMap[constructKey(lat, long)]
	existing = append(existing, results...)
	fws.resultsMap[constructKey(lat, long)] = existing
}

func (fws *MockWeatherService) ServiceGetCalls() int {
	return fws.getCalls
}
