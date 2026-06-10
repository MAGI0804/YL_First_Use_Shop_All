package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"
)

type MemberDetailResult struct {
	Member models.Member      `json:"member"`
	Tags   []models.MemberTag `json:"tags"`
}

func CreateMember(req requestbody.MemberCreateRequest, operator BackendOperatorSnapshot, requestMeta OperationRequestMeta) (*models.Member, error) {
	if err := validateMemberMobile(req.Mobile); err != nil {
		return nil, err
	}
	req.ManualUniqueCode = strings.TrimSpace(req.ManualUniqueCode)

	var member models.Member
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := ensureMemberUnique(tx, 0, req.Mobile, req.ManualUniqueCode); err != nil {
			return err
		}
		member = models.Member{
			MemberNo:         strings.TrimSpace(req.MemberNo),
			UserID:           req.UserID,
			OpenID:           normalizeMemberOpenID(req.OpenID),
			Mobile:           strings.TrimSpace(req.Mobile),
			ManualUniqueCode: req.ManualUniqueCode,
			Nickname:         strings.TrimSpace(req.Nickname),
			Status:           normalizeMemberStatus(req.Status),
			Source:           normalizeMemberSource(req.Source),
			TmallID:          strings.TrimSpace(req.TmallID),
			TmallAmount:      req.TmallAmount,
			YouzanID:         strings.TrimSpace(req.YouzanID),
			YouzanAmount:     req.YouzanAmount,
			Remarks:          strings.TrimSpace(req.Remarks),
		}
		if err := tx.Create(&member).Error; err != nil {
			return err
		}
		return recordBackendOperation(tx, BackendOperationLogInput{
			Operator:   operator,
			Action:     ActionMemberCreate,
			Module:     OperationModuleMember,
			TargetType: "member",
			TargetID:   strconv.FormatUint(uint64(member.ID), 10),
			MemberID:   member.ID,
			UserID:     member.UserID,
			AfterData:  memberOperationSnapshot(member),
			RequestID:  requestMeta.RequestID,
			ClientIP:   requestMeta.ClientIP,
			UserAgent:  requestMeta.UserAgent,
		})
	})
	if err != nil {
		return nil, err
	}
	return &member, nil
}

