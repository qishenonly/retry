# Retry 库

一个高可用性的 Go 重试库，支持多种重试策略和灵活的配置选项。

## 特性

- 支持多种重试策略（固定间隔、指数退避、线性增长等）
- 支持重试次数限制
- 支持超时控制
- 支持自定义重试条件
- 提供简洁易用的 API
- 良好的错误处理和日志记录

## 安装

```bash
go get github.com/qishenonly/retry
```

## 使用示例

### 基本用法

```go
package main

import (
	"fmt"
	"github.com/qishenonly/retry"
	"time"
)

func main() {
	err := retry.Do(
		func() error {
			// 你的业务逻辑
			return fmt.Errorf("some error")
		},
		retry.WithMaxAttempts(3),
		retry.WithBackoff(retry.ConstantBackoff(1*time.Second)),
	)
	
	if err != nil {
		fmt.Printf("Failed after retries: %v\n", err)
	}
}
```

### 带上下文的重试

```go
package main

import (
	"context"
	"fmt"
	"github.com/qishenonly/retry"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	err := retry.DoWithContext(
		ctx,
		func(ctx context.Context) error {
			// 你的业务逻辑
			return fmt.Errorf("some error")
		},
		retry.WithMaxAttempts(5),
		retry.WithBackoff(retry.ExponentialBackoff(100*time.Millisecond, 5*time.Second)),
		retry.WithOnRetry(func(attempt int, err error) {
			fmt.Printf("Retry %d after error: %v\n", attempt, err)
		}),
	)
	
	if err != nil {
		fmt.Printf("Failed after retries: %v\n", err)
	}
}
```

### 自定义重试条件

```go
package main

import (
	"fmt"
	"github.com/qishenonly/retry"
	"time"
)

func main() {
	err := retry.Do(
		func() error {
			// 你的业务逻辑
			return &retry.HTTPError{StatusCode: 503, Message: "Service Unavailable"}
		},
		retry.WithMaxAttempts(3),
		retry.WithBackoff(retry.ExponentialBackoffWithJitter(100*time.Millisecond, 5*time.Second, 0.2)),
		retry.WithIsRetryable(retry.IsRetryableHTTPError),
	)
	
	if err != nil {
		fmt.Printf("Failed after retries: %v\n", err)
	}
}
```

## 重试策略

### 固定间隔 (ConstantBackoff)

每次重试使用相同的时间间隔。

```go
retry.WithBackoff(retry.ConstantBackoff(1*time.Second))
```

### 指数退避 (ExponentialBackoff)

每次重试的间隔呈指数增长，公式为：`interval * 2^attempt`。

```go
retry.WithBackoff(retry.ExponentialBackoff(100*time.Millisecond, 5*time.Second))
```

### 带抖动的指数退避 (ExponentialBackoffWithJitter)

在指数退避的基础上增加随机抖动，避免多个客户端同时重试导致的"惊群效应"。

```go
retry.WithBackoff(retry.ExponentialBackoffWithJitter(100*time.Millisecond, 5*time.Second, 0.2))
```

### 线性增长 (LinearBackoff)

每次重试的间隔线性增长，公式为：`interval * (attempt + 1)`。

```go
retry.WithBackoff(retry.LinearBackoff(100*time.Millisecond, 5*time.Second))
```

## 错误处理

库提供了几种预定义的错误类型和判断函数：

- `ErrMaxAttemptsReached`: 达到最大重试次数
- `ErrContextCanceled`: 上下文被取消
- `ErrContextDeadlineExceeded`: 上下文超时
- `IsNetworkError`: 判断是否为网络错误
- `IsHTTPRetryable`: 判断HTTP状态码是否可重试
- `IsRetryableHTTPError`: 判断HTTP错误是否可重试

## 许可证

MIT