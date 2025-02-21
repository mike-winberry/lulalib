package message_test

import (
	"bytes"
	"testing"

	"github.com/mike-winberry/lulalib/src/pkg/message"
	"github.com/stretchr/testify/assert"
)

func TestUseBuffer(t *testing.T) {
	message.DisableColor()
	message.NoProgress = true
	message.SetLogLevel(message.DebugLevel)

	var buf bytes.Buffer
	message.UseBuffer(&buf)

	message.Info("info msg")
	message.Debug("debug msg")
	message.Warn("warn msg")
	message.Success("success msg")
	message.Detail("detail msg")
	message.Fail("fail msg")
	message.Note("note msg")
	message.Printf("printf msg")
	message.Question("question msg")

	bufOut := buf.String()
	assert.Contains(t, bufOut, "INFO: info msg")
	assert.Contains(t, bufOut, "DEBUG: debug msg")
	assert.Contains(t, bufOut, "WARNING: warn msg")
	assert.Contains(t, bufOut, "SUCCESS: success msg")
	assert.Contains(t, bufOut, "DETAIL: detail msg")
	assert.Contains(t, bufOut, "FAIL: fail msg")
	assert.Contains(t, bufOut, "NOTE: note msg")
	assert.Contains(t, bufOut, "printf msg")
	assert.Contains(t, bufOut, "QUESTION: question msg")
}
