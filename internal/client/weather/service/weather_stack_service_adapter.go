package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	st "seneca/api/type"
	"seneca/internal/client/weather"
	"seneca/internal/util"
	"seneca/internal/util/data"
	"sort"
	"strconv"
	"time"
)

// TODO(lucaloncar): utilize remote secrets.
const (
	historicalEndpoint = "https://api.weatherstack.com/historical"
	apiKeyHeaderKey    = "access_key"
	apiKey             = "7a1b09e9b45a37bd0fbaec20f04edbae"

	locationKey = "query"
	dateKey     = "historical_date"
	// Get full days weather broken down by hour.
	hourlyKey                     = "hourly"
	intervalKey                   = "interval"
	successKey                    = "success"
	subMapKeyHistorical           = "historical"
	subMapKeyHourly               = "hourly"
	subMapKeyMinutesSinceDayStart = "time"
	subMapKeyWeatherCode          = "weather_code"
	subMapKeyLocation             = "location"
	subMapKeyUTCOffset            = "utc_offset"
)

type WeatherStackService struct {
	httpClient *http.Client
	apiCalls   int64
}

func NewWeatherStackService(requestTimeout time.Duration) *WeatherStackService {
	httpClient := &http.Client{
		Timeout: requestTimeout,
		Transport: weatherStackHTTPRoundTripper{
			apiKey:              apiKey,
			wrappedRoundTripper: http.DefaultTransport,
		},
	}

	return &WeatherStackService{
		httpClient: httpClient,
	}
}

func (wss *WeatherStackService) GetHistoricalWeather(timestamp time.Time, lat *st.Latitude, long *st.Longitude) ([]*weather.TimestampedWeatherCondition, error) {
	return wss.getHistoricalWeatherRecurse(timestamp, lat, long, true)
}

