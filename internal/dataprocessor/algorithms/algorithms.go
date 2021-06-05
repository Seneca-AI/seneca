package algorithms

import (
	"fmt"
	"seneca/internal/client/weather"
	"seneca/internal/dataprocessor"
)

type AlgorithmFactory struct {
	algorithmsCatalogue map[string]dataprocessor.AlgorithmInterface
	weatherService      weather.WeatherServiceInterface
}

func NewFactory(weatherService weather.WeatherServiceInterface) (*AlgorithmFactory, error) {
	factory := &AlgorithmFactory{
		algorithmsCatalogue: map[string]dataprocessor.AlgorithmInterface{},
		weatherService:      weatherService,
	}

	base := newBase()
	factory.algorithmsCatalogue[base.Tag()] = base

	decelerationV0, err := newDecelerationV0()
	if err != nil {
		return nil, fmt.Errorf("newDecelerationV0() returns err: %w", err)
	}
	factory.algorithmsCatalogue[decelerationV0.Tag()] = decelerationV0

	accelerationV0, err := newAccelerationV0()
	if err != nil {
		return nil, fmt.Errorf("newAccelerationV0() returns err: %w", err)
	}
	factory.algorithmsCatalogue[accelerationV0.Tag()] = accelerationV0

	weatherV0 := newWeatherV0(weatherService)
	factory.algorithmsCatalogue[weatherV0.Tag()] = weatherV0

	return factory, nil
}

func (af *AlgorithmFactory) GetAlgorithm(algoTag string) (dataprocessor.AlgorithmInterface, error) {
	algo, ok := af.algorithmsCatalogue[algoTag]
	if !ok {
		return nil, fmt.Errorf("no algorithm with tag %q", algoTag)
	}
	return algo, nil
}
