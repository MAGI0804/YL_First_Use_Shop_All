package method

import (
	"Member_shop/db"
	"fmt"
)

// 查询记录是否存在
func SearchExistence(TableName string, Query string, selectField interface{}) bool {
	// 1. 校验入参（避免空表名/空字段名导致SQL错误）
	if TableName == "" || Query == "" {
		return false
	}
	var count int64

	// 3. 构建查询：只统计匹配记录数，Limit(1)提升性能
	err := db.DB.Table(TableName).
		Where(fmt.Sprintf("%s = ?", Query), selectField). // 正确拼接字段=值条件
		Limit(1).                                         // 只查1条，减少数据库开销
		Count(&count).                                    // 统计匹配记录数（所有GORM版本都支持）
		Error

	// 4. 逻辑判断：有错误返回false；计数>0则记录存在
	if err != nil {
		// 可选：添加日志排查问题，如 log.Printf("查询记录失败：%v", err)
		return false
	}
	// count>0 表示存在匹配记录，否则不存在
	return count > 0
}

// 查询指定信息
func GetField(TableName string, whereField string, whereValue string, selectField string) (string, error) {
	// ========== 关键修复1：初始化可反射的结果容器 ==========
	// 替换原空接口，改用 map[string]interface{} 并初始化（GORM 能正确反射）
	result := make(map[string]interface{}) // 必须初始化，不能是 nil

	// ========== 关键修复2：安全构建 WHERE 条件（避免 SQL 注入） ==========
	// 替换 fmt.Sprintf 拼接，改用 GORM 原生的 map 条件（防注入）
	condition := map[string]interface{}{
		whereField: whereValue,
	}

	// ========== 关键修复3：规范执行查询（获取 tx 实例，而非直接用全局 DB） ==========
	// 链式调用后用 tx 接收，保证 RowsAffected 是当前查询的结果（避免全局值被覆盖）
	tx := db.DB.Table(TableName).
		Select(selectField). // 仅查询目标字段
		Where(condition).    // 安全的条件查询
		Limit(1).            // 限制单条结果，提升性能
		Find(&result)        // 传入初始化的 map 指针

	// 优先处理数据库查询错误
	if tx.Error != nil {
		return "", fmt.Errorf("数据库查询失败：%w", tx.Error)
	}

	// ========== 关键修复4：正确判断是否有查询结果 ==========
	// 从 tx 实例获取 RowsAffected，而非全局 db.DB（线程安全）
	if tx.RowsAffected == 0 {
		return "", nil // 无匹配记录，返回空字符串+nil
	}

	// ========== 关键修复5：从 map 中提取目标字段值 ==========
	// 原逻辑错误：直接断言空接口类型，现在从初始化的 map 中取值
	value, ok := result[selectField]
	if !ok {
		return "", fmt.Errorf("字段 %s 在表 %s 中不存在", selectField, TableName)
	}

	// ========== 类型转换：兼容各种字段类型转为字符串 ==========
	switch v := value.(type) {
	case string:
		return v, nil
	case int, int8, int16, int32, int64:
		return fmt.Sprintf("%d", v), nil
	case uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32, float64:
		return fmt.Sprintf("%f", v), nil
	case bool:
		return fmt.Sprintf("%t", v), nil
	case nil:
		return "", nil // 字段值为 NULL，返回空字符串
	default:
		// 兜底：兼容时间、字节数组等其他类型
		return fmt.Sprintf("%v", v), nil
	}
}
