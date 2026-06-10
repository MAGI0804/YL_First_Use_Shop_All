package method

import "testing"

func TestUniqueUintRemovesZeroAndDuplicates(t *testing.T) {
	got := uniqueUint([]uint{2, 0, 2, 5, 5})
	if len(got) != 2 || got[0] != 2 || got[1] != 5 {
		t.Fatalf("uniqueUint returned %v", got)
	}
}

func TestNormalizeMemberDefaults(t *testing.T) {
	if normalizeMemberStatus("") != "active" {
		t.Fatalf("empty status should default active")
	}
	if normalizeMemberSource("") != "backend" {
		t.Fatalf("empty source should default backend")
	}
}

func TestNormalizeMemberOpenIDTreatsZeroAsEmpty(t *testing.T) {
	if got := normalizeMemberOpenID("0"); got != "" {
		t.Fatalf("zero placeholder should be normalized to empty, got %q", got)
	}
	if got := normalizeMemberOpenID(" openid-a "); got != "openid-a" {
		t.Fatalf("real openid should be trimmed and kept, got %q", got)
	}
}

func TestBuildMemberImportRowMapsChineseHeaders(t *testing.T) {
	headers := buildMemberImportHeaderMap([]string{"手机号(必填)", "唯一字段", "昵称", "天猫金额", "有赞金额", "备注"})
	row := buildMemberImportRow(2, []string{"13800138000", "VIP001", "张三", "12.30", "45.60", "首批导入"}, headers)
	if row.Mobile != "13800138000" || row.ManualUniqueCode != "VIP001" || row.Nickname != "张三" {
		t.Fatalf("row basic fields not mapped: %+v", row)
	}
	if row.TmallAmount != 12.30 || row.YouzanAmount != 45.60 {
		t.Fatalf("row amounts not parsed: %+v", row)
	}
	if len(row.Errors) != 0 {
		t.Fatalf("expected no parse errors, got %v", row.Errors)
	}
}

func TestBuildMemberImportRowRejectsInvalidAmount(t *testing.T) {
	headers := buildMemberImportHeaderMap([]string{"手机号", "天猫金额"})
	row := buildMemberImportRow(2, []string{"13800138000", "abc"}, headers)
	if len(row.Errors) != 1 || row.Errors[0] != "天猫金额格式不正确" {
		t.Fatalf("expected invalid amount error, got %v", row.Errors)
	}
}

func TestIsEmptyMemberImportRow(t *testing.T) {
	if !isEmptyMemberImportRow(MemberImportRow{}) {
		t.Fatalf("zero value row should be empty")
	}
	if isEmptyMemberImportRow(MemberImportRow{Mobile: "13800138000"}) {
		t.Fatalf("row with mobile should not be empty")
	}
}

func TestUniqueStringsTrimsAndRemovesDuplicates(t *testing.T) {
	got := uniqueStrings([]string{" a ", "", "a", "b"})
	if len(got) != 2 || got[0] != "a" || got[1] != "b" {
		t.Fatalf("uniqueStrings returned %v", got)
	}
}
