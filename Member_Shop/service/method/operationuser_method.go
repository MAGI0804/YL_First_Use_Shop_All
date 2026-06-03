package method

import (
	"Member_shop/config"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
	"Member_shop/utils"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AddServiceUser 添加客服用户
// 在Customer_service_user表中创建新的客服账号
// 包含手机号唯一性校验、密码哈希处理、重试机制保证高可用性
func AddServiceUser(req requestbody.AddServiceUserRequest) (string, error) {
	maxRetries := 3
	retryCount := 0

	for retryCount < maxRetries {
		tryAgain := false
		retryCount++

		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM Customer_service_user WHERE mobile = ?)"
		err := db.DB.Raw(query, req.Mobile).Scan(&exists).Error
		if err != nil {
			log.Printf("数据库检查手机号是否存在失败: %v", err)
			tryAgain = true
		} else if exists {
			return "", fmt.Errorf("手机号已存在")
		}

		if tryAgain {
			continue
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码哈希失败: %v", err)
			tryAgain = true
		}

		if tryAgain {
			continue
		}

		user := models.DjangoCustomerServiceUser{
			Nickname: req.Nickname,
			Mobile:   req.Mobile,
			Password: string(hashedPassword),
		}

		sqlDB, err := db.DB.DB()
		if err != nil {
			log.Printf("获取数据库连接失败: %v", err)
			tryAgain = true
		} else {
			sqlTx, err := sqlDB.Begin()
			if err != nil {
				log.Printf("开始事务失败: %v", err)
				tryAgain = true
			} else {
				txStarted := true
				defer func() {
					if txStarted {
						sqlTx.Rollback()
					}
				}()

				if err := user.BeforeSave(sqlTx); err != nil {
					log.Printf("生成user_id失败: %v", err)
					tryAgain = true
				} else {
					_, err = sqlTx.Exec("INSERT INTO Customer_service_user (user_id, nickname, mobile, password) VALUES (?, ?, ?, ?)",
						user.UserID, user.Nickname, user.Mobile, user.Password)
					if err != nil {
						log.Printf("插入用户失败: %v", err)
						if strings.Contains(err.Error(), "duplicate key") {
							return "", fmt.Errorf("手机号或昵称已存在")
						}
						tryAgain = true
					} else {
						if err := sqlTx.Commit(); err != nil {
							log.Printf("提交事务失败: %v", err)
							tryAgain = true
						} else {
							txStarted = false
							log.Printf("客服用户创建成功: %s", user.UserID)
							return user.UserID, nil
						}
					}
				}
			}
		}

		if tryAgain && retryCount < maxRetries {
			shiftValue := 1 << uint(retryCount-1)
			sleepTime := 0.5 * float64(shiftValue)
			log.Printf("%.2f秒后重试，尝试 %d/%d", sleepTime, retryCount, maxRetries)
			time.Sleep(time.Duration(sleepTime * float64(time.Second)))
		}
	}

	return "", fmt.Errorf("服务器错误，重试多次后失败")
}

// AddOperationUser 添加运营用户
// 在Operation_user表中创建新的运营账号
// 支持不同级别的运营用户（level字段），用于权限控制
func AddOperationUser(req requestbody.AddOperationUserRequest) (string, error) {
	var level int
	switch v := req.Level.(type) {
	case int:
		level = v
	case float64:
		level = int(v)
	case string:
		if _, err := fmt.Sscanf(v, "%d", &level); err != nil {
			return "", fmt.Errorf("level格式无效，必须是数字")
		}
	default:
		return "", fmt.Errorf("level类型无效")
	}

	if level == 0 {
		return "", fmt.Errorf("level不能为0")
	}

	maxRetries := 3
	retryCount := 0

	for retryCount < maxRetries {
		tryAgain := false
		retryCount++

		var exists bool
		query := "SELECT EXISTS(SELECT 1 FROM Operation_user WHERE mobile = ?)"
		err := db.DB.Raw(query, req.Mobile).Scan(&exists).Error
		if err != nil {
			log.Printf("数据库检查手机号是否存在失败: %v", err)
			tryAgain = true
		} else if exists {
			return "", fmt.Errorf("手机号已存在")
		}

		if tryAgain {
			continue
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			log.Printf("密码哈希失败: %v", err)
			tryAgain = true
		}

		if tryAgain {
			continue
		}

		user := models.DjangoOperationUser{
			Nickname: req.Nickname,
			Mobile:   req.Mobile,
			Password: string(hashedPassword),
			Level:    level,
		}

		sqlDB, err := db.DB.DB()
		if err != nil {
			log.Printf("获取数据库连接失败: %v", err)
			tryAgain = true
		} else {
			sqlTx, err := sqlDB.Begin()
			if err != nil {
				log.Printf("开始事务失败: %v", err)
				tryAgain = true
			} else {
				txStarted := true
				defer func() {
					if txStarted {
						sqlTx.Rollback()
					}
				}()

				if err := user.BeforeSave(sqlTx); err != nil {
					log.Printf("生成user_id失败: %v", err)
					tryAgain = true
				} else {
					_, err = sqlTx.Exec("INSERT INTO Operation_user (user_id, nickname, mobile, password, level) VALUES (?, ?, ?, ?, ?)",
						user.UserID, user.Nickname, user.Mobile, user.Password, user.Level)
					if err != nil {
						log.Printf("插入用户失败: %v", err)
						if strings.Contains(err.Error(), "duplicate key") {
							return "", fmt.Errorf("手机号或昵称已存在")
						}
						tryAgain = true
					} else {
						if err := sqlTx.Commit(); err != nil {
							log.Printf("提交事务失败: %v", err)
							tryAgain = true
						} else {
							txStarted = false
							log.Printf("运营用户创建成功: %s", user.UserID)
							return user.UserID, nil
						}
					}
				}
			}
		}

		if tryAgain && retryCount < maxRetries {
			shiftValue := 1 << uint(retryCount-1)
			sleepTime := 0.5 * float64(shiftValue)
			log.Printf("%.2f秒后重试，尝试 %d/%d", sleepTime, retryCount, maxRetries)
			time.Sleep(time.Duration(sleepTime * float64(time.Second)))
		}
	}

	return "", fmt.Errorf("服务器错误，重试多次后失败")
}