func UpdateMember(req requestbody.MemberUpdateRequest, operator BackendOperatorSnapshot, requestMeta OperationRequestMeta) (*models.Member, error) {
	var updated models.Member
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var member models.Member
		if err := tx.Where("id = ?", req.ID).First(&member).Error; err != nil {
			return err
		}
		before := memberOperationSnapshot(member)
		mobile := strings.TrimSpace(req.Mobile)
		if mobile == "" {
			mobile = member.Mobile
		}
		if err := validateMemberMobile(mobile); err != nil {
			return err
		}
		manualUniqueCode := strings.TrimSpace(req.ManualUniqueCode)
		if err := ensureMemberUnique(tx, req.ID, mobile, manualUniqueCode); err != nil {
			return err
		}

		member.MemberNo = keepOrTrim(req.MemberNo, member.MemberNo)
		member.UserID = keepOrInt(req.UserID, member.UserID)
		member.OpenID = keepOrMemberOpenID(req.OpenID, member.OpenID)
		member.Mobile = mobile
		member.ManualUniqueCode = manualUniqueCode
		member.Nickname = keepOrTrim(req.Nickname, member.Nickname)
		member.Status = normalizeMemberStatus(keepOrTrim(req.Status, member.Status))
		member.Source = normalizeMemberSource(keepOrTrim(req.Source, member.Source))
		member.TmallID = strings.TrimSpace(req.TmallID)
		member.TmallAmount = req.TmallAmount
		member.YouzanID = strings.TrimSpace(req.YouzanID)
		member.YouzanAmount = req.YouzanAmount
		member.Remarks = strings.TrimSpace(req.Remarks)

		if err := tx.Save(&member).Error; err != nil {
			return err
		}
		updated = member
		return recordBackendOperation(tx, BackendOperationLogInput{
			Operator:   operator,
			Action:     ActionMemberUpdate,
			Module:     OperationModuleMember,
			TargetType: "member",
			TargetID:   strconv.FormatUint(uint64(member.ID), 10),
			MemberID:   member.ID,
			UserID:     member.UserID,
			BeforeData: before,
			AfterData:  memberOperationSnapshot(member),
			RequestID:  requestMeta.RequestID,
			ClientIP:   requestMeta.ClientIP,
			UserAgent:  requestMeta.UserAgent,
		})
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func QueryMembers(req requestbody.MemberListRequest) ([]models.Member, int64, error) {
	page, pageSize := normalizeBackendPage(req.Page, req.PageSize)
	query := db.DB.Model(&models.Member{})
	if strings.TrimSpace(req.Mobile) != "" {
		query = query.Where("mobile LIKE ?", "%"+strings.TrimSpace(req.Mobile)+"%")
	}
	if strings.TrimSpace(req.MemberNo) != "" {
		query = query.Where("member_no LIKE ?", "%"+strings.TrimSpace(req.MemberNo)+"%")
	}
	if strings.TrimSpace(req.ManualUniqueCode) != "" {
		query = query.Where("manual_unique_code LIKE ?", "%"+strings.TrimSpace(req.ManualUniqueCode)+"%")
	}
	if strings.TrimSpace(req.Nickname) != "" {
		query = query.Where("nickname LIKE ?", "%"+strings.TrimSpace(req.Nickname)+"%")
	}
	if strings.TrimSpace(req.Status) != "" {
		query = query.Where("status = ?", strings.TrimSpace(req.Status))
	}
	if req.TagID > 0 {
		query = query.Joins("JOIN member_tag_relation mtr ON mtr.member_id = member_info.id").Where("mtr.tag_id = ?", req.TagID)
	}
	if strings.TrimSpace(req.TagName) != "" {
		query = query.Joins("JOIN member_tag_relation mtr_name ON mtr_name.member_id = member_info.id").
			Joins("JOIN member_tag mt_name ON mt_name.id = mtr_name.tag_id").
			Where("mt_name.name LIKE ?", "%"+strings.TrimSpace(req.TagName)+"%")
	}

	var total int64
	if err := query.Distinct("member_info.id").Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var members []models.Member
	if err := query.Distinct("member_info.*").Order("created_at DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&members).Error; err != nil {
		return nil, 0, err
	}
	return members, total, nil
}

func GetMemberDetail(req requestbody.MemberDetailRequest) (*MemberDetailResult, error) {
	member, err := ResolveMember(req.ID, req.MemberNo, req.Mobile, req.UserID)
	if err != nil {
		return nil, err
	}
	tags, err := GetMemberTags(member.ID)
	if err != nil {
		return nil, err
	}
	return &MemberDetailResult{Member: *member, Tags: tags}, nil
}

func CreateMemberTag(req requestbody.MemberTagCreateRequest, operator BackendOperatorSnapshot) (*models.MemberTag, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, fmt.Errorf("tag name is required")
	}
	var tag models.MemberTag
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("name = ?", name).First(&tag).Error; err == nil {
			return fmt.Errorf("tag already exists")
		} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		tag = models.MemberTag{
			Name:      name,
			Color:     strings.TrimSpace(req.Color),
			Remarks:   strings.TrimSpace(req.Remarks),
			CreatedBy: operator.ID,
		}
		return tx.Create(&tag).Error
	})
	if err != nil {
		return nil, err
	}
	return &tag, nil
}

func QueryMemberTags(req requestbody.MemberTagListRequest) ([]models.MemberTag, int64, error) {
	page, pageSize := normalizeBackendPage(req.Page, req.PageSize)
	query := db.DB.Model(&models.MemberTag{})
	if strings.TrimSpace(req.Name) != "" {
		query = query.Where("name LIKE ?", "%"+strings.TrimSpace(req.Name)+"%")
	}
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var tags []models.MemberTag
	if err := query.Order("created_at DESC, id DESC").Offset((page - 1) * pageSize).Limit(pageSize).Find(&tags).Error; err != nil {
		return nil, 0, err
	}
	return tags, total, nil
}

func SetMemberTags(req requestbody.MemberTagSetRequest, operator BackendOperatorSnapshot, requestMeta OperationRequestMeta) ([]models.MemberTag, error) {
	var tags []models.MemberTag
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		var member models.Member
		if err := tx.Where("id = ?", req.MemberID).First(&member).Error; err != nil {
			return err
		}
		beforeTags, err := getMemberTagsTx(tx, req.MemberID)
		if err != nil {
			return err
		}
		if len(req.TagIDs) > 0 {
			var count int64
			if err := tx.Model(&models.MemberTag{}).Where("id IN ?", req.TagIDs).Count(&count).Error; err != nil {
				return err
			}
			if count != int64(len(uniqueUint(req.TagIDs))) {
				return fmt.Errorf("tag not found")
			}
		}
		if err := tx.Where("member_id = ?", req.MemberID).Delete(&models.MemberTagRelation{}).Error; err != nil {
			return err
		}
		for _, tagID := range uniqueUint(req.TagIDs) {
			relation := models.MemberTagRelation{MemberID: req.MemberID, TagID: tagID, CreatedBy: operator.ID}
			if err := tx.Create(&relation).Error; err != nil {
				return err
			}
		}
		afterTags, err := getMemberTagsTx(tx, req.MemberID)
		if err != nil {
			return err
		}
		tags = afterTags
		return recordBackendOperation(tx, BackendOperationLogInput{
			Operator:   operator,
			Action:     ActionMemberTagSet,
			Module:     OperationModuleMember,
			TargetType: "member",
			TargetID:   strconv.FormatUint(uint64(member.ID), 10),
			MemberID:   member.ID,
			UserID:     member.UserID,
			BeforeData: map[string]any{"tags": beforeTags},
			AfterData:  map[string]any{"tags": afterTags},
			RequestID:  requestMeta.RequestID,
			ClientIP:   requestMeta.ClientIP,
			UserAgent:  requestMeta.UserAgent,
		})
	})
	if err != nil {
		return nil, err
	}
	return tags, nil
}

