package headerparse

import (
	"fmt"
	"strings"
)

type unprocessedExifGPSData struct {
	datetime  string
	latitude  string
	longitude string
	speed     float64
	speedRef  string
}

type unprocessedExifData struct {
	startTime string
	duration  string
	gpsData   []*unprocessedExifGPSData
}

func (emt *ExifMP4Tool) extractData(rawData map[string]interface{}) (DashCamName, *unprocessedExifData, error) {
	dashCamName, err := inferDashCamName(rawData)
	if err != nil {
		return "", nil, fmt.Errorf("error inferring dashcam name: %w", err)
	}

	unprocessedData := &unprocessedExifData{}

	mainDataObj, ok := rawData[exifToolMetadataMainKey]
	if !ok {
		return "", nil, fmt.Errorf("no %q data", exifToolMetadataMainKey)
	}

	unprocessedData.startTime, unprocessedData.duration, err = extractMainMetadata(dashCamName, mainDataObj)
	if err != nil {
		return "", nil, fmt.Errorf("error extracting %q data - err: %w", exifToolMetadataMainKey, err)
	}

	for _, subMapObj := range rawData {
		subMap, ok := subMapObj.(map[string]interface{})
		if !ok {
			emt.logger.Warning("Expected map[string]interface{}")
		}

		gpsData, err := extractGPSData(dashCamName, subMap)
		if err != nil {
			return "", nil, fmt.Errorf("error extracting GPS data: %w", err)
		}
		if gpsData != nil {
			unprocessedData.gpsData = append(unprocessedData.gpsData, gpsData)
		}
	}

	unprocessedData.gpsData = removeDuplicates(unprocessedData.gpsData)
	return dashCamName, unprocessedData, nil
}

func inferDashCamName(rawData map[string]interface{}) (DashCamName, error) {
	for name, fn := range dashCamInferencesFuncMap {
		if fn(rawData) {
			return name, nil
		}
	}
	return "", fmt.Errorf("dashCam could not be inferred")
}

var dashCamInferencesFuncMap = map[DashCamName]func(rawData map[string]interface{}) bool{
	BlackVueDR750X1CH: func(rawData map[string]interface{}) bool {
		mainMetadataMapObj, ok := rawData[exifToolMetadataMainKey]
		if !ok {
			return false
		}
		mainMetadataMap, ok := mainMetadataMapObj.(map[string]interface{})
		if !ok {
			return false
		}
		copyright, ok := stringExists("Copyright", mainMetadataMap)
		if !ok {
			return false
		}

		return strings.Contains(copyright, "Pittasoft")
	},
	Garmin55: func(rawData map[string]interface{}) bool {
		mainMetadataMapObj, ok := rawData[exifToolMetadataMainKey]
		if !ok {
			return false
		}
		mainMetadataMap, ok := mainMetadataMapObj.(map[string]interface{})
		if !ok {
			return false
		}

		// These are identifiers that I hope are unique to this camera.
		majorBrand, ok := stringExists("MajorBrand", mainMetadataMap)
		if !ok {
			return false
		}
		compressorName, ok := stringExists("CompressorName", mainMetadataMap)
		if !ok {
			return false
		}
		return strings.Contains(majorBrand, "MP4 Base w/ AVC ext [ISO 14496-12:2005]") && strings.Contains(compressorName, "Ambarella AVC encoder")
	},
}

func removeDuplicates(dataList []*unprocessedExifGPSData) []*unprocessedExifGPSData {
	gpsDateTimes := map[string]bool{}
	newList := []*unprocessedExifGPSData{}
	for _, gpsData := range dataList {
		if gpsDateTimes[gpsData.datetime] {
			continue
		}
		gpsDateTimes[gpsData.datetime] = true
		newList = append(newList, gpsData)
	}
	return newList
}

func extractMainMetadata(dashCamName DashCamName, obj interface{}) (string, string, error) {
	startTime := ""
	duration := ""

	mainData, ok := obj.(map[string]interface{})
	if !ok {
		return "", "", fmt.Errorf("expected map[string]interface{} for main metadata %q, got %T", exifToolMetadataMainKey, obj)
	}

	startTimeObj, ok := mainData[cameraDataLayouts[dashCamName].videoStartTimeKey]
	if !ok {
		return "", "", fmt.Errorf("no %q data", cameraDataLayouts[dashCamName].videoStartTimeKey)
	}
	startTime, ok = startTimeObj.(string)
	if !ok {
		return "", "", fmt.Errorf("expected string for %q, got %T", cameraDataLayouts[dashCamName].videoStartTimeKey, startTimeObj)
	}

	durationObj, ok := mainData[exifDurationKey]
	if !ok {
		return "", "", fmt.Errorf("no %q data", exifDurationKey)
	}
	duration, ok = durationObj.(string)
	if !ok {
		return "", "", fmt.Errorf("expected string for %q, got %T", exifDurationKey, durationObj)
	}

	return startTime, duration, nil
}

func extractGPSData(dashCamName DashCamName, gpsMap map[string]interface{}) (*unprocessedExifGPSData, error) {
	gpsData := &unprocessedExifGPSData{}

	gpsKeys := []string{cameraDataLayouts[dashCamName].gpsTimeKey, exifGPSLatKey, exifGPSLongKey, exifGPSSpeedKey, exifGPSSpeedRefKey, exifGPSSampleTimeKey}
	for _, key := range gpsKeys {
		if _, ok := gpsMap[key]; !ok {
			return nil, nil
		}
	}

	datetime, ok := gpsMap[cameraDataLayouts[dashCamName].gpsTimeKey].(string)
	if !ok {
		return nil, fmt.Errorf("expected string for %q, got %T", cameraDataLayouts[dashCamName].gpsTimeKey, gpsMap[cameraDataLayouts[dashCamName].gpsTimeKey])
	}
	gpsData.datetime = datetime

	lat, ok := gpsMap[exifGPSLatKey].(string)
	if !ok {
		return nil, fmt.Errorf("expected string for %q, got %T", exifGPSLatKey, gpsMap[exifGPSLatKey])
	}
	gpsData.latitude = lat

	long, ok := gpsMap[exifGPSLongKey].(string)
	if !ok {
		return nil, fmt.Errorf("expected string for %q, got %T", exifGPSLongKey, gpsMap[exifGPSLongKey])
	}
	gpsData.longitude = long

	speed, ok := gpsMap[exifGPSSpeedKey].(float64)
	if !ok {
		return nil, fmt.Errorf("expected float64 for %q, got %T", exifGPSSpeedKey, gpsMap[exifGPSSpeedKey])
	}
	gpsData.speed = speed

	speedRef, ok := gpsMap[exifGPSSpeedRefKey].(string)
	if !ok {
		return nil, fmt.Errorf("expected string for %q, got %T", exifGPSSpeedRefKey, gpsMap[exifGPSSpeedRefKey])
	}
	found := false
	for _, validRef := range exifGPSSpeedRefs {
		if speedRef == validRef {
			found = true
		}
	}
	if !found {
		return nil, fmt.Errorf("invalid speedRef %q", speedRef)
	}
	gpsData.speedRef = speedRef

	return gpsData, nil
}
