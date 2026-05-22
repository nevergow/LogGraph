package parser

import (
	"testing"

	"loggraph/internal/model"
)

func TestParse_Completed(t *testing.T) {
	r := Parse("~~完成了这项测试~~")
	if r.Status != model.StatusCompleted {
		t.Fatalf("expected completed, got %s", r.Status)
	}
}

func TestParse_Blocked(t *testing.T) {
	r := Parse("[BLOCK] 等待设备到位 ~~部分完成~~")
	if r.Status != model.StatusBlocked {
		t.Fatalf("expected blocked, got %s", r.Status)
	}
}

func TestParse_Active(t *testing.T) {
	r := Parse("今天开始进行 #GB38031 测试")
	if r.Status != model.StatusActive {
		t.Fatalf("expected active, got %s", r.Status)
	}
}

func TestParse_Tags(t *testing.T) {
	r := Parse("#项目A #GB38031 的测试数据已上传")
	if len(r.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d: %v", len(r.Tags), r.Tags)
	}
	if r.Tags[0] != "项目A" || r.Tags[1] != "GB38031" {
		t.Fatalf("unexpected tags: %v", r.Tags)
	}
}

func TestParse_Mentions(t *testing.T) {
	r := Parse("@张三 @李四 请确认 #项目A 进度")
	if len(r.Mentions) != 2 {
		t.Fatalf("expected 2 mentions, got %d: %v", len(r.Mentions), r.Mentions)
	}
}

func TestParse_References(t *testing.T) {
	r := Parse("参考 ^a1b2c3d4-e5f6-7890-abcd-ef1234567890 的测试方案")
	if len(r.References) != 1 {
		t.Fatalf("expected 1 reference, got %d: %v", len(r.References), r.References)
	}
}

func TestParse_Combined(t *testing.T) {
	r := Parse("#GB38031 ~~振动测试通过~~ @王工 ^a1b2c3d4-e5f6-7890-abcd-ef1234567890")
	if r.Status != model.StatusCompleted {
		t.Fatalf("expected completed, got %s", r.Status)
	}
	if len(r.Tags) != 1 || r.Tags[0] != "GB38031" {
		t.Fatalf("unexpected tags: %v", r.Tags)
	}
	if len(r.Mentions) != 1 || r.Mentions[0] != "王工" {
		t.Fatalf("unexpected mentions: %v", r.Mentions)
	}
	if len(r.References) != 1 {
		t.Fatalf("unexpected references: %v", r.References)
	}
}
