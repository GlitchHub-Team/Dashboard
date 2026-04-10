package main

import "encoding/json"

func (s *EcgData) Serialize() ([]byte, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (s *HeartRateData) Serialize() ([]byte, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (s *PulseOximeterData) Serialize() ([]byte, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (s *HealthThermometerData) Serialize() ([]byte, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func (s *EnvironmentalSensingData) Serialize() ([]byte, error) {
	jsonData, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}
