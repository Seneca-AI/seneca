package client

import (
	"log"
	"math"
	st "seneca/api/type"
	"seneca/internal/client/weather"
	"seneca/internal/client/weather/service"
	"seneca/internal/util/data"
	"testing"
	"time"
)

func TestGetHistoricalWeather(t *testing.T) {
	minutesIncrease := 1
	secondsIncrease := float64(27)

	latBase := &st.Latitude{
		Degrees:       int32(50),
		DegreeMinutes: int32(1),
		DegreeSeconds: float64(35.67),
		LatDirection:  st.Latitude_NORTH,
	}
	longBase := &st.Longitude{
		Degrees:       int32(50),
		DegreeMinutes: int32(1),
		DegreeSeconds: float64(35.67),
		LongDirection: st.Longitude_WEST,
	}

	// Every 3rd location will be included in the radius.  We'll increase each location by a degreeMinute.
	radius := data.DistanceMiles(
		latBase,
		longBase,
		&st.Latitude{
			Degrees:       latBase.Degrees,
			DegreeMinutes: latBase.DegreeMinutes + int32(minutesIncrease),
			DegreeSeconds: latBase.DegreeSeconds + float64(secondsIncrease),
			LatDirection:  latBase.LatDirection,
		},
		&st.Longitude{
			Degrees:       longBase.Degrees,
			DegreeMinutes: longBase.DegreeMinutes + int32(minutesIncrease),
			DegreeSeconds: longBase.DegreeSeconds + float64(secondsIncrease),
			LongDirection: longBase.LongDirection,
		},
	) * float64(4)

	fakeWeatherService := service.NewMock()
	weatherClient := New(fakeWeatherService, uint(math.Round(radius)))

	timestamps := []time.Time{}
	latitudes := []*st.Latitude{}
	longitudes := []*st.Longitude{}

	timeStart := time.Date(2021, 5, 30, 2021, 0, 0, 0, time.UTC)

	for i := 0; i < 24; i++ {
		lat := &st.Latitude{
			Degrees:       latBase.Degrees,
			DegreeMinutes: latBase.DegreeMinutes + int32(i*minutesIncrease),
			DegreeSeconds: latBase.DegreeSeconds + float64(i)*secondsIncrease,
			LatDirection:  st.Latitude_NORTH,
		}
		long := &st.Longitude{
			Degrees:       longBase.Degrees,
			DegreeMinutes: longBase.DegreeMinutes + int32(i*minutesIncrease),
			DegreeSeconds: longBase.DegreeSeconds + float64(i)*secondsIncrease,
			LongDirection: st.Longitude_WEST,
		}
		latitudes = append(latitudes, lat)
		longitudes = append(longitudes, long)
		timestamps = append(timestamps, timeStart.Add(time.Hour*time.Duration(i)))

		if i%4 == 0 {
			twcs := []*weather.TimestampedWeatherCondition{}
			for j := 0; j < 24; j++ {
				twc := &weather.TimestampedWeatherCondition{
					StartTime:   timeStart.Add(time.Hour * time.Duration(j)),
					EndTime:     timeStart.Add((time.Hour * time.Duration(j+1))),
					Source:      weather.FakeWeatherSource,
					Lat:         lat,
					Long:        long,
					WeatherCode: i / 4,
				}
				twcs = append(twcs, twc)
			}
			fakeWeatherService.InsertHistoricalWeather(twcs, lat, long)
		}
	}

	wantWeatherCodes := []int{0, 1, 2, 3, 4, 5, 6}
	gotTWCs := []*weather.TimestampedWeatherCondition{}

	for i := 0; i < 24; i++ {
		twc, err := weatherClient.GetHistoricalWeather(timestamps[i], latitudes[i], longitudes[i])
		if err != nil {
			t.Fatalf("weatherClient.GetHistoricalWeather() returns err: %v", err)
		}
		gotTWCs = append(gotTWCs, twc)
	}

	for i, twc := range gotTWCs {
		gotLocation := &st.Location{
			Lat:  twc.Lat,
			Long: twc.Long,
		}
		wantLocation := &st.Location{
			Lat:  latitudes[i],
			Long: longitudes[i],
		}
		if !data.LocationsEqual(wantLocation, gotLocation) {
			t.Fatalf("Want location %v, got location %v", wantLocation, gotLocation)
		}

		wantStartTime := timeStart.Add(time.Hour * time.Duration(i))
		if wantStartTime != twc.StartTime {
			t.Fatalf("Want startTime %v, got %v", wantStartTime, twc.StartTime)
		}

		wantEndTime := timeStart.Add(time.Hour * time.Duration(i+1))
		if wantEndTime != twc.EndTime {
			t.Fatalf("Want endTime %v, got %v", wantEndTime, twc.EndTime)
		}

		wantWeatherCode := wantWeatherCodes[i/4]
		if wantWeatherCode != twc.WeatherCode {
			t.Fatalf("Want weatherCode %d, got %d", wantWeatherCode, twc.WeatherCode)
		}
	}

	if fakeWeatherService.ServiceGetCalls() > 6 {
		log.Fatalf("Want 4 API Get calls, got %d", fakeWeatherService.ServiceGetCalls())
	}
}