// RegisterBackendUserByPhone 在backend_operation_user中创建员工/运营账号
// 与users_user和member_info分开：
// - users_user存储微信用户
// - member_info存储会员业务数据
// - backend_operation_user存储后台员工登录数据
// 使用手机号验证码方式进行注册验证
func RegisterBackendUserByPhone(req requestbody.BackendRegisterByPhoneRequest) (*models.BackendUser, error) {
	if !IsValidMobile(req.Mobile) {
		return nil, fmt.Errorf("手机号格式无效")
	}
	if len(req.Password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}
	if err := db.VerifyCaptcha(req.Mobile, req.Captcha); err != nil {
		return nil, err
	}

	var backendUser models.BackendUser
	err := db.DB.Where("mobile = ?", req.Mobile).First(&backendUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		isBootstrap, countErr := hasNoBackendUsers()
		if countErr != nil {
			return nil, countErr
		}
		if !isBootstrap {
			return nil, fmt.Errorf("mobile has not been invited by an administrator")
		}
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if backendUser.ID != 0 && backendUser.Status == "disabled" {
		return nil, fmt.Errorf("account is disabled")
	}
	if backendUser.ID != 0 && backendUser.Status == "active" && backendUser.Password != "" {
		return nil, fmt.Errorf("account has already been activated")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	if backendUser.ID == 0 {
		backendUser = models.BackendUser{
			Nickname: "超级管理员",
			Mobile:   req.Mobile,
			Role:     "admin",
			Level:    9,
			Status:   "active",
		}
	}
	if strings.TrimSpace(req.Nickname) != "" {
		backendUser.Nickname = strings.TrimSpace(req.Nickname)
	}
	backendUser.Password = string(hashedPassword)
	backendUser.Status = "active"
	if err := db.DB.Save(&backendUser).Error; err != nil {
		return nil, err
	}
	return &backendUser, nil
}

// VerificationResult 验证结果结构体
// 包含用户ID、昵称、密码（加密后）
// BackendUserSession is returned to the web app after login or token validation.
type BackendUserSession struct {
	ID           uint     `json:"id"`
	OperatorNo   string   `json:"operator_no"`
	Mobile       string   `json:"mobile"`
	Nickname     string   `json:"nickname"`
	Role         string   `json:"role"`
	Level        int      `json:"level"`
	Status       string   `json:"status"`
	Permissions  []string `json:"permissions"`
	Token        string   `json:"token,omitempty"`
	RefreshToken string   `json:"refresh_token,omitempty"`
}

// BuildBackendUserSession converts a model into the stable web auth response.
func BuildBackendUserSession(user *models.BackendUser, token, refreshToken string) BackendUserSession {
	return BackendUserSession{
		ID:           user.ID,
		OperatorNo:   user.OperatorNo,
		Mobile:       user.Mobile,
		Nickname:     user.Nickname,
		Role:         user.Role,
		Level:        user.Level,
		Status:       user.Status,
		Permissions:  BackendUserPermissions(user),
		Token:        token,
		RefreshToken: refreshToken,
	}
}

func BackendUserPermissions(user *models.BackendUser) []string {
	base := []string{"dashboard", "home-manage", "product", "inventory", "order", "after-sales", "reviews", "member", "report"}
	if IsBackendAdmin(user) {
		return append(base, "users")
	}
	return base
}

func IsBackendAdmin(user *models.BackendUser) bool {
	if user == nil {
		return false
	}
	return user.Role == "admin" || user.Level >= 9
}

func AddBackendUserInvite(req requestbody.AddBackendUserInviteRequest) (*models.BackendUser, error) {
	if !IsValidMobile(req.Mobile) {
		return nil, fmt.Errorf("invalid mobile")
	}
	if strings.TrimSpace(req.Nickname) == "" {
		return nil, fmt.Errorf("nickname is required")
	}
	var existing models.BackendUser
	err := db.DB.Where("mobile = ?", req.Mobile).First(&existing).Error
	if err == nil {
		return nil, fmt.Errorf("mobile already exists")
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	role := strings.TrimSpace(req.Role)
	if role == "" {
		role = "operation"
	}
	level := req.Level
	if level <= 0 {
		level = 1
	}
	user := models.BackendUser{
		Nickname: strings.TrimSpace(req.Nickname),
		Mobile:   req.Mobile,
		Password: "",
		Role:     role,
		Level:    level,
		Status:   "pending",
		Remarks:  req.Remarks,
	}
	if err := db.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func CanSendBackendRegisterCaptcha(mobile string) error {
	if !IsValidMobile(mobile) {
		return fmt.Errorf("invalid mobile")
	}
	var user models.BackendUser
	if err := db.DB.Where("mobile = ?", mobile).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			isBootstrap, countErr := hasNoBackendUsers()
			if countErr != nil {
				return countErr
			}
			if isBootstrap {
				return nil
			}
			return fmt.Errorf("mobile has not been invited by an administrator")
		}
		return err
	}
	if user.Status == "disabled" {
		return fmt.Errorf("account is disabled")
	}
	if user.Status == "active" && user.Password != "" {
		return fmt.Errorf("account has already been activated")
	}
	return nil
}

func hasNoBackendUsers() (bool, error) {
	var count int64
	if err := db.DB.Model(&models.BackendUser{}).Count(&count).Error; err != nil {
		return false, err
	}
	return count == 0, nil
}

func BackendLogin(req requestbody.BackendLoginRequest) (*models.BackendUser, string, string, error) {
	if !IsValidMobile(req.Mobile) {
		return nil, "", "", fmt.Errorf("invalid mobile")
	}
	var user models.BackendUser
	if err := db.DB.Where("mobile = ?", req.Mobile).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "", "", fmt.Errorf("account not found")
		}
		return nil, "", "", err
	}
	if user.Status != "active" {
		return nil, "", "", fmt.Errorf("account is not active")
	}
	if user.Password == "" {
		return nil, "", "", fmt.Errorf("account has not been activated")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, "", "", fmt.Errorf("invalid password")
	}
	accessToken, refreshToken, err := utils.GenerateTokens(int(user.ID), config.LoadConfig())
	if err != nil {
		return nil, "", "", err
	}
	return &user, accessToken, refreshToken, nil
}

