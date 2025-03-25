package retry

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrMaxAttemptsReached 表示达到最大重试次数
	ErrMaxAttemptsReached = errors.New("maximum retry attempts reached")
	// ErrContextCanceled 表示上下文被取消
	ErrContextCanceled = errors.New("context canceled")
	// ErrContextDeadlineExceeded 表示上下文超时
	ErrContextDeadlineExceeded = errors.New("context deadline exceeded")
)

// RetryableFunc 是可重试的函数类型
type RetryableFunc func() error

// RetryableFuncWithContext 是带上下文的可重试函数类型
type RetryableFuncWithContext func(ctx context.Context) error

// IsRetryableFunc 判断错误是否可重试的函数类型
type IsRetryableFunc func(err error) bool

// BackoffFunc 计算重试间隔的函数类型
type BackoffFunc func(attempt int) time.Duration

// Option 是重试选项的函数类型
type Option func(*Options)

// Options 包含重试的配置选项
type Options struct {
	// MaxAttempts 最大重试次数，默认为 3
	MaxAttempts int
	// Backoff 重试间隔计算函数
	Backoff BackoffFunc
	// IsRetryable 判断错误是否可重试的函数
	IsRetryable IsRetryableFunc
	// OnRetry 每次重试前调用的函数
	OnRetry func(attempt int, err error)
}

// defaultOptions 返回默认选项
func defaultOptions() *Options {
	return &Options{
		MaxAttempts: 3,
		Backoff:     ConstantBackoff(1 * time.Second),
		IsRetryable: func(err error) bool { return err != nil },
		OnRetry:     func(attempt int, err error) {},
	}
}

// WithMaxAttempts 设置最大重试次数
func WithMaxAttempts(attempts int) Option {
	return func(o *Options) {
		if attempts > 0 {
			o.MaxAttempts = attempts
		}
	}
}

// WithBackoff 设置重试间隔计算函数
func WithBackoff(backoff BackoffFunc) Option {
	return func(o *Options) {
		o.Backoff = backoff
	}
}

// WithIsRetryable 设置判断错误是否可重试的函数
func WithIsRetryable(isRetryable IsRetryableFunc) Option {
	return func(o *Options) {
		o.IsRetryable = isRetryable
	}
}

// WithOnRetry 设置每次重试前调用的函数
func WithOnRetry(onRetry func(attempt int, err error)) Option {
	return func(o *Options) {
		o.OnRetry = onRetry
	}
}

// Do 执行带重试的函数
func Do(fn RetryableFunc, opts ...Option) error {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	var err error
	for attempt := 0; attempt < options.MaxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}

		if !options.IsRetryable(err) {
			return err
		}

		if attempt+1 < options.MaxAttempts {
			options.OnRetry(attempt+1, err)
			time.Sleep(options.Backoff(attempt))
		}
	}

	return errors.Join(ErrMaxAttemptsReached, err)
}

// DoWithContext 执行带上下文的重试函数
func DoWithContext(ctx context.Context, fn RetryableFuncWithContext, opts ...Option) error {
	options := defaultOptions()
	for _, opt := range opts {
		opt(options)
	}

	var err error
	for attempt := 0; attempt < options.MaxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			switch ctx.Err() {
			case context.Canceled:
				return errors.Join(ErrContextCanceled, err)
			case context.DeadlineExceeded:
				return errors.Join(ErrContextDeadlineExceeded, err)
			default:
				return ctx.Err()
			}
		default:
			err = fn(ctx)
			if err == nil {
				return nil
			}

			if !options.IsRetryable(err) {
				return err
			}

			if attempt+1 < options.MaxAttempts {
				options.OnRetry(attempt+1, err)

				backoffDuration := options.Backoff(attempt)
				timer := time.NewTimer(backoffDuration)
				select {
				case <-ctx.Done():
					timer.Stop()
					switch ctx.Err() {
					case context.Canceled:
						return errors.Join(ErrContextCanceled, err)
					case context.DeadlineExceeded:
						return errors.Join(ErrContextDeadlineExceeded, err)
					default:
						return ctx.Err()
					}
				case <-timer.C:
					// 继续下一次重试
				}
			}
		}
	}

	return errors.Join(ErrMaxAttemptsReached, err)
}
