package retry

import (
	"math"
	"math/rand"
	"time"
)

// ConstantBackoff 返回固定间隔的重试策略
func ConstantBackoff(interval time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		return interval
	}
}

// ExponentialBackoff 返回指数退避的重试策略
// 公式: interval * 2^attempt
func ExponentialBackoff(interval time.Duration, maxInterval time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		backoff := interval * time.Duration(math.Pow(2, float64(attempt)))
		if backoff > maxInterval {
			backoff = maxInterval
		}
		return backoff
	}
}

// ExponentialBackoffWithJitter 返回带抖动的指数退避重试策略
// 公式: random(interval * 2^attempt * (1-jitter), interval * 2^attempt)
func ExponentialBackoffWithJitter(interval time.Duration, maxInterval time.Duration, jitter float64) BackoffFunc {
	if jitter < 0 {
		jitter = 0
	}
	if jitter > 1 {
		jitter = 1
	}

	return func(attempt int) time.Duration {
		backoff := float64(interval) * math.Pow(2, float64(attempt))
		if backoff > float64(maxInterval) {
			backoff = float64(maxInterval)
		}

		min := backoff * (1 - jitter)
		max := backoff

		// 在 min 和 max 之间生成随机值
		backoff = min + rand.Float64()*(max-min)

		return time.Duration(backoff)
	}
}

// LinearBackoff 返回线性增长的重试策略
// 公式: interval * (attempt + 1)
func LinearBackoff(interval time.Duration, maxInterval time.Duration) BackoffFunc {
	return func(attempt int) time.Duration {
		backoff := interval * time.Duration(attempt+1)
		if backoff > maxInterval {
			backoff = maxInterval
		}
		return backoff
	}
}
