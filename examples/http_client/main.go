package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/qishenonly/retry"
)

func main() {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 带上下文的HTTP请求重试示例
	err := retry.DoWithContext(
		ctx,
		func(ctx context.Context) error {
			req, err := http.NewRequestWithContext(ctx, "GET", "https://example.com", nil)
			if err != nil {
				return err
			}

			resp, err := client.Do(req)
			if err != nil {
				return err
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 500 {
				return retry.NewHTTPError(resp.StatusCode, "服务器错误")
			}

			fmt.Println("请求成功，状态码:", resp.StatusCode)
			return nil
		},
		retry.WithMaxAttempts(5),
		retry.WithBackoff(retry.ExponentialBackoffWithJitter(100*time.Millisecond, 5*time.Second, 0.2)),
		retry.WithIsRetryable(retry.IsRetryableHTTPError),
		retry.WithOnRetry(func(attempt int, err error) {
			fmt.Printf("HTTP请求重试 %d，错误: %v\n", attempt, err)
		}),
	)

	if err != nil {
		log.Fatalf("HTTP请求最终失败: %v", err)
	}
}
