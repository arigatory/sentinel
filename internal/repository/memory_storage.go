package repository

type MemStorage struct {
	counters map[string]int64
	gauges   map[string]float64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		counters: make(map[string]int64),
		gauges:   make(map[string]float64),
	}
}

func (m *MemStorage) UpdateCounter(name string, delta int64) {
	m.counters[name] += delta
}

func (m *MemStorage) UpdateGauge(name string, value float64) {
	m.gauges[name] = value
}

func (m *MemStorage) GetCounter(name string) (int64, bool) {
	value, exists := m.counters[name]
	return value, exists
}

func (m *MemStorage) GetGauge(name string) (float64, bool) {
	value, exists := m.gauges[name]
	return value, exists
}

func (m *MemStorage) GetAllGauges() map[string]float64 {
	result := make(map[string]float64, len(m.gauges))
	for name, value := range m.gauges {
		result[name] = value
	}
	return result
}

func (m *MemStorage) GetAllCounters() map[string]int64 {
	// Создаем копию карты
	result := make(map[string]int64, len(m.counters))
	for name, value := range m.counters {
		result[name] = value
	}
	return result
}
