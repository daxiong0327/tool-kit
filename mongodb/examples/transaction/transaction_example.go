package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/daxiong0327/tool-kit/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Account 账户结构体
type Account struct {
	ID      primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name    string             `bson:"name" json:"name"`
	Balance float64            `bson:"balance" json:"balance"`
}

// Transfer 转账记录结构体
type Transfer struct {
	ID        primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	From      primitive.ObjectID `bson:"from" json:"from"`
	To        primitive.ObjectID `bson:"to" json:"to"`
	Amount    float64            `bson:"amount" json:"amount"`
	CreatedAt time.Time          `bson:"created_at" json:"created_at"`
}

func main() {
	fmt.Println("=== MongoDB 事务操作示例 ===")

	// 创建MongoDB客户端
	config := &mongodb.Config{
		URI:            "mongodb://localhost:27017",
		Database:       "bank_db",
		ConnectTimeout: 10 * time.Second,
		SocketTimeout:  5 * time.Second,
		MaxPoolSize:    100,
		RetryWrites:    true,
		RetryReads:     true,
	}

	client, err := mongodb.New(config)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Close(context.Background())

	ctx := context.Background()

	// 测试连接
	err = client.Ping(ctx)
	if err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}
	fmt.Println("✅ 成功连接到MongoDB")

	// 清理测试数据
	client.DeleteMany(ctx, "accounts", bson.M{})
	client.DeleteMany(ctx, "transfers", bson.M{})
	fmt.Println("🧹 清理测试数据")

	// 示例1：基本事务操作
	fmt.Println("\n1. 基本事务操作 - 创建账户:")
	err = createAccountsWithTransaction(client, ctx)
	if err != nil {
		log.Fatalf("Failed to create accounts: %v", err)
	}

	// 示例2：转账事务
	fmt.Println("\n2. 转账事务 - 成功转账:")
	err = transferMoneyWithTransaction(client, ctx, "张三", "李四", 100.0)
	if err != nil {
		log.Fatalf("Failed to transfer money: %v", err)
	}

	// 示例3：转账事务 - 余额不足
	fmt.Println("\n3. 转账事务 - 余额不足:")
	err = transferMoneyWithTransaction(client, ctx, "张三", "李四", 1000.0)
	if err != nil {
		fmt.Printf("❌ 转账失败 (预期): %v\n", err)
	}

	// 示例4：批量转账事务
	fmt.Println("\n4. 批量转账事务:")
	err = batchTransferWithTransaction(client, ctx)
	if err != nil {
		log.Fatalf("Failed to batch transfer: %v", err)
	}

	// 示例5：事务选项配置
	fmt.Println("\n5. 带选项的事务操作:")
	err = transactionWithOptions(client, ctx)
	if err != nil {
		log.Fatalf("Failed to execute transaction with options: %v", err)
	}

	// 最终账户状态
	fmt.Println("\n📊 最终账户状态:")
	showAccountBalances(client, ctx)
}

// createAccountsWithTransaction 使用事务创建账户
func createAccountsWithTransaction(client *mongodb.Client, ctx context.Context) error {
	return client.WithTransaction(ctx, func(sc mongo.SessionContext) error {
		// 创建张三的账户
		zhangSan := Account{
			Name:    "张三",
			Balance: 500.0,
		}
		_, err := client.InsertOne(sc, "accounts", zhangSan)
		if err != nil {
			return fmt.Errorf("failed to create 张三 account: %w", err)
		}

		// 创建李四的账户
		liSi := Account{
			Name:    "李四",
			Balance: 300.0,
		}
		_, err = client.InsertOne(sc, "accounts", liSi)
		if err != nil {
			return fmt.Errorf("failed to create 李四 account: %w", err)
		}

		// 创建王五的账户
		wangWu := Account{
			Name:    "王五",
			Balance: 200.0,
		}
		_, err = client.InsertOne(sc, "accounts", wangWu)
		if err != nil {
			return fmt.Errorf("failed to create 王五 account: %w", err)
		}

		fmt.Println("✅ 成功创建3个账户")
		return nil
	})
}

// transferMoneyWithTransaction 使用事务进行转账
func transferMoneyWithTransaction(client *mongodb.Client, ctx context.Context, fromName, toName string, amount float64) error {
	return client.WithTransaction(ctx, func(sc mongo.SessionContext) error {
		// 查找转出账户
		var fromAccount Account
		err := client.FindOne(sc, "accounts", bson.M{"name": fromName}, &fromAccount)
		if err != nil {
			return fmt.Errorf("failed to find from account: %w", err)
		}

		// 查找转入账户
		var toAccount Account
		err = client.FindOne(sc, "accounts", bson.M{"name": toName}, &toAccount)
		if err != nil {
			return fmt.Errorf("failed to find to account: %w", err)
		}

		// 检查余额
		if fromAccount.Balance < amount {
			return fmt.Errorf("insufficient balance: %s has %.2f, trying to transfer %.2f", fromName, fromAccount.Balance, amount)
		}

		// 更新转出账户余额
		fromFilter := bson.M{"_id": fromAccount.ID}
		fromUpdate := bson.M{"$inc": bson.M{"balance": -amount}}
		_, err = client.UpdateOne(sc, "accounts", fromFilter, fromUpdate)
		if err != nil {
			return fmt.Errorf("failed to update from account: %w", err)
		}

		// 更新转入账户余额
		toFilter := bson.M{"_id": toAccount.ID}
		toUpdate := bson.M{"$inc": bson.M{"balance": amount}}
		_, err = client.UpdateOne(sc, "accounts", toFilter, toUpdate)
		if err != nil {
			return fmt.Errorf("failed to update to account: %w", err)
		}

		// 记录转账
		transfer := Transfer{
			From:      fromAccount.ID,
			To:        toAccount.ID,
			Amount:    amount,
			CreatedAt: time.Now(),
		}
		_, err = client.InsertOne(sc, "transfers", transfer)
		if err != nil {
			return fmt.Errorf("failed to record transfer: %w", err)
		}

		fmt.Printf("✅ 成功转账 %.2f 从 %s 到 %s\n", amount, fromName, toName)
		return nil
	})
}

