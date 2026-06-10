package method

import (
	"Member_shop/db"
	"Member_shop/models"
	"Member_shop/requestbody"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
)

const memberImportMaxRows = 1000

var memberImportHeaders = []string{
	"手机号(必填)",
	"唯一字段",
	"昵称",
	"天猫ID",
	"天猫金额",
	"有赞ID",
	"有赞金额",
	"备注",
}

type MemberImportRow struct {
	RowIndex         int      `json:"row_index"`
	Mobile           string   `json:"mobile"`
	ManualUniqueCode string   `json:"manual_unique_code"`
	Nickname         string   `json:"nickname"`
	TmallID          string   `json:"tmall_id"`
	TmallAmount      float64  `json:"tmall_amount"`
	YouzanID         string   `json:"youzan_id"`
	YouzanAmount     float64  `json:"youzan_amount"`
	Remarks          string   `json:"remarks"`
	Matched          bool     `json:"matched"`
	Errors           []string `json:"errors"`
}

type MemberImportMatchResult struct {
	Items        []MemberImportRow `json:"items"`
	TotalRows    int               `json:"total_rows"`
	MatchedCount int               `json:"matched_count"`
	InvalidCount int               `json:"invalid_count"`
}

type MemberImportConfirmResult struct {
	ImportedCount int             `json:"imported_count"`
	Members       []models.Member `json:"members"`
}

func BuildMemberImportTemplate() ([]byte, error) {
	file := excelize.NewFile()
	defer file.Close()

	sheet := "会员导入模板"
	defaultSheet := file.GetSheetName(0)
	if err := file.SetSheetName(defaultSheet, sheet); err != nil {
		return nil, err
	}
	for index, header := range memberImportHeaders {
		cell, err := excelize.CoordinatesToCellName(index+1, 1)
		if err != nil {
			return nil, err
		}
		if err := file.SetCellValue(sheet, cell, header); err != nil {
			return nil, err
		}
	}
	widths := []float64{18, 18, 16, 18, 12, 18, 12, 28}
	for index, width := range widths {
		column := excelColumnName(index + 1)
		if err := file.SetColWidth(sheet, column, column, width); err != nil {
			return nil, err
		}
	}
	samples := []any{"13800138000", "VIP001", "张三", "tmall_001", 99.9, "youzan_001", 88.8, "示例行，导入前可删除"}
	for index, value := range samples {
		cell, err := excelize.CoordinatesToCellName(index+1, 2)
		if err != nil {
			return nil, err
		}
		if err := file.SetCellValue(sheet, cell, value); err != nil {
			return nil, err
		}
	}
	buffer, err := file.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func MatchMemberImportFile(reader io.Reader) (*MemberImportMatchResult, error) {
	file, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("无法读取Excel文件")
	}
	defer file.Close()

	sheets := file.GetSheetList()
	if len(sheets) == 0 {
		return nil, fmt.Errorf("Excel文件没有工作表")
	}
	rows, err := file.GetRows(sheets[0])
	if err != nil {
		return nil, err
	}
	if len(rows) < 2 {
		return &MemberImportMatchResult{Items: []MemberImportRow{}}, nil
	}
	headerMap := buildMemberImportHeaderMap(rows[0])
	if _, ok := headerMap["mobile"]; !ok {
		return nil, fmt.Errorf("模板缺少手机号列")
	}

	items := make([]MemberImportRow, 0, len(rows)-1)
	seenMobiles := map[string]int{}
	seenUniqueCodes := map[string]int{}
	for rowIndex := 1; rowIndex < len(rows); rowIndex++ {
		if len(items) >= memberImportMaxRows {
			return nil, fmt.Errorf("单次最多导入%d行", memberImportMaxRows)
		}
		item := buildMemberImportRow(rowIndex+1, rows[rowIndex], headerMap)
		if isEmptyMemberImportRow(item) {
			continue
		}
		validateMemberImportRow(&item)
		if item.Mobile != "" {
			if firstRow, ok := seenMobiles[item.Mobile]; ok {
				item.Errors = append(item.Errors, fmt.Sprintf("手机号与第%d行重复", firstRow))
			} else {
				seenMobiles[item.Mobile] = item.RowIndex
			}
		}
		if item.ManualUniqueCode != "" {
			if firstRow, ok := seenUniqueCodes[item.ManualUniqueCode]; ok {
				item.Errors = append(item.Errors, fmt.Sprintf("唯一字段与第%d行重复", firstRow))
			} else {
				seenUniqueCodes[item.ManualUniqueCode] = item.RowIndex
			}
		}
		items = append(items, item)
	}
	if err := markExistingMemberImportRows(items); err != nil {
		return nil, err
	}
	return summarizeMemberImportRows(items), nil
}

