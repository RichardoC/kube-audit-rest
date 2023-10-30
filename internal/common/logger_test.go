package common_test

import (
	"testing"

	"github.com/RichardoC/kube-audit-rest/internal/common"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"
)

func TestConfigGlobalLogger(t *testing.T) {
	testCases := []struct {
		name          string
		desiredLevel  common.LogCfg
		expectedLevel zapcore.Level
	}{
		{"dbg", common.Dbg, zapcore.DebugLevel},
		{"prod", common.Prod, zapcore.InfoLevel},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			common.ConfigGlobalLogger(tc.desiredLevel)
			assert.Equal(t, common.Logger.Level(), tc.expectedLevel)
		})
	}
}