func (wss *WeatherStackService) getHistoricalWeatherRecurse(timestamp time.Time, lat *st.Latitude, long *st.Longitude, recurse bool) ([]*weather.TimestampedWeatherCondition, error) {
	// TODO(lucaloncar): fix this bs
	// We're somehow in an infinite loop if we get here. Crash.
	if wss.apiCalls > 1000 {
		log.Fatalf("Made over 1000 weather API calls.  Proobably an infinite loop.")
	}

	req, err := constructHistoricalGetRequest(timestamp, lat, long)
	if err != nil {
		return nil, fmt.Errorf("error constructing GET request: %w", err)
	}

	resp, err := wss.httpClient.Do(req)
	wss.apiCalls++
	if err != nil {
		return nil, fmt.Errorf("error making GET request: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("received status code %d with message %s in response", resp.StatusCode, resp.Body)
	}

	timestampedWeatherConditions, utcOffset, err := parseHistoricalGetResponse(lat, long, resp)
	if len(timestampedWeatherConditions) == 0 || err != nil {
		return nil, fmt.Errorf("error parsing HTTP response: %w", err)
	}

	// Since we request in UTC times but the weather stack API assumes we are asking in local time, and
	// gives us weather for just that day, we might be outside of the time range we want.
	utcTime := timestamp.Add(time.Hour * time.Duration(utcOffset))
	if recurse && utcTime.Before(timestampedWeatherConditions[0].StartTime) || utcTime.After(timestampedWeatherConditions[len(timestampedWeatherConditions)-1].EndTime) {
		otherTimestampedWeatherConditions, err := wss.getHistoricalWeatherRecurse(utcTime, lat, long, false)
		if err != nil {
			return nil, fmt.Errorf("error getting other TimestampedWeatherConditions: %w", err)
		}
		timestampedWeatherConditions = append(timestampedWeatherConditions, otherTimestampedWeatherConditions...)
	}

	return timestampedWeatherConditions, nil
}

func constructHistoricalGetRequest(timestamp time.Time, lat *st.Latitude, long *st.Longitude) (*http.Request, error) {
	req, err := http.NewRequest(http.MethodGet, historicalEndpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("error initializing GET request: %w", err)
	}

	q := req.URL.Query()

	q.Add(locationKey, fmt.Sprintf("%s,%s", util.RemoveTrailingZeroes(fmt.Sprintf("%f", data.LatitudeToFloat64(lat))), util.RemoveTrailingZeroes(fmt.Sprintf("%f", data.LongitudeToFloat64(long)))))

	q.Add(dateKey, fmt.Sprintf("%d-%02d-%02d", timestamp.Year(), timestamp.Month(), timestamp.Day()))

	req.URL.RawQuery = q.Encode()

	return req, nil
}

func parseHistoricalGetResponse(lat *st.Latitude, long *st.Longitude, resp *http.Response) ([]*weather.TimestampedWeatherCondition, int, error) {
	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, fmt.Errorf("error reading response body bytes: %w", err)
	}
	jsonMap := map[string]interface{}{}
	if err := json.Unmarshal(bodyBytes, &jsonMap); err != nil {
		return nil, 0, fmt.Errorf("error unmarshalling json: %w", err)
	}

	utcOffset, err := getUTCOffset(jsonMap[subMapKeyLocation])
	if err != nil {
		return nil, 0, fmt.Errorf("error getting utc offset from jsonMap %v: %w", jsonMap, err)
	}

	historicalMap, ok := util.StringToInterfaceMapOrFalse(subMapKeyHistorical, jsonMap)
	if !ok {
		return nil, 0, fmt.Errorf("could not find submap response[%s]", subMapKeyHistorical)
	}

	output := []*weather.TimestampedWeatherCondition{}
	for k, dateMapObj := range historicalMap {
		dayStartTime, err := time.Parse("2006-01-02", k)
		if err != nil {
			return nil, 0, fmt.Errorf("error parsing key response[%s][%s] into time.Time: %w", subMapKeyHistorical, k, err)
		}

		dateMap, ok := dateMapObj.(map[string]interface{})
		if !ok {
			return nil, 0, fmt.Errorf("could not find submap response[%s][date]", subMapKeyHistorical)
		}

		hourlyListObj, ok := dateMap[subMapKeyHourly]
		if !ok {
			return nil, 0, fmt.Errorf("could not find submap response[%s][date][%s]", subMapKeyHistorical, subMapKeyHourly)
		}
		hourlyList, ok := hourlyListObj.([]interface{})
		if !ok {
			return nil, 0, fmt.Errorf("want []interface{} at key response[%s][date][%s], got %T", subMapKeyHistorical, subMapKeyHourly, hourlyListObj)
		}

		for _, hourlyMapObj := range hourlyList {
			hourlyMap, ok := hourlyMapObj.(map[string]interface{})
			if !ok {
				return nil, 0, fmt.Errorf("want map[string]interface{} at list item in response[%s][date][%s][], got %T", subMapKeyHistorical, subMapKeyHourly, hourlyMapObj)
			}

			weirdTimeMeasurementSinceDayStartObj, ok := hourlyMap[subMapKeyMinutesSinceDayStart]
			if !ok {
				return nil, 0, fmt.Errorf("could not find submap response[%s][date][%s][][%s]", subMapKeyHistorical, subMapKeyHourly, subMapKeyMinutesSinceDayStart)
			}
			weirdTimeMeasurementSinceDayStartStr, ok := weirdTimeMeasurementSinceDayStartObj.(string)
			if !ok {
				return nil, 0, fmt.Errorf("want string for response[%s][date][%s][][%s], got %T", subMapKeyHistorical, subMapKeyHourly, subMapKeyMinutesSinceDayStart, weirdTimeMeasurementSinceDayStartObj)
			}

			weirdTimeMeasurementSinceDayStart, err := strconv.Atoi(weirdTimeMeasurementSinceDayStartStr)
			if err != nil {
				return nil, 0, fmt.Errorf("error converting string %q at response[%s][date][%s][][%s] to int: %w", weirdTimeMeasurementSinceDayStartStr, subMapKeyHistorical, subMapKeyHourly, subMapKeyMinutesSinceDayStart, err)
			}

			weatherCodeObj, ok := hourlyMap[subMapKeyWeatherCode]
			if !ok {
				return nil, 0, fmt.Errorf("could not find submap response[%s][date][%s][][%s]", subMapKeyHistorical, subMapKeyHourly, subMapKeyWeatherCode)
			}

			weatherCode, ok := weatherCodeObj.(float64)
			if !ok {
				return nil, 0, fmt.Errorf("want float64 for response[%s][date][%s][][%s], got %T", subMapKeyHistorical, subMapKeyHourly, subMapKeyWeatherCode, weatherCodeObj)
			}

			// Why does this API return the time as '1300' is 13 hours since day start? God knows.
			startTime := dayStartTime.Add(time.Hour * (time.Duration(weirdTimeMeasurementSinceDayStart) / 100))
			output = append(output, &weather.TimestampedWeatherCondition{
				StartTime:   startTime.Add(time.Duration(utcOffset) * time.Hour * -1),
				EndTime:     startTime.Add(time.Hour).Add(time.Duration(utcOffset) * time.Hour * -1),
				Lat:         lat,
				Long:        long,
				Source:      weather.WeatherStack,
				WeatherCode: int(weatherCode),
			})
		}
	}

	sort.Slice(output, func(i, j int) bool { return output[i].StartTime.Before(output[j].StartTime) })

	return output, utcOffset, nil
}

func getUTCOffset(locationMapObj interface{}) (int, error) {
	if locationMapObj == nil {
		return 0, fmt.Errorf("no location map in response")
	}

	locationMap, ok := locationMapObj.(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("want map[string]interface{} for location map, got %T", locationMapObj)
	}

	utcOffsetObj, ok := locationMap[subMapKeyUTCOffset]
	if !ok {
		return 0, fmt.Errorf("nothing found at response[%s][%s]", subMapKeyLocation, subMapKeyUTCOffset)
	}

	utcOffsetStr, ok := utcOffsetObj.(string)
	if !ok {
		return 0, fmt.Errorf("want string at response[%s][%s], got %T", subMapKeyLocation, subMapKeyUTCOffset, utcOffsetObj)
	}

	utcOffsetFloat64, err := strconv.ParseFloat(utcOffsetStr, 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing %q at response[%s][%s] into float64: %w", utcOffsetStr, subMapKeyLocation, subMapKeyUTCOffset, err)
	}

	return int(utcOffsetFloat64), nil
}

type weatherStackHTTPRoundTripper struct {
	apiKey              string
	wrappedRoundTripper http.RoundTripper
}

func (trpr weatherStackHTTPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	q := req.URL.Query()

	q.Add(apiKeyHeaderKey, apiKey)
	q.Add(hourlyKey, "1")
	q.Add(intervalKey, "1")

	req.URL.RawQuery = q.Encode()

	return trpr.wrappedRoundTripper.RoundTrip(req)
}
