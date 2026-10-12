package adapter_test

import (
	"testing"

	"github.com/zeshanziya/mcputils/internal/adapter"
	"github.com/zeshanziya/mcputils/internal/adapter/claudecode"
)

func TestInterfaceCompileCheck(t *testing.T) {
	var _ adapter.ClientAdapter = (*claudecode.Adapter)(nil)
}
