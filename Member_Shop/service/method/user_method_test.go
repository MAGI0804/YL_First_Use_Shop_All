package method

import (
	"Member_shop/models"
	"testing"
)

func TestValidateMemberWechatLinkRejectsDisabledMember(t *testing.T) {
	member := models.Member{Mobile: "13800138000", Status: "disabled"}
	user := models.User{UserID: 10, Mobile: "13800138000", OpenID: "openid-a"}

	if err := validateMemberWechatLink(member, user, "openid-a", "13800138000"); err == nil {
		t.Fatalf("expected disabled member to be rejected")
	}
}

func TestValidateMemberWechatLinkRejectsDifferentWechatUser(t *testing.T) {
	member := models.Member{UserID: 10, Mobile: "13800138000", OpenID: "openid-a", Status: "active"}
	user := models.User{UserID: 10, Mobile: "13800138000", OpenID: "openid-a"}

	if err := validateMemberWechatLink(member, user, "openid-b", "13800138000"); err == nil {
		t.Fatalf("expected different openid to be rejected")
	}
}

func TestValidateMemberWechatLinkAllowsSameMemberUserAndMobile(t *testing.T) {
	member := models.Member{UserID: 10, Mobile: "13800138000", OpenID: "openid-a", Status: "active"}
	user := models.User{UserID: 10, Mobile: "13800138000", OpenID: "openid-a"}

	if err := validateMemberWechatLink(member, user, "openid-a", "13800138000"); err != nil {
		t.Fatalf("expected valid member link, got %v", err)
	}
}

func TestShouldUpdateWechatNicknameOnlyReplacesDefaults(t *testing.T) {
	if !shouldUpdateWechatNickname("微信用户_openid", "幼岚会员") {
		t.Fatalf("default nickname should be replaceable")
	}
	if shouldUpdateWechatNickname("已设置昵称", "幼岚会员") {
		t.Fatalf("custom nickname should not be overwritten")
	}
}

func TestNormalizedWechatAvatarIgnoresTemporaryMiniProgramPath(t *testing.T) {
	if got := normalizedWechatAvatar("wxfile://tmp/avatar.jpg"); got != "" {
		t.Fatalf("temporary wxfile avatar should not be persisted, got %q", got)
	}
	if got := normalizedWechatAvatar("https://example.com/avatar.jpg"); got != "https://example.com/avatar.jpg" {
		t.Fatalf("http avatar should be kept, got %q", got)
	}
}