func GetMemberTags(memberID uint) ([]models.MemberTag, error) {
	return getMemberTagsTx(db.DB, memberID)
}

func getMemberTagsTx(tx *gorm.DB, memberID uint) ([]models.MemberTag, error) {
	var tags []models.MemberTag
	err := tx.Table("member_tag").
		Joins("JOIN member_tag_relation ON member_tag_relation.tag_id = member_tag.id").
		Where("member_tag_relation.member_id = ?", memberID).
		Order("member_tag.name ASC").
		Find(&tags).Error
	return tags, err
}

func ResolveMember(memberID uint, memberNo, mobile string, userID int) (*models.Member, error) {
	query := db.DB.Model(&models.Member{})
	if memberID > 0 {
		query = query.Where("id = ?", memberID)
	} else if strings.TrimSpace(memberNo) != "" {
		query = query.Where("member_no = ?", strings.TrimSpace(memberNo))
	} else if strings.TrimSpace(mobile) != "" {
		query = query.Where("mobile = ?", strings.TrimSpace(mobile))
	} else if userID > 0 {
		query = query.Where("user_id = ?", userID)
	} else {
		return nil, fmt.Errorf("missing member query condition")
	}
	var member models.Member
	if err := query.First(&member).Error; err != nil {
		return nil, err
	}
	return &member, nil
}

type OperationRequestMeta struct {
	RequestID string
	ClientIP  string
	UserAgent string
}

func validateMemberMobile(mobile string) error {
	mobile = strings.TrimSpace(mobile)
	if !IsValidMobile(mobile) {
		return fmt.Errorf("invalid mobile")
	}
	return nil
}

func normalizeMemberOpenID(openID string) string {
	normalized := strings.TrimSpace(openID)
	if normalized == "0" {
		return ""
	}
	return normalized
}

func keepOrMemberOpenID(next, current string) string {
	normalized := normalizeMemberOpenID(next)
	if normalized == "" {
		return normalizeMemberOpenID(current)
	}
	return normalized
}

func ensureMemberUnique(tx *gorm.DB, currentID uint, mobile, manualUniqueCode string) error {
	var count int64
	mobile = strings.TrimSpace(mobile)
	if mobile != "" {
		query := tx.Model(&models.Member{}).Where("mobile = ?", mobile)
		if currentID > 0 {
			query = query.Where("id <> ?", currentID)
		}
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("mobile already exists")
		}
	}
	manualUniqueCode = strings.TrimSpace(manualUniqueCode)
	if manualUniqueCode != "" {
		query := tx.Model(&models.Member{}).Where("manual_unique_code = ?", manualUniqueCode)
		if currentID > 0 {
			query = query.Where("id <> ?", currentID)
		}
		if err := query.Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return fmt.Errorf("manual unique code already exists")
		}
	}
	return nil
}

func normalizeMemberStatus(status string) string {
	status = strings.TrimSpace(status)
	if status == "" {
		return "active"
	}
	return status
}

func normalizeMemberSource(source string) string {
	source = strings.TrimSpace(source)
	if source == "" {
		return "backend"
	}
	return source
}

func keepOrTrim(next, current string) string {
	next = strings.TrimSpace(next)
	if next == "" {
		return current
	}
	return next
}

func keepOrInt(next, current int) int {
	if next == 0 {
		return current
	}
	return next
}

func uniqueUint(values []uint) []uint {
	seen := make(map[uint]bool, len(values))
	result := make([]uint, 0, len(values))
	for _, value := range values {
		if value == 0 || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func memberOperationSnapshot(member models.Member) map[string]any {
	return map[string]any{
		"id":                 member.ID,
		"member_no":          member.MemberNo,
		"user_id":            member.UserID,
		"openid":             member.OpenID,
		"mobile":             member.Mobile,
		"manual_unique_code": member.ManualUniqueCode,
		"nickname":           member.Nickname,
		"status":             member.Status,
		"source":             member.Source,
		"tmall_id":           member.TmallID,
		"tmall_amount":       member.TmallAmount,
		"youzan_id":          member.YouzanID,
		"youzan_amount":      member.YouzanAmount,
		"remarks":            member.Remarks,
	}
}