func TestGetHistoricalWeatherChecksTimestamp(t *testing.T) {
	lat := &st.Latitude{
		Degrees:       int32(50),
		DegreeMinutes: int32(1),
		DegreeSeconds: float64(35.67),
		LatDirection:  st.Latitude_NORTH,
	}
	long := &st.Longitude{
		Degrees:       int32(50),
		DegreeMinutes: int32(1),
		DegreeSeconds: float64(35.67),
		LongDirection: st.Longitude_WEST,
	}

	fakeWeatherService := service.NewMock()
	weatherClient := New(fakeWeatherService, 1)
	timeStart := time.Date(2021, 5, 30, 2021, 0, 0, 0, time.UTC)

	// First test for every other hour.
	for i := 0; i < 12; i++ {
		twcs := []*weather.TimestampedWeatherCondition{}
		twc := &weather.TimestampedWeatherCondition{
			StartTime:   timeStart.Add(time.Hour * time.Duration(i*2)),
			EndTime:     timeStart.Add(time.Hour * time.Duration(i*2+1)),
			Source:      weather.FakeWeatherSource,
			Lat:         lat,
			Long:        long,
			WeatherCode: i,
		}
		twcs = append(twcs, twc)
		fakeWeatherService.InsertHistoricalWeather(twcs, lat, long)
	}

	gotTWCs := []*weather.TimestampedWeatherCondition{}

	for i := 0; i < 12; i++ {
		twc, err := weatherClient.GetHistoricalWeather(timeStart.Add(time.Hour*time.Duration(i*2)), lat, long)
		if err != nil {
			t.Fatalf("weatherClient.GetHistoricalWeather() returns err: %v", err)
		}
		gotTWCs = append(gotTWCs, twc)
	}

	for i, twc := range gotTWCs {
		wantStartTime := timeStart.Add(time.Hour * time.Duration(i*2))
		if wantStartTime != twc.StartTime {
			t.Fatalf("Want startTime %v, got %v", wantStartTime, twc.StartTime)
		}

		wantEndTime := timeStart.Add(time.Hour * time.Duration(i*2+1))
		if wantEndTime != twc.EndTime {
			t.Fatalf("Want endTime %v, got %v", wantEndTime, twc.EndTime)
		}

		wantWeatherCode := i
		if wantWeatherCode != twc.WeatherCode {
			t.Fatalf("Want weatherCode %d, got %d", wantWeatherCode, twc.WeatherCode)
		}
	}

	if fakeWeatherService.ServiceGetCalls() != 1 {
		t.Fatalf("Want 1 from fakeWeatherService.ServiceGetCalls(), got %d", fakeWeatherService.ServiceGetCalls())
	}

	// Now test for the next hours.
	for i := 0; i < 12; i++ {
		twcs := []*weather.TimestampedWeatherCondition{}
		twc := &weather.TimestampedWeatherCondition{
			StartTime:   timeStart.Add(time.Hour * time.Duration(i*2+1)),
			EndTime:     timeStart.Add(time.Hour * time.Duration(i*2+2)),
			Source:      weather.FakeWeatherSource,
			Lat:         lat,
			Long:        long,
			WeatherCode: i,
		}
		twcs = append(twcs, twc)
		fakeWeatherService.InsertHistoricalWeather(twcs, lat, long)
	}

	gotTWCs = []*weather.TimestampedWeatherCondition{}

	for i := 0; i < 12; i++ {
		twc, err := weatherClient.GetHistoricalWeather(timeStart.Add(time.Hour*time.Duration(i*2+1)), lat, long)
		if err != nil {
			t.Fatalf("weatherClient.GetHistoricalWeather() returns err: %v", err)
		}
		gotTWCs = append(gotTWCs, twc)
	}

	for i, twc := range gotTWCs {
		wantStartTime := timeStart.Add(time.Hour * time.Duration(i*2+1))
		if wantStartTime != twc.StartTime {
			t.Fatalf("Want startTime %v, got %v", wantStartTime, twc.StartTime)
		}

		wantEndTime := timeStart.Add(time.Hour * time.Duration(i*2+2))
		if wantEndTime != twc.EndTime {
			t.Fatalf("Want endTime %v, got %v", wantEndTime, twc.EndTime)
		}

		wantWeatherCode := i
		if wantWeatherCode != twc.WeatherCode {
			t.Fatalf("Want weatherCode %d, got %d", wantWeatherCode, twc.WeatherCode)
		}
	}

	if fakeWeatherService.ServiceGetCalls() != 2 {
		t.Fatalf("Want 2 from fakeWeatherService.ServiceGetCalls(), got %d", fakeWeatherService.ServiceGetCalls())
	}
}
