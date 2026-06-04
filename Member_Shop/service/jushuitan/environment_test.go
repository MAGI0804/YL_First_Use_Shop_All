package jushuitan

import "testing"

func TestNormalizeJushuitanStage(t *testing.T) {
	tests := []struct {
		name  string
		stage string
		want  string
	}{
		{name: "empty defaults to test", stage: "", want: jushuitanStageTest},
		{name: "test uses test", stage: "test", want: jushuitanStageTest},
		{name: "develop uses test", stage: "develop", want: jushuitanStageTest},
		{name: "prod uses prod", stage: "prod", want: jushuitanStageProd},
		{name: "formal uses prod", stage: "formal", want: jushuitanStageProd},
		{name: "unknown defaults to test", stage: "preview", want: jushuitanStageTest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeJushuitanStage(tt.stage); got != tt.want {
				t.Fatalf("normalizeJushuitanStage(%q)=%q, want %q", tt.stage, got, tt.want)
			}
		})
	}
}