// batchTransferWithTransaction 批量转账事务
func batchTransferWithTransaction(client *mongodb.Client, ctx context.Context) error {
	return client.WithTransaction(ctx, func(sc mongo.SessionContext) error {
		// 批量转账操作
		transfers := []struct {
			from   string
			to     string
			amount float64
		}{
			{"张三", "李四", 50.0},
			{"李四", "王五", 30.0},
			{"王五", "张三", 20.0},
		}

		for _, t := range transfers {
			// 查找转出账户
			var fromAccount Account
			err := client.FindOne(sc, "accounts", bson.M{"name": t.from}, &fromAccount)
			if err != nil {
				return fmt.Errorf("failed to find from account %s: %w", t.from, err)
			}

			// 查找转入账户
			var toAccount Account
			err = client.FindOne(sc, "accounts", bson.M{"name": t.to}, &toAccount)
			if err != nil {
				return fmt.Errorf("failed to find to account %s: %w", t.to, err)
			}

			// 检查余额
			if fromAccount.Balance < t.amount {
				return fmt.Errorf("insufficient balance for %s: has %.2f, trying to transfer %.2f", t.from, fromAccount.Balance, t.amount)
			}

			// 更新转出账户余额
			fromFilter := bson.M{"_id": fromAccount.ID}
			fromUpdate := bson.M{"$inc": bson.M{"balance": -t.amount}}
			_, err = client.UpdateOne(sc, "accounts", fromFilter, fromUpdate)
			if err != nil {
				return fmt.Errorf("failed to update from account %s: %w", t.from, err)
			}

			// 更新转入账户余额
			toFilter := bson.M{"_id": toAccount.ID}
			toUpdate := bson.M{"$inc": bson.M{"balance": t.amount}}
			_, err = client.UpdateOne(sc, "accounts", toFilter, toUpdate)
			if err != nil {
				return fmt.Errorf("failed to update to account %s: %w", t.to, err)
			}

			// 记录转账
			transfer := Transfer{
				From:      fromAccount.ID,
				To:        toAccount.ID,
				Amount:    t.amount,
				CreatedAt: time.Now(),
			}
			_, err = client.InsertOne(sc, "transfers", transfer)
			if err != nil {
				return fmt.Errorf("failed to record transfer: %w", err)
			}

			fmt.Printf("  ✅ %s -> %s: %.2f\n", t.from, t.to, t.amount)
		}

		fmt.Println("✅ 批量转账完成")
		return nil
	})
}

// transactionWithOptions 带选项的事务操作
func transactionWithOptions(client *mongodb.Client, ctx context.Context) error {
	// 配置事务选项
	transactionOpts := mongodb.NewTransactionOptions().
		SetMaxCommitTime(5000) // 5秒超时

	return client.WithTransaction(ctx, func(sc mongo.SessionContext) error {
		// 在事务中执行操作
		var accounts []Account
		err := client.Find(sc, "accounts", bson.M{}, &accounts)
		if err != nil {
			return fmt.Errorf("failed to find accounts: %w", err)
		}

		// 为所有账户增加1%的利息
		for _, account := range accounts {
			interest := account.Balance * 0.01
			filter := bson.M{"_id": account.ID}
			update := bson.M{"$inc": bson.M{"balance": interest}}

			_, err = client.UpdateOne(sc, "accounts", filter, update)
			if err != nil {
				return fmt.Errorf("failed to update account %s: %w", account.Name, err)
			}

			fmt.Printf("  ✅ %s 获得利息: %.2f\n", account.Name, interest)
		}

		fmt.Println("✅ 利息计算完成")
		return nil
	}, transactionOpts)
}

// showAccountBalances 显示账户余额
func showAccountBalances(client *mongodb.Client, ctx context.Context) {
	var accounts []Account
	err := client.Find(ctx, "accounts", bson.M{}, &accounts)
	if err != nil {
		log.Printf("Failed to find accounts: %v", err)
		return
	}

	for _, account := range accounts {
		fmt.Printf("  %s: %.2f\n", account.Name, account.Balance)
	}

	// 显示转账记录
	var transfers []Transfer
	err = client.Find(ctx, "transfers", bson.M{}, &transfers)
	if err != nil {
		log.Printf("Failed to find transfers: %v", err)
		return
	}

	fmt.Printf("\n📋 转账记录 (%d 条):\n", len(transfers))
	for _, transfer := range transfers {
		// 查找账户名称
		var fromAccount, toAccount Account
		client.FindOne(ctx, "accounts", bson.M{"_id": transfer.From}, &fromAccount)
		client.FindOne(ctx, "accounts", bson.M{"_id": transfer.To}, &toAccount)

		fmt.Printf("  %s -> %s: %.2f (%s)\n",
			fromAccount.Name, toAccount.Name, transfer.Amount,
			transfer.CreatedAt.Format("2006-01-02 15:04:05"))
	}
}
