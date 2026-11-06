package store

import (
	"database/sql"
	"fmt"
	"time"

	"aipic_go/internal/model"
	"aipic_go/internal/utils"
)

// CreateUser 创建用户
func CreateUser(user *model.User) error {
	_, err := db.Exec(`
		INSERT INTO users (user_id, random_id, avatar, is_vip, vip_level, vip_expire_at, credits, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		user.UserID, user.RandomID, user.Avatar, user.IsVip, user.VipLevel,
		user.VipExpireAt, user.Credits, time.Now(), time.Now(),
	)
	return err
}

// GetUserByUserID 通过 UserID 查询用户
func GetUserByUserID(userID string) (*model.User, error) {
	user := &model.User{}
	var vipExpireAt sql.NullString
	err := db.QueryRow(`
		SELECT user_id, random_id, avatar, is_vip, vip_level, vip_expire_at, credits
		FROM users WHERE user_id = ?
	`, userID).Scan(
		&user.UserID, &user.RandomID, &user.Avatar, &user.IsVip,
		&user.VipLevel, &vipExpireAt, &user.Credits,
	)
	if err != nil {
		return nil, err
	}
	if vipExpireAt.Valid {
		user.VipExpireAt = &vipExpireAt.String
	} else {
		user.VipExpireAt = nil
	}
	return user, nil
}

// UpdateUserAvatar 更新用户头像
func UpdateUserAvatar(userID string, avatarURL string) (*model.User, error) {
	_, err := db.Exec(`UPDATE users SET avatar = ?, updated_at = ? WHERE user_id = ?`, avatarURL, time.Now(), userID)
	if err != nil {
		return nil, err
	}
	return GetUserByUserID(userID)
}

// DeductUserCredits 扣除用户积分（事务安全，积分不足返回错误）
func DeductUserCredits(userID string, amount int64) (int64, error) {
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	var credits int64
	err = tx.QueryRow(`SELECT credits FROM users WHERE user_id = ?`, userID).Scan(&credits)
	if err != nil {
		return 0, err
	}

	if credits < amount {
		return 0, fmt.Errorf("积分不足，当前剩余 %d，需要 %d", credits, amount)
	}

	newCredits := credits - amount
	_, err = tx.Exec(`UPDATE users SET credits = ?, updated_at = ? WHERE user_id = ?`, newCredits, time.Now(), userID)
	if err != nil {
		return 0, err
	}

	return newCredits, tx.Commit()
}

// AddUserCredits 增加用户积分
func AddUserCredits(userID string, amount int64) (int64, error) {
	_, err := db.Exec(`UPDATE users SET credits = credits + ?, updated_at = ? WHERE user_id = ?`, amount, time.Now(), userID)
	if err != nil {
		return 0, err
	}
	user, err := GetUserByUserID(userID)
	if err != nil {
		return 0, err
	}
	return user.Credits, nil
}

// CreateGenerateTask 创建生成任务
func CreateGenerateTask(task *model.GenerateTask) error {
	_, err := db.Exec(`
		INSERT INTO generate_tasks (task_id, user_id, prompt, ratio, resolution, model, custom_width, custom_height, reference_image, status, progress, credits_cost, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		task.TaskID, task.UserID, task.Prompt, task.Ratio, task.Resolution, task.Model,
		task.CustomWidth, task.CustomHeight, task.ReferenceImage,
		task.Status, task.Progress, task.CreditsCost,
		time.Now(), time.Now(),
	)
	return err
}

// GetGenerateTaskByTaskID 通过 TaskID 查询任务
func GetGenerateTaskByTaskID(taskID string) (*model.GenerateTask, error) {
	task := &model.GenerateTask{}
	err := db.QueryRow(`
		SELECT task_id, user_id, prompt, ratio, resolution, model, custom_width, custom_height,
			   reference_image, status, progress, result_url, error_message, credits_cost, created_at
		FROM generate_tasks WHERE task_id = ?
	`, taskID).Scan(
		&task.TaskID, &task.UserID, &task.Prompt, &task.Ratio, &task.Resolution, &task.Model,
		&task.CustomWidth, &task.CustomHeight, &task.ReferenceImage,
		&task.Status, &task.Progress, &task.ResultURL, &task.ErrorMessage, &task.CreditsCost,
		&task.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// UpdateGenerateTaskStatus 更新任务状态
func UpdateGenerateTaskStatus(taskID string, status string, progress int, resultURL string, errorMessage string) error {
	_, err := db.Exec(`
		UPDATE generate_tasks SET status = ?, progress = ?, result_url = ?, error_message = ?, updated_at = ?
		WHERE task_id = ?
	`, status, progress, resultURL, errorMessage, time.Now(), taskID)
	return err
}

// DeleteGenerateTask 删除生成任务
func DeleteGenerateTask(taskID string) error {
	_, err := db.Exec(`DELETE FROM generate_tasks WHERE task_id = ?`, taskID)
	return err
}

// CreateGenerateRecord 创建生成记录
func CreateGenerateRecord(record *model.GenerateRecord) error {
	_, err := db.Exec(`
		INSERT INTO generate_records (record_id, user_id, prompt, ratio, resolution, image_url, status, credits_cost, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		record.RecordID, record.UserID, record.Prompt, record.Ratio,
		record.Resolution, record.ImageURL, record.Status, record.CreditsCost,
		time.Now(),
	)
	return err
}

// GetGenerateRecords 分页查询生成记录
func GetGenerateRecords(userID string, page, pageSize int) ([]*model.GenerateRecord, int64, error) {
	// 查询总数
	var total int64
	err := db.QueryRow(`SELECT COUNT(*) FROM generate_records WHERE user_id = ?`, userID).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// 查询列表
	offset := (page - 1) * pageSize
	rows, err := db.Query(`
		SELECT record_id, user_id, prompt, ratio, resolution, image_url, status, credits_cost, created_at
		FROM generate_records WHERE user_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var records []*model.GenerateRecord
	for rows.Next() {
		r := &model.GenerateRecord{}
		err := rows.Scan(&r.RecordID, &r.UserID, &r.Prompt, &r.Ratio,
			&r.Resolution, &r.ImageURL, &r.Status, &r.CreditsCost, &r.CreatedAt)
		if err != nil {
			return nil, 0, err
		}
		records = append(records, r)
	}

	return records, total, nil
}

// DeleteGenerateRecord 删除单条记录（校验所属用户）
func DeleteGenerateRecord(recordID string, userID string) error {
	result, err := db.Exec(`DELETE FROM generate_records WHERE record_id = ? AND user_id = ?`, recordID, userID)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("记录不存在或无权删除")
	}
	return nil
}

// DeleteAllGenerateRecords 删除用户所有记录
func DeleteAllGenerateRecords(userID string) error {
	_, err := db.Exec(`DELETE FROM generate_records WHERE user_id = ?`, userID)
	return err
}

// GetTodayGenerateCount 获取今日生成次数
func GetTodayGenerateCount(userID string) (int64, error) {
	today := time.Now().Format("2006-01-02")
	var count int64
	err := db.QueryRow(`
		SELECT COUNT(*) FROM generate_records
		WHERE user_id = ? AND date(created_at) = ?
	`, userID, today).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// CreatePaymentOrder 创建支付订单
func CreatePaymentOrder(order *model.PaymentOrder) error {
	_, err := db.Exec(`
		INSERT INTO payment_orders (order_id, user_id, order_type, amount, credits, status, plan_type, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		order.OrderID, order.UserID, order.OrderType, order.Amount,
		order.Credits, order.Status, order.PlanType,
		time.Now(), time.Now(),
	)
	return err
}

// GetPaymentOrderByOrderID 通过 OrderID 查询订单
func GetPaymentOrderByOrderID(orderID string) (*model.PaymentOrder, error) {
	order := &model.PaymentOrder{}
	err := db.QueryRow(`
		SELECT order_id, user_id, order_type, amount, credits, status, plan_type
		FROM payment_orders WHERE order_id = ?
	`, orderID).Scan(
		&order.OrderID, &order.UserID, &order.OrderType, &order.Amount,
		&order.Credits, &order.Status, &order.PlanType,
	)
	if err != nil {
		return nil, err
	}
	return order, nil
}

// GetPaymentOrderByOrderIDAndUser 通过 OrderID 和 UserID 查询订单（越权校验）
func GetPaymentOrderByOrderIDAndUser(orderID string, userID string) (*model.PaymentOrder, error) {
	order := &model.PaymentOrder{}
	err := db.QueryRow(`
		SELECT order_id, user_id, order_type, amount, credits, status, plan_type
		FROM payment_orders WHERE order_id = ? AND user_id = ?
	`, orderID, userID).Scan(
		&order.OrderID, &order.UserID, &order.OrderType, &order.Amount,
		&order.Credits, &order.Status, &order.PlanType,
	)
	if err != nil {
		return nil, err
	}
	return order, nil
}

// UpdatePaymentOrderStatus 更新订单状态
func UpdatePaymentOrderStatus(orderID string, status string) error {
	_, err := db.Exec(`UPDATE payment_orders SET status = ?, updated_at = ? WHERE order_id = ?`, status, time.Now(), orderID)
	return err
}

// EnsureTestUser 确保测试用户存在（开发用）
func EnsureTestUser() {
	userID := "u_001"
	_, err := GetUserByUserID(userID)
	if err == sql.ErrNoRows {
		testUser := &model.User{
			UserID:      userID,
			RandomID:    utils.GenerateRandomID(),
			Avatar:      "",
			IsVip:       true,
			VipLevel:    "vip",
			VipExpireAt: strPtr("2027-01-01T00:00:00.000Z"),
			Credits:     1000,
		}
		_ = CreateUser(testUser)
	}
}

// ExecUpdateVIP 更新用户VIP状态
func ExecUpdateVIP(userID string, isVip bool, vipLevel string, vipExpireAt string) error {
	_, err := db.Exec(`
		UPDATE users SET is_vip = ?, vip_level = ?, vip_expire_at = ?, updated_at = ?
		WHERE user_id = ?
	`, isVip, vipLevel, vipExpireAt, time.Now(), userID)
	return err
}

// GetUserVipExpireAt 获取用户当前 VIP 到期时间
func GetUserVipExpireAt(userID string) (*time.Time, error) {
	var vipExpireAt sql.NullString
	err := db.QueryRow(`SELECT vip_expire_at FROM users WHERE user_id = ?`, userID).Scan(&vipExpireAt)
	if err != nil {
		return nil, err
	}
	if !vipExpireAt.Valid {
		return nil, nil
	}
	t, err := time.Parse("2006-01-02T15:04:05.000Z", vipExpireAt.String)
	if err != nil {
		// 尝试其他格式
		t, err = time.Parse(time.RFC3339, vipExpireAt.String)
		if err != nil {
			return nil, err
		}
	}
	return &t, nil
}

// ConfirmPaymentInTransaction 事务内确认支付（订单状态更新 + 积分增加 + VIP 更新，原子操作）
func ConfirmPaymentInTransaction(orderID string, vipPlanType string, vipExpireDate string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// 1. 查询订单并加锁（防止并发重复处理）
	var orderUserID string
	var orderType string
	var orderStatus string
	var orderCredits int64
	err = tx.QueryRow(`
		SELECT user_id, order_type, status, credits FROM payment_orders WHERE order_id = ?
	`, orderID).Scan(&orderUserID, &orderType, &orderStatus, &orderCredits)
	if err != nil {
		return fmt.Errorf("订单不存在")
	}

	// 2. 校验订单状态
	if orderStatus != "pending" {
		return fmt.Errorf("订单状态不允许确认: %s", orderStatus)
	}

	// 3. 更新订单状态为 paid
	_, err = tx.Exec(`UPDATE payment_orders SET status = 'paid', updated_at = ? WHERE order_id = ?`, time.Now(), orderID)
	if err != nil {
		return err
	}

	// 4. 增加用户积分
	_, err = tx.Exec(`UPDATE users SET credits = credits + ?, updated_at = ? WHERE user_id = ?`, orderCredits, time.Now(), orderUserID)
	if err != nil {
		return err
	}

	// 5. 如果是 VIP 订单，更新 VIP 状态
	if orderType == "vip" && vipPlanType != "" && vipExpireDate != "" {
		_, err = tx.Exec(`
			UPDATE users SET is_vip = ?, vip_level = ?, vip_expire_at = ?, updated_at = ?
			WHERE user_id = ?
		`, true, "vip", vipExpireDate, time.Now(), orderUserID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func strPtr(s string) *string {
	return &s
}
