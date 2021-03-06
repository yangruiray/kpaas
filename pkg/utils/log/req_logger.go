// Copyright 2019 Shanghai JingDuo Information Technology co., Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"io"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/kpaas-io/kpaas/pkg/constant"
)

func genReqId() string {
	var b [12]byte
	io.ReadFull(rand.Reader, b[:])
	return base64.URLEncoding.EncodeToString(b[:])
}

func ReqLoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		logrus.SetFormatter(&logrus.TextFormatter{TimestampFormat: time.RFC3339})

		reqId := c.Request.Header.Get(constant.RequestID)
		if reqId == "" {
			reqId = genReqId()
			c.Request.Header.Set(constant.RequestID, reqId)
		}
		c.Set(constant.RequestID, reqId)
		// Set request id into response header
		c.Writer.Header().Set(constant.RequestID, reqId)

		c.Next()

		end := time.Now()
		latency := end.Sub(start)

		entry := logrus.WithFields(logrus.Fields{
			"reqId":      reqId,
			"status":     c.Writer.Status(),
			"method":     c.Request.Method,
			"path":       c.Request.URL,
			"size":       c.Writer.Size(),
			"ip":         c.ClientIP(),
			"latency":    latency,
			"user-agent": c.Request.UserAgent(),
		})

		if len(c.Errors) > 0 {
			entry.Info(c.Errors.String())
		} else {
			entry.Info()
		}
	}
}

// usage: ReqEntry(c).Debug(".....")
func ReqEntry(c context.Context) *logrus.Entry {
	reqId, _ := c.Value(constant.RequestID).(string)
	return logrus.WithField("reqId", reqId)
}
