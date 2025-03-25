package main

import (
	"fmt"
	"log"
	"time"

	"github.com/qishenonly/retry"
)

func main() {
	// 基本用法示例
	err := retry.Do(
		func() error {
			fmt.Println("尝试执行操作...")
			// 模拟失败
			return fmt.Errorf("操作失败")
		},
		retry.WithMaxAttempts(3),
		retry.WithBackoff(retry.ConstantBackoff(1*time.Second)),
		retry.WithOnRetry(func(attempt int, err error) {
			fmt.Printf("第 %d 次重试，错误: %v\n", attempt, err)
		}),
	)

	if err != nil {
		log.Fatalf("最终失败: %v", err)
	}
}
