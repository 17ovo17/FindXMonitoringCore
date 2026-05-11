package store

import (
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type startupRetry struct {
	attempts int
	interval time.Duration
}

func storageStartupRetry(attemptsKey, intervalKey string, defaultAttempts int, defaultInterval time.Duration) startupRetry {
	attempts := viper.GetInt(attemptsKey)
	if attempts < 1 {
		attempts = defaultAttempts
	}
	return startupRetry{
		attempts: attempts,
		interval: storageRetryDuration(intervalKey, defaultInterval),
	}
}

func storageRetryDuration(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(viper.GetString(key))
	if raw != "" {
		if value, err := time.ParseDuration(raw); err == nil && value > 0 {
			return value
		}
	}
	if value := viper.GetInt(key); value > 0 {
		return time.Duration(value) * time.Second
	}
	return fallback
}

func (r startupRetry) run(label string, fn func() error) error {
	var lastErr error
	for attempt := 1; attempt <= r.attempts; attempt++ {
		if err := fn(); err != nil {
			lastErr = err
			if attempt < r.attempts {
				logrus.Warnf("%s attempt %d/%d failed, retrying: %v", label, attempt, r.attempts, err)
				time.Sleep(r.interval)
				continue
			}
			break
		}
		return nil
	}
	return lastErr
}
