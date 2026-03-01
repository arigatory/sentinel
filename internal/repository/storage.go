package repository

type Storage interface {
	UpdateCounter(name string, delta int64)

	UpdateGauge(name string, value float64)

	GetCounter(name string) (int64, bool)

	GetGauge(name string) (float64, bool)

	GetAllGauges() map[string]float64

	GetAllCounters() map[string]int64

	Save(path string) error

	Load(path string) error
}
