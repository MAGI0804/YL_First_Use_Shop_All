package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
	"Member_shop/utils"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

var mobileRegex = regexp.MustCompile(`^1[3-9]\d{9}$`)

// IsValidMobile validates mainland China mobile numbers used by member/backend registration.
func IsValidMobile(mobile string) bool {
	return mobileRegex.MatchString(mobile)
}

// SelectUserInfo returns WeChat user data and, when linked, member business data.
func SelectUserInfo(t gin.Context, userID int) *map[string]any {
	var user models.User
	err := db.DB.Where("user_id = ?", userID).First(&user).Error
	if err != nil {
		return nil
	}

	info := map[string]any{
		"user_id":           user.UserID,
		"openid":            user.OpenID,
		"mobile":            user.Mobile,
		"nickname":          user.Nickname,
		"default_receiver":  user.DefaultReceiver,
		"province":          user.Province,
		"city":              user.City,
		"county":            user.County,
		"detailed_address":  user.DetailedAddress,
		"membership_level":  user.MembershipLevel,
		"registration_date": user.RegistrationDate,
		"total_spending":    user.TotalSpending,
		"remarks":           user.Remarks,
		"is_active":         user.IsActive,
		"is_staff":          user.IsStaff,
	}

	// Member is separate from User, so member totals and platform IDs are read from member_info.
	var member models.Member
	memberErr := db.DB.Where("user_id = ?", user.UserID).First(&member).Error
	if memberErr != nil && user.OpenID != "" {
		memberErr = db.DB.Where("openid = ?", user.OpenID).First(&member).Error
	}
	if memberErr == nil {
		info["member_no"] = member.MemberNo
		info["member_mobile"] = member.Mobile
		info["total_order_amount"] = member.TotalOrderAmount
		info["total_paid_amount"] = member.TotalPaidAmount
		info["tmall_id"] = member.TmallID
		info["tmall_amount"] = member.TmallAmount
		info["youzan_id"] = member.YouzanID
		info["youzan_amount"] = member.YouzanAmount
	}

	if user.UserImg != "" {
		proto := utils.GetRequestProto(&t)
		baseURL := fmt.Sprintf("%s://%s", proto, t.Request.Host)
		fullImagePath := user.UserImg
		if strings.HasPrefix(user.UserImg, "media/") {
			fullImagePath = "/" + user.UserImg
		} else if !strings.HasPrefix(user.UserImg, "/") &&
			!strings.HasPrefix(user.UserImg, "http://") &&
			!strings.HasPrefix(user.UserImg, "https://") {
			fullImagePath = "/media/" + user.UserImg
		}
		if strings.HasPrefix(user.UserImg, "http://") || strings.HasPrefix(user.UserImg, "https://") {
			info["user_img"] = user.UserImg
		} else {
			info["user_img"] = utils.BuildFullImageURL(baseURL, fullImagePath)
		}
	}
	return &info
}

// BindWechatPhone links a WeChat user to a member mobile.
// User remains the WeChat identity table; Member stores the actual member record.
func BindWechatPhone(req requestbody.BindWechatPhoneRequest) (*models.Member, error) {
	if !IsValidMobile(req.Mobile) {
		return nil, fmt.Errorf("invalid mobile")
	}
	if req.Captcha != "" {
		if err := db.VerifyCaptcha(req.Mobile, req.Captcha); err != nil {
			return nil, err
		}
	}

	tx := db.DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}

	var member models.Member
	if err := tx.Where("mobile = ?", req.Mobile).First(&member).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("mobile is not a member")
		}
		return nil, err
	}
	if err := validateMemberStatus(member); err != nil {
		tx.Rollback()
		return nil, err
	}
	if member.OpenID != "" && member.OpenID != req.OpenID {
		tx.Rollback()
		return nil, fmt.Errorf("mobile already linked to another wechat user")
	}

	user, err := findOrCreateWechatMemberUser(tx, member, req)
	if err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := validateMemberWechatLink(member, user, req.OpenID, req.Mobile); err != nil {
		tx.Rollback()
		return nil, err
	}

	memberUpdates := map[string]interface{}{
		"user_id": user.UserID,
		"openid":  req.OpenID,
	}
	if strings.TrimSpace(member.Nickname) == "" && strings.TrimSpace(req.Nickname) != "" {
		memberUpdates["nickname"] = strings.TrimSpace(req.Nickname)
	}
	if err := tx.Model(&member).Updates(memberUpdates).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.First(&member, member.ID).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	return &member, nil
}

