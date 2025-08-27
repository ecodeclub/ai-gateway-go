package template

import (
	"context"
	"testing"
)

func TestFCall(t *testing.T) {
	ctx := NewContext(context.Background())
	if err := ctx.SetVariable(NewVariable("level", "advance")); err != nil {
		t.Fatalf("设置变量失败: %v", err)
	}

	r := NewDefaultRender(DefaultConfig())
	// 使用内置 replace(old, new, s) 函数，把模板中的占位符 ${level} 替换为变量 .level
	tpl := `{{ replace "${level}" .level "{{ .name.${level} }}" }}`

	out, err := r.Render(ctx, tpl)
	if err != nil {
		t.Fatalf("渲染失败: %v", err)
	}

	expected := "{{ .name.advance }}"
	if out != expected {
		t.Fatalf("期望 %q, 实际 %q", expected, out)
	}
}
