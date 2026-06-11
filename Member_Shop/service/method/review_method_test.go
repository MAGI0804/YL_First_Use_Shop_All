package method

import (
	"Member_shop/models"
	"mime/multipart"
	"net/textproto"
	"strings"
	"testing"
)

func TestNormalizeReviewCreateInput(t *testing.T) {
	input := ReviewCreateInput{
		Content: "  面料舒服，尺码也合适  ",
		Images: []string{
			" https://cdn.example.com/reviews/a.jpg ",
			"/media/reviews/b.png",
			"",
		},
		Tags: []string{"质量好", "质量好", "发货快", ""},
	}

	if err := normalizeReviewCreateInput(&input); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if input.Content != "面料舒服，尺码也合适" {
		t.Fatalf("content was not normalized: %q", input.Content)
	}
	if len(input.Images) != 2 {
		t.Fatalf("images length = %d, want 2", len(input.Images))
	}
	if len(input.Tags) != 2 {
		t.Fatalf("tags length = %d, want 2", len(input.Tags))
	}
}

func TestNormalizeReviewCreateInputRejectsInvalidContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "empty", content: "  "},
		{name: "too long", content: strings.Repeat("好", maxReviewContentLength+1)},
		{name: "script", content: "<script>alert(1)</script>"},
		{name: "javascript url", content: "javascript:alert(1)"},
		{name: "sensitive word", content: "这是诈骗信息"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := ReviewCreateInput{Content: tt.content}
			if err := normalizeReviewCreateInput(&input); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestNormalizeReviewImagesRejectsInvalidURLs(t *testing.T) {
	tests := []struct {
		name   string
		images []string
	}{
		{name: "too many", images: []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}},
		{name: "data url", images: []string{"data:image/png;base64,abc"}},
		{name: "missing host", images: []string{"https:///a.jpg"}},
		{name: "too long", images: []string{"https://example.com/" + strings.Repeat("a", maxReviewImageURLLen)}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := normalizeReviewImages(tt.images); err == nil {
				t.Fatalf("expected error")
			}
		})
	}
}

func TestNormalizeReviewTagsRejectsUnknownTag(t *testing.T) {
	if _, err := normalizeReviewTags([]string{"未知标签"}); err == nil {
		t.Fatalf("expected error")
	}
}

func TestParseStoredReviewStringList(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  []string
	}{
		{name: "json array", value: `["质量好"," 发货快 ",""]`, want: []string{"质量好", "发货快"}},
		{name: "comma separated", value: "质量好, 发货快, ", want: []string{"质量好", "发货快"}},
		{name: "empty", value: "   ", want: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseStoredReviewStringList(tt.value)
			if len(got) != len(tt.want) {
				t.Fatalf("len = %d, want %d: %#v", len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Fatalf("item %d = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestEnsureReviewUserCanMutate(t *testing.T) {
	tests := []struct {
		name    string
		review  models.ProductReview
		userID  int
		wantErr bool
	}{
		{
			name:    "owner can mutate pending review",
			review:  models.ProductReview{UserID: 7, Status: models.ReviewStatusPending},
			userID:  7,
			wantErr: false,
		},
		{
			name:    "rejects other user",
			review:  models.ProductReview{UserID: 7, Status: models.ReviewStatusPending},
			userID:  8,
			wantErr: true,
		},
		{
			name:    "rejects approved review",
			review:  models.ProductReview{UserID: 7, Status: models.ReviewStatusApproved},
			userID:  7,
			wantErr: true,
		},
		{
			name:    "rejects hidden review",
			review:  models.ProductReview{UserID: 7, Status: models.ReviewStatusHidden},
			userID:  7,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ensureReviewUserCanMutate(tt.review, tt.userID)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestValidateReviewImageUpload(t *testing.T) {
	tests := []struct {
		name    string
		file    *multipart.FileHeader
		wantErr bool
	}{
		{
			name: "valid jpeg",
			file: reviewUploadFileHeader("review.jpg", "image/jpeg", 1024),
		},
		{
			name: "valid png",
			file: reviewUploadFileHeader("review.png", "image/png", 1024),
		},
		{
			name:    "nil file",
			file:    nil,
			wantErr: true,
		},
		{
			name:    "empty file",
			file:    reviewUploadFileHeader("review.jpg", "image/jpeg", 0),
			wantErr: true,
		},
		{
			name:    "too large",
			file:    reviewUploadFileHeader("review.jpg", "image/jpeg", MaxReviewUploadImageSize+1),
			wantErr: true,
		},
		{
			name:    "unsupported extension",
			file:    reviewUploadFileHeader("review.svg", "image/svg+xml", 1024),
			wantErr: true,
		},
		{
			name:    "content type mismatch",
			file:    reviewUploadFileHeader("review.jpg", "image/png", 1024),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateReviewImageUpload(tt.file)
			if tt.wantErr && err == nil {
				t.Fatalf("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func reviewUploadFileHeader(filename, contentType string, size int64) *multipart.FileHeader {
	header := textproto.MIMEHeader{}
	if contentType != "" {
		header.Set("Content-Type", contentType)
	}
	return &multipart.FileHeader{
		Filename: filename,
		Header:   header,
		Size:     size,
	}
}