func findOrCreateWechatMemberUser(tx *gorm.DB, member models.Member, req requestbody.BindWechatPhoneRequest) (models.User, error) {
	var user models.User
	openIDErr := tx.Where("openid = ?", req.OpenID).First(&user).Error
	if openIDErr == nil {
		return updateWechatMemberUser(tx, user, req)
	}
	if openIDErr != nil && !errors.Is(openIDErr, gorm.ErrRecordNotFound) {
		return models.User{}, openIDErr
	}

	mobileErr := tx.Where("mobile = ?", req.Mobile).First(&user).Error
	if mobileErr == nil {
		if user.OpenID != "" && user.OpenID != req.OpenID {
			return models.User{}, fmt.Errorf("mobile already bound to another openid")
		}
		if member.UserID != 0 && member.UserID != user.UserID {
			return models.User{}, fmt.Errorf("member already linked to another user")
		}
		return updateWechatMemberUser(tx, user, req)
	}
	if mobileErr != nil && !errors.Is(mobileErr, gorm.ErrRecordNotFound) {
		return models.User{}, mobileErr
	}

	if member.UserID != 0 {
		idErr := tx.Where("user_id = ?", member.UserID).First(&user).Error
		if idErr == nil {
			return updateWechatMemberUser(tx, user, req)
		}
		if idErr != nil && !errors.Is(idErr, gorm.ErrRecordNotFound) {
			return models.User{}, idErr
		}
	}

	user = models.User{
		OpenID:           req.OpenID,
		Mobile:           req.Mobile,
		Nickname:         normalizedWechatNickname(req.Nickname, req.OpenID),
		UserImg:          normalizedWechatAvatar(req.AvatarURL),
		RegistrationDate: time.Now(),
		LastLogin:        time.Now(),
		IsActive:         true,
		IsStaff:          false,
	}
	if err := tx.Create(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func updateWechatMemberUser(tx *gorm.DB, user models.User, req requestbody.BindWechatPhoneRequest) (models.User, error) {
	if user.Mobile != "" && user.Mobile != req.Mobile {
		return models.User{}, fmt.Errorf("openid already bound to another mobile")
	}
	if user.OpenID != "" && user.OpenID != req.OpenID {
		return models.User{}, fmt.Errorf("mobile already bound to another openid")
	}

	updates := map[string]interface{}{
		"openid":     req.OpenID,
		"mobile":     req.Mobile,
		"last_login": time.Now(),
		"is_active":  true,
	}
	nickname := strings.TrimSpace(req.Nickname)
	if shouldUpdateWechatNickname(user.Nickname, nickname) {
		updates["nickname"] = nickname
	}
	avatarURL := normalizedWechatAvatar(req.AvatarURL)
	if user.UserImg == "" && avatarURL != "" {
		updates["user_img"] = avatarURL
	}
	if err := tx.Model(&user).Updates(updates).Error; err != nil {
		return models.User{}, err
	}
	if err := tx.First(&user, user.UserID).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func validateMemberStatus(member models.Member) error {
	if strings.EqualFold(strings.TrimSpace(member.Status), "disabled") {
		return fmt.Errorf("member disabled")
	}
	return nil
}

func validateMemberWechatLink(member models.Member, user models.User, openID, mobile string) error {
	if member.Mobile != mobile {
		return fmt.Errorf("member mobile mismatch")
	}
	if err := validateMemberStatus(member); err != nil {
		return err
	}
	if member.OpenID != "" && member.OpenID != openID {
		return fmt.Errorf("mobile already linked to another wechat user")
	}
	if member.UserID != 0 && user.UserID != 0 && member.UserID != user.UserID {
		return fmt.Errorf("member already linked to another user")
	}
	if user.OpenID != "" && user.OpenID != openID {
		return fmt.Errorf("mobile already bound to another openid")
	}
	if user.Mobile != "" && user.Mobile != mobile {
		return fmt.Errorf("openid already bound to another mobile")
	}
	return nil
}

func shouldUpdateWechatNickname(current, next string) bool {
	current = strings.TrimSpace(current)
	next = strings.TrimSpace(next)
	return next != "" && (current == "" || strings.HasPrefix(current, "微信用户_") || strings.HasPrefix(current, "wechat_user_") || strings.HasPrefix(current, "mobile_user_"))
}

func normalizedWechatNickname(nickname, openID string) string {
	if strings.TrimSpace(nickname) != "" {
		return strings.TrimSpace(nickname)
	}
	if len(openID) > 8 {
		return "微信用户_" + openID[:8]
	}
	return "微信用户_" + openID
}

func normalizedWechatAvatar(avatarURL string) string {
	avatarURL = strings.TrimSpace(avatarURL)
	if strings.HasPrefix(avatarURL, "wxfile://") || strings.HasPrefix(avatarURL, "http://tmp/") {
		return ""
	}
	return avatarURL
}

func EnsureActiveMemberUser(userID int) error {
	if userID <= 0 {
		return fmt.Errorf("user is not a member")
	}
	var member models.Member
	if err := db.DB.Where("user_id = ?", userID).First(&member).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user is not a member")
		}
		return err
	}
	return validateMemberStatus(member)
}

// UpdatePlatformInfo updates member platform IDs and amounts.
func UpdatePlatformInfo(req requestbody.UpdatePlatformInfoRequest) (*models.Member, error) {
	var member models.Member
	if err := db.DB.Where("user_id = ?", req.UserID).First(&member).Error; err != nil {
		return nil, err
	}
	member.TmallID = req.TmallID
	member.TmallAmount = req.TmallAmount
	member.YouzanID = req.YouzanID
	member.YouzanAmount = req.YouzanAmount
	if err := db.DB.Save(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

// MemberAmountSummary reads totals from the member table.
func MemberAmountSummary(req requestbody.MemberAmountSummaryRequest) (*models.Member, error) {
	query := db.DB.Model(&models.Member{})
	switch {
	case req.UserID > 0:
		query = query.Where("user_id = ?", req.UserID)
	case req.MemberNo != "":
		query = query.Where("member_no = ?", req.MemberNo)
	case req.Mobile != "":
		query = query.Where("mobile = ?", req.Mobile)
	default:
		return nil, fmt.Errorf("missing query condition")
	}

	var member models.Member
	if err := query.First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

// AddUserOrderAmount accumulates the member's total order amount after backend order creation.
func AddUserOrderAmount(userID int, amount float64) error {
	if userID <= 0 || amount <= 0 {
		return nil
	}
	return db.DB.Model(&models.Member{}).
		Where("user_id = ?", userID).
		Update("total_order_amount", gorm.Expr("total_order_amount + ?", amount)).Error
}

// AddUserPaidAmount accumulates the member's paid amount after a delivered order is paid.
func AddUserPaidAmount(userID int, amount float64) error {
	if userID <= 0 || amount <= 0 {
		return nil
	}
	return db.DB.Model(&models.Member{}).
		Where("user_id = ?", userID).
		Update("total_paid_amount", gorm.Expr("total_paid_amount + ?", amount)).Error
}
