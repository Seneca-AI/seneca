package client

import (
	"fmt"
	"math"
	"seneca/api/senecaerror"
	st "seneca/api/type"
	"seneca/internal/client/weather"
	"seneca/internal/util"
	"seneca/internal/util/data"
	"sort"
	"time"
)

type cacheValue struct {
	lat      *st.Latitude
	long     *st.Longitude
	rangeMap *util.RangeMap
}

//	WeatherClient caches the results of the WeatherService by checking if the Lat/Long in the request was within some distance of a cached value.
//nolint
type WeatherClient struct {
	WeatherInterface weather.WeatherServiceInterface
	// weather client is for each user.  key is lat/long, values are RangeMaps with ranges on the timestamps holding the TimestampedWeatherCondition.
	cache       map[string]*cacheValue
	cacheRadius uint
	// Infinite loop protection.
	calls int
}

// 	New returns a new WeatherClient.
//	Params:
//		wrappedWeatherServiceInterface WeatherInterface: the weather service to use
//		radiusMiles uint: 	the radius that a weather report will cover.  Reports are anchored at certain coordinates, but we re-use them for nearby coordinates.
//						  	The radius is not inclusive, and we round distances  e.g. if a distance between two coordinates is 4 and this value is set to 4, it will
//						  	not be included in the cache.
//	Returns:
//		*WeatherClient
func New(wrappedWeatherServiceInterface weather.WeatherServiceInterface, radiusMiles uint) *WeatherClient {
	return &WeatherClient{
		WeatherInterface: wrappedWeatherServiceInterface,
		cache:            map[string]*cacheValue{},
		cacheRadius:      radiusMiles,
	}
}

func (wc *WeatherClient) GetHistoricalWeather(timestamp time.Time, lat *st.Latitude, long *st.Longitude) (*weather.TimestampedWeatherCondition, error) {
	defer func() { wc.calls = 0 }()
	wc.calls++
	if wc.calls > 10000 {
		return nil, senecaerror.NewDevError(fmt.Errorf("weatherclient.GetHistoricalWeather() made 10000 recursive calls, so it's probably in an infinite loop"))
	}

	// Search for something in range.
	var cacheVal *cacheValue
	for _, val := range wc.cache {
		if uint(math.Round(data.DistanceMiles(lat, long, val.lat, val.long))) < wc.cacheRadius {
			cacheVal = val
			break
		}
	}

	if cacheVal != nil {
		twcObj, ok := cacheVal.rangeMap.Get(util.TimeToMilliseconds(timestamp))
		if ok {
			twc, ok := twcObj.(*weather.TimestampedWeatherCondition)
			if !ok {
				return nil, senecaerror.NewDevError(fmt.Errorf("want *TimestampedWeatherCondition in cache.rangeMap, got %T", twcObj))
			}
			twc.Lat = lat
			twc.Long = long
			return twc, nil
		}
	}

	timestampedWeatherConditions, err := wc.WeatherInterface.GetHistoricalWeather(timestamp, lat, long)
	if err != nil {
		return nil, fmt.Errorf("wrappedWeatherServiceInterface.GetHistoricalWeather() returns err: %w", err)
	}

	sort.Slice(timestampedWeatherConditions, func(i, j int) bool {
		return timestampedWeatherConditions[i].StartTime.Before(timestampedWeatherConditions[j].StartTime)
	})

	// Update cache.
	if err := wc.updateCache(cacheVal, lat, long, timestampedWeatherConditions); err != nil {
		return nil, fmt.Errorf("updateCache() returns err: %w", err)
	}

	// Now that the cache is filled, this should work.
	return wc.GetHistoricalWeather(timestamp, lat, long)
}

func (wc *WeatherClient) updateCache(cacheVal *cacheValue, lat *st.Latitude, long *st.Longitude, timestampedWeatherConditions []*weather.TimestampedWeatherCondition) error {
	if len(timestampedWeatherConditions) == 0 {
		return fmt.Errorf("attempting to insert empty TimestampedWeatherConditions into the cache")
	}

	if cacheVal == nil {
		rangeMap, err := util.NewRangeMap(nil, nil)
		if err != nil {
			return fmt.Errorf("NewRangeMap(nil, nil) returns err: %w", err)
		}

		for _, t := range timestampedWeatherConditions {
			rn := util.Range{L: util.TimeToMilliseconds(t.StartTime), U: util.TimeToMilliseconds(t.EndTime)}
			if err := rangeMap.Insert(rn, t); err != nil {
				return fmt.Errorf("rangeMap.Insert(%q, _) returns err: %w", rn, err)
			}
		}

		wc.cache[fmt.Sprintf("%f/%f", data.LatitudeToFloat64(lat), data.LongitudeToFloat64(long))] = &cacheValue{
			lat:      lat,
			long:     long,
			rangeMap: rangeMap,
		}
	} else {
		for _, t := range timestampedWeatherConditions {
			rn := util.Range{L: util.TimeToMilliseconds(t.StartTime), U: util.TimeToMilliseconds(t.EndTime)}
			if err := cacheVal.rangeMap.Insert(rn, t); err != nil {
				return fmt.Errorf("RangeMap().Insert(%q, _) returns err: %w", rn, err)
			}
		}
		wc.cache[fmt.Sprintf("%f/%f", data.LatitudeToFloat64(cacheVal.lat), data.LongitudeToFloat64(cacheVal.long))] = cacheVal
	}

	return nil
}
