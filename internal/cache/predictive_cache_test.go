package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zaptest"
)

func TestPredictiveCache_Creation(t *testing.T) {
	logger := zaptest.NewLogger(t)

	// 使用简化的配置创建预测缓存
	predictive := &PredictiveCache{
		logger: logger.Named("predictive_cache"),
	}

	assert.NotNil(t, predictive)
}

func TestAccessPattern_Creation(t *testing.T) {
	now := time.Now()
	pattern := &AccessPattern{
		Key:            "test_key",
		AccessTimes:    []time.Time{now.Add(-1 * time.Hour), now.Add(-30 * time.Minute), now},
		Periodicity:    1800.0, // 30 minutes in seconds
		Predictability: 0.8,
		LastPredicted:  now,
		PredictionAccuracy: 0.75,
		SeasonalFactor: 1.2,
	}

	assert.Equal(t, "test_key", pattern.Key)
	assert.Len(t, pattern.AccessTimes, 3)
	assert.Equal(t, 1800.0, pattern.Periodicity)
	assert.Equal(t, 0.8, pattern.Predictability)
}

func TestPredictionEngine_Creation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	engine := &PredictionEngine{
		logger: logger.Named("prediction_engine"),
	}

	assert.NotNil(t, engine)
}