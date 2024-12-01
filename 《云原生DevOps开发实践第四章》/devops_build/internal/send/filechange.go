package send

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"devops_build/config"
	"github.com/sirupsen/logrus"
	"go.elastic.co/apm"
)

func FileChange(ctx context.Context, keyName string) {
	span, ctx := apm.StartSpan(ctx, "FileChange", "func")
	defer span.End()

	data := Data{}
	// /nighting-build/Project/voice-call/configmap/release
	keys := strings.Split(keyName, "/")
	if len(keys) <= 5 {
		return
	}
	projectName := keys[3]
	channel := keys[5]
	if channel == "release" {
		return
	}

	id := projectName + strconv.FormatInt(time.Now().Unix(), 10)
	h := sha1.New()
	h.Write([]byte(id))
	id = hex.EncodeToString(h.Sum(nil))
	data.Id = id
	data.ProjectName = projectName
	data.Channel = channel
	data.EventType = "config_change"

	dataJson, err := json.Marshal(data)
	if err != nil {
		logrus.WithContext(ctx).Errorf("Data marshal in filechange failed: %v", err)
		return
	}

	logrus.WithContext(ctx).Infof("Send to release because of config changes: %s", dataJson)
	url := config.ApplicationConfig.NightingReleaseUrl
	client := &http.Client{}
	reader := strings.NewReader(string(dataJson))
	req, err := http.NewRequestWithContext(context.Background(), "POST", url, reader)
	if err != nil {
		logrus.WithContext(ctx).Errorf("Data marshal in filechange failed: %v", err)
		return
	}
	req.Close = true
	res, err := client.Do(req)
	if err != nil {
		logrus.WithContext(ctx).Errorf("Data marshal in filechange failed: %v", err)
		return
	}
	defer func() {
		err := res.Body.Close()
		if err != nil {
			logrus.WithContext(ctx).Errorf("Error: response body close failed, err: %s", err)
		}
	}()
}