func QueryBackendUsers(req requestbody.QueryBackendUsersRequest) ([]models.BackendUser, int64, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	query := db.DB.Model(&models.BackendUser{})
	if req.Mobile != "" {
		query = query.Where("mobile LIKE ?", "%"+req.Mobile+"%")
	}
	if req.Status != "" {
		query = query.Where("status = ?", req.Status)
	}
	if req.Role != "" {
		query = query.Where("role = ?", req.Role)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var users []models.BackendUser
	if err := query.Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func UpdateBackendUserStatus(req requestbody.UpdateBackendUserStatusRequest) (*models.BackendUser, error) {
	if req.Status != "pending" && req.Status != "active" && req.Status != "disabled" {
		return nil, fmt.Errorf("invalid status")
	}
	var user models.BackendUser
	if err := db.DB.Where("id = ?", req.ID).First(&user).Error; err != nil {
		return nil, err
	}
	if req.Status == "disabled" && IsBackendAdmin(&user) {
		var activeAdminCount int64
		if err := db.DB.Model(&models.BackendUser{}).
			Where("status = ? AND (role = ? OR level >= ?)", "active", "admin", 9).
			Count(&activeAdminCount).Error; err != nil {
			return nil, err
		}
		if activeAdminCount <= 1 {
			return nil, fmt.Errorf("cannot disable the last active administrator")
		}
	}
	user.Status = req.Status
	if err := db.DB.Save(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

type VerificationResult struct {
	UserID   string
	Nickname string
	Password string
}

// VerificationStatus 验证登录状态
// 根据手机号查询用户信息，用于登录验证
// objectNum区分用户类型：1-运营用户(Operation_user)，2-客服用户(Customer_service_user)
func VerificationStatus(req requestbody.VerificationStatusRequest) (*VerificationResult, error) {
	var tableName string
	if req.ObjectNum == "1" {
		tableName = "Operation_user"
	} else if req.ObjectNum == "2" {
		tableName = "Customer_service_user"
	} else {
		return nil, fmt.Errorf("object_num参数错误")
	}

	maxRetries := 3
	retryCount := 0

	var userID, nickname, password string
	var queryErr error

	for retryCount < maxRetries {
		retryCount++
		log.Printf("开始查询用户(尝试%d/%d): mobile=%s", retryCount, maxRetries, req.Mobile)

		query := fmt.Sprintf("SELECT user_id, nickname, password FROM %s WHERE mobile = ?", tableName)
		if db.DB != nil {
			sqlDB, err := db.DB.DB()
			if err != nil {
				log.Printf("获取数据库连接失败: %v", err)
				queryErr = fmt.Errorf("数据库连接错误")
			} else {
				queryErr = sqlDB.QueryRow(query, req.Mobile).Scan(&userID, &nickname, &password)
			}
		} else {
			queryErr = fmt.Errorf("数据库实例未初始化")
		}

		if queryErr != nil {
			if queryErr == sql.ErrNoRows {
				log.Printf("手机号未注册: %s", req.Mobile)
				return nil, fmt.Errorf("手机号未注册")
			}
			log.Printf("用户查询异常(尝试%d/%d): %v", retryCount, maxRetries, queryErr)
			if retryCount >= maxRetries {
				return nil, fmt.Errorf("用户信息查询失败，请稍后重试")
			}
			shiftValue := 1 << uint(retryCount-1)
			sleepTime := 0.5 * float64(shiftValue)
			log.Printf("等待%.2f秒后重试...", sleepTime)
			time.Sleep(time.Duration(sleepTime * float64(time.Second)))
			continue
		}
		break
	}

	return &VerificationResult{
		UserID:   userID,
		Nickname: nickname,
		Password: password,
	}, nil
}

// ChangePassword 修改密码
// 根据objectNum判断用户类型：1-运营用户，2-客服用户
// 验证旧密码正确后，使用bcrypt加密新密码并更新到数据库
func ChangePassword(req requestbody.ChangePasswordRequest) error {
	var tableName string
	if req.ObjectNum == 1 {
		tableName = "Operation_user"
	} else {
		tableName = "Customer_service_user"
	}

	maxRetries := 3
	retryCount := 0

	for retryCount < maxRetries {
		tryAgain := false
		retryCount++

		var userID, currentPassword string
		var err error
		query := fmt.Sprintf("SELECT user_id, password FROM %s WHERE mobile = ?", tableName)
		if db.DB != nil {
			sqlDB, err := db.DB.DB()
			if err != nil {
				log.Printf("获取数据库连接失败: %v", err)
				err = fmt.Errorf("数据库连接错误")
			} else {
				err = sqlDB.QueryRow(query, req.Mobile).Scan(&userID, &currentPassword)
			}
		} else {
			err = fmt.Errorf("数据库实例未初始化")
		}

		if err != nil {
			if err == sql.ErrNoRows {
				return fmt.Errorf("用户不存在")
			}
			log.Printf("数据库查询用户失败: %v", err)
			tryAgain = true
		} else {
			err = bcrypt.CompareHashAndPassword([]byte(currentPassword), []byte(req.OldPassword))
			if err != nil {
				return fmt.Errorf("旧密码错误")
			}

			if tryAgain {
				continue
			}

			newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
			if err != nil {
				log.Printf("密码哈希失败: %v", err)
				tryAgain = true
			}

			if tryAgain {
				continue
			}

			tx := db.DB.Begin()
			if tx.Error != nil {
				log.Printf("开始事务失败: %v", tx.Error)
				tryAgain = true
			} else {
				updateQuery := fmt.Sprintf("UPDATE %s SET password = ? WHERE user_id = ?", tableName)
				err = tx.Exec(updateQuery, string(newHashedPassword), userID).Error
				if err != nil {
					log.Printf("更新密码失败: %v", err)
					tx.Rollback()
					tryAgain = true
				} else {
					if err := tx.Commit(); err != nil {
						log.Printf("提交事务失败: %v", err)
						tx.Rollback()
						tryAgain = true
					} else {
						log.Printf("用户 %s 密码更新成功", userID)
						return nil
					}
				}
			}
		}

		if tryAgain && retryCount < maxRetries {
			shiftValue := 1 << uint(retryCount-1)
			sleepTime := 0.5 * float64(shiftValue)
			log.Printf("%.2f秒后重试，尝试 %d/%d", sleepTime, retryCount, maxRetries)
			time.Sleep(time.Duration(sleepTime * float64(time.Second)))
		}
	}

	return fmt.Errorf("服务器错误，重试多次后失败")
}
