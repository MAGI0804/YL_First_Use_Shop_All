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

	var user models.User
	openIDErr := tx.Where("openid = ?", req.OpenID).First(&user).Error
	if errors.Is(openIDErr, gorm.ErrRecordNotFound) {
		user = models.User{
			OpenID:   req.OpenID,
			Mobile:   req.Mobile,
			IsActive: true,
		}
		if err := tx.Create(&user).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
	} else if openIDErr != nil {
		tx.Rollback()
		return nil, openIDErr
	} else if user.Mobile != "" && user.Mobile != req.Mobile {
		tx.Rollback()
		return nil, fmt.Errorf("openid already bound to another mobile")
	} else if user.Mobile == "" {
		// Keep the existing users_user mobile field as the WeChat authorized mobile.
		if err := tx.Model(&user).Update("mobile", req.Mobile).Error; err != nil {
			tx.Rollback()
			return nil, err
		}
		user.Mobile = req.Mobile
	}

	var otherUser models.User
	otherUserErr := tx.Where("mobile = ? AND openid <> ?", req.Mobile, req.OpenID).First(&otherUser).Error
	if otherUserErr == nil {
		tx.Rollback()
		return nil, fmt.Errorf("mobile already bound to another openid")
	}
	if otherUserErr != nil && !errors.Is(otherUserErr, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, otherUserErr
	}

	var member models.Member
	openIDMemberErr := tx.Where("openid = ?", req.OpenID).First(&member).Error
	if openIDMemberErr == nil && member.Mobile != req.Mobile {
		tx.Rollback()
		return nil, fmt.Errorf("openid already linked to another member mobile")
	}
	if openIDMemberErr != nil && !errors.Is(openIDMemberErr, gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, openIDMemberErr
	}

	if errors.Is(openIDMemberErr, gorm.ErrRecordNotFound) {
		memberErr := tx.Where("mobile = ?", req.Mobile).First(&member).Error
		if memberErr == nil {
			if member.OpenID != "" && member.OpenID != req.OpenID {
				tx.Rollback()
				return nil, fmt.Errorf("mobile already linked to another member")
			}
			if member.UserID != 0 && member.UserID != user.UserID {
				tx.Rollback()
				return nil, fmt.Errorf("member already linked to another user")
			}
			if err := tx.Model(&member).Updates(map[string]interface{}{
				"user_id": user.UserID,
				"openid":  req.OpenID,
			}).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		} else if errors.Is(memberErr, gorm.ErrRecordNotFound) {
			member = models.Member{
				UserID:   user.UserID,
				OpenID:   req.OpenID,
				Mobile:   req.Mobile,
				Nickname: user.Nickname,
			}
			if err := tx.Create(&member).Error; err != nil {
				tx.Rollback()
				return nil, err
			}
		} else {
			tx.Rollback()
			return nil, memberErr
		}
	} else if err := tx.Model(&member).Updates(map[string]interface{}{
		"user_id": user.UserID,
		"openid":  req.OpenID,
	}).Error; err != nil {
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