func ConfirmMemberImport(req requestbody.MemberImportConfirmRequest, operator BackendOperatorSnapshot, requestMeta OperationRequestMeta) (*MemberImportConfirmResult, error) {
	items := make([]MemberImportRow, 0, len(req.Items))
	seenMobiles := map[string]int{}
	seenUniqueCodes := map[string]int{}
	for index, input := range req.Items {
		if index >= memberImportMaxRows {
			return nil, fmt.Errorf("单次最多导入%d行", memberImportMaxRows)
		}
		item := memberImportRowFromRequest(index+1, input)
		validateMemberImportRow(&item)
		if item.Mobile != "" {
			if firstRow, ok := seenMobiles[item.Mobile]; ok {
				item.Errors = append(item.Errors, fmt.Sprintf("手机号与第%d条重复", firstRow))
			} else {
				seenMobiles[item.Mobile] = item.RowIndex
			}
		}
		if item.ManualUniqueCode != "" {
			if firstRow, ok := seenUniqueCodes[item.ManualUniqueCode]; ok {
				item.Errors = append(item.Errors, fmt.Sprintf("唯一字段与第%d条重复", firstRow))
			} else {
				seenUniqueCodes[item.ManualUniqueCode] = item.RowIndex
			}
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("没有可导入的数据")
	}
	if err := markExistingMemberImportRows(items); err != nil {
		return nil, err
	}
	for _, item := range items {
		if len(item.Errors) > 0 {
			return nil, fmt.Errorf("导入数据已失效，请重新匹配后再确认")
		}
	}

	result := &MemberImportConfirmResult{Members: make([]models.Member, 0, len(items))}
	err := db.DB.Transaction(func(tx *gorm.DB) error {
		for _, item := range items {
			member := models.Member{
				Mobile:           item.Mobile,
				ManualUniqueCode: item.ManualUniqueCode,
				Nickname:         item.Nickname,
				Status:           "active",
				Source:           "backend_import",
				TmallID:          item.TmallID,
				TmallAmount:      item.TmallAmount,
				YouzanID:         item.YouzanID,
				YouzanAmount:     item.YouzanAmount,
				Remarks:          item.Remarks,
			}
			if err := ensureMemberUnique(tx, 0, member.Mobile, member.ManualUniqueCode); err != nil {
				return err
			}
			if err := tx.Create(&member).Error; err != nil {
				return err
			}
			if err := recordBackendOperation(tx, BackendOperationLogInput{
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
				Remark:     "会员名单导入",
			}); err != nil {
				return err
			}
			result.Members = append(result.Members, member)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	result.ImportedCount = len(result.Members)
	return result, nil
}

func buildMemberImportHeaderMap(headers []string) map[string]int {
	result := map[string]int{}
	for index, header := range headers {
		key := normalizeMemberImportHeader(header)
		if key != "" {
			result[key] = index
		}
	}
	return result
}

func normalizeMemberImportHeader(header string) string {
	header = strings.TrimSpace(strings.ToLower(header))
	header = strings.ReplaceAll(header, "（必填）", "")
	header = strings.ReplaceAll(header, "(必填)", "")
	header = strings.ReplaceAll(header, " ", "")
	switch header {
	case "手机号", "mobile":
		return "mobile"
	case "唯一字段", "manual_unique_code":
		return "manual_unique_code"
	case "昵称", "nickname":
		return "nickname"
	case "天猫id", "tmall_id":
		return "tmall_id"
	case "天猫金额", "tmall_amount":
		return "tmall_amount"
	case "有赞id", "youzan_id":
		return "youzan_id"
	case "有赞金额", "youzan_amount":
		return "youzan_amount"
	case "备注", "remarks":
		return "remarks"
	default:
		return ""
	}
}

func buildMemberImportRow(rowIndex int, cells []string, headerMap map[string]int) MemberImportRow {
	item := MemberImportRow{RowIndex: rowIndex}
	item.Mobile = memberImportCell(cells, headerMap, "mobile")
	item.ManualUniqueCode = memberImportCell(cells, headerMap, "manual_unique_code")
	item.Nickname = memberImportCell(cells, headerMap, "nickname")
	item.TmallID = memberImportCell(cells, headerMap, "tmall_id")
	item.YouzanID = memberImportCell(cells, headerMap, "youzan_id")
	item.Remarks = memberImportCell(cells, headerMap, "remarks")
	item.TmallAmount = parseMemberImportAmount(memberImportCell(cells, headerMap, "tmall_amount"), &item, "天猫金额")
	item.YouzanAmount = parseMemberImportAmount(memberImportCell(cells, headerMap, "youzan_amount"), &item, "有赞金额")
	return item
}

func memberImportRowFromRequest(rowIndex int, input requestbody.MemberImportItemRequest) MemberImportRow {
	return MemberImportRow{
		RowIndex:         rowIndex,
		Mobile:           strings.TrimSpace(input.Mobile),
		ManualUniqueCode: strings.TrimSpace(input.ManualUniqueCode),
		Nickname:         strings.TrimSpace(input.Nickname),
		TmallID:          strings.TrimSpace(input.TmallID),
		TmallAmount:      input.TmallAmount,
		YouzanID:         strings.TrimSpace(input.YouzanID),
		YouzanAmount:     input.YouzanAmount,
		Remarks:          strings.TrimSpace(input.Remarks),
	}
}

func memberImportCell(cells []string, headerMap map[string]int, key string) string {
	index, ok := headerMap[key]
	if !ok || index >= len(cells) {
		return ""
	}
	return strings.TrimSpace(cells[index])
}

func parseMemberImportAmount(value string, item *MemberImportRow, label string) float64 {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0
	}
	amount, err := strconv.ParseFloat(strings.ReplaceAll(value, ",", ""), 64)
	if err != nil || amount < 0 {
		item.Errors = append(item.Errors, label+"格式不正确")
		return 0
	}
	return amount
}

func validateMemberImportRow(item *MemberImportRow) {
	if item.Mobile == "" {
		item.Errors = append(item.Errors, "手机号不能为空")
	} else if err := validateMemberMobile(item.Mobile); err != nil {
		item.Errors = append(item.Errors, "手机号格式不正确")
	}
	if item.TmallAmount < 0 {
		item.Errors = append(item.Errors, "天猫金额不能小于0")
	}
	if item.YouzanAmount < 0 {
		item.Errors = append(item.Errors, "有赞金额不能小于0")
	}
}

func isEmptyMemberImportRow(item MemberImportRow) bool {
	return item.Mobile == "" &&
		item.ManualUniqueCode == "" &&
		item.Nickname == "" &&
		item.TmallID == "" &&
		item.TmallAmount == 0 &&
		item.YouzanID == "" &&
		item.YouzanAmount == 0 &&
		item.Remarks == ""
}

func markExistingMemberImportRows(items []MemberImportRow) error {
	mobiles := make([]string, 0, len(items))
	uniqueCodes := make([]string, 0, len(items))
	for index := range items {
		if len(items[index].Errors) > 0 {
			continue
		}
		mobiles = append(mobiles, items[index].Mobile)
		if items[index].ManualUniqueCode != "" {
			uniqueCodes = append(uniqueCodes, items[index].ManualUniqueCode)
		}
	}

	existingMobiles := map[string]bool{}
	if len(mobiles) > 0 {
		var values []string
		if err := db.DB.Model(&models.Member{}).Where("mobile IN ?", uniqueStrings(mobiles)).Pluck("mobile", &values).Error; err != nil {
			return err
		}
		for _, value := range values {
			existingMobiles[value] = true
		}
	}

	existingUniqueCodes := map[string]bool{}
	if len(uniqueCodes) > 0 {
		var values []string
		if err := db.DB.Model(&models.Member{}).Where("manual_unique_code IN ?", uniqueStrings(uniqueCodes)).Pluck("manual_unique_code", &values).Error; err != nil {
			return err
		}
		for _, value := range values {
			existingUniqueCodes[value] = true
		}
	}

	for index := range items {
		if existingMobiles[items[index].Mobile] {
			items[index].Errors = append(items[index].Errors, "手机号已存在")
		}
		if items[index].ManualUniqueCode != "" && existingUniqueCodes[items[index].ManualUniqueCode] {
			items[index].Errors = append(items[index].Errors, "唯一字段已存在")
		}
	}
	return nil
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func summarizeMemberImportRows(items []MemberImportRow) *MemberImportMatchResult {
	result := &MemberImportMatchResult{Items: items, TotalRows: len(items)}
	for index := range result.Items {
		result.Items[index].Matched = len(result.Items[index].Errors) == 0
		if result.Items[index].Matched {
			result.MatchedCount++
		} else {
			result.InvalidCount++
		}
	}
	return result
}
