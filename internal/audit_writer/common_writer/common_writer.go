package commonwriter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/RichardoC/kube-audit-rest/internal/common"
	"github.com/tidwall/sjson"
)

func LogEvent(body []byte, writer io.Writer) {
	requestStr := string(body)
	updatedObj, err := addTimestamp(requestStr)
	if err != nil {
		common.Logger.Debugw("failed to add timestamp", "error", err)
	}

	// Compact the json for single line use regardless of request prettiness
	dst := &bytes.Buffer{}
	json.Compact(dst, []byte(updatedObj))

	_, err = fmt.Fprintln(writer, dst)
	if err != nil {
		common.Logger.Error(err)
	}
}

func addTimestamp(requestBody string) (string, error) {
	currentTime := time.Now().Format(time.RFC3339Nano)
	return sjson.Set(requestBody, "requestReceivedTimestamp", currentTime)
}
