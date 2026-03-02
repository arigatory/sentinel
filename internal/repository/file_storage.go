package repository

import (
	"encoding/json"
	"os"

	models "github.com/arigatory/sentinel/internal/model"
)

func (m *MemStorage) Save(path string) error {
	gauges := m.GetAllGauges()
	counters := m.GetAllCounters()

	metrics := make([]models.Metrics, 0, len(gauges)+len(counters))

	for name, value := range gauges {
		v := value
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &v,
		})
	}
	for name, delta := range counters {
		d := delta
		metrics = append(metrics, models.Metrics{
			ID:    name,
			MType: "counter",
			Delta: &d,
		})
	}

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (m *MemStorage) Load(path string) error {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	var metrics []models.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return err
	}

	for _, metric := range metrics {
		switch metric.MType {
		case "gauge":
			if metric.Value != nil {
				m.UpdateGauge(metric.ID, *metric.Value)
			}
		case "counter":
			if metric.Delta != nil {
				m.UpdateCounter(metric.ID, *metric.Delta)
			}
		}
	}
	return nil
}
