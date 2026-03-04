-- Откат создания индексов на таблице counters
DROP INDEX IF EXISTS idx_counters_name;

-- Откат создания индексов на таблице gauges
DROP INDEX IF EXISTS idx_gauges_name;

-- Откат создания таблицы counters
DROP TABLE IF EXISTS counters;

-- Откат создания таблицы gauges
DROP TABLE IF EXISTS gauges;
