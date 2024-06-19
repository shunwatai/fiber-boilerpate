package logging

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"golang-api-starter/internal/auth"
	"golang-api-starter/internal/config"
	"golang-api-starter/internal/helper"
	"golang-api-starter/internal/helper/logger/zap_log"
	"golang-api-starter/internal/helper/utils"
	customLog "golang-api-starter/internal/modules/log"
	"golang-api-starter/internal/rabbitmq"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/gofiber/fiber/v2"
)

var cfg = config.Cfg

type Logger struct{}

/*
 * Log is a middleware for showing the http req & resp info
 */
func (l *Logger) Log() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// zlog.Printf("I AM LOGGER....")

		bodyBytes := c.BodyRaw()
		// log.Printf("1reqBody: %+v, %+v \n", len(string(bodyBytes)), string(bodyBytes))
		var reqBodyJson, respBodyJson *string
		if len(string(bodyBytes)) > 0 {
			if string(c.Response().Header.ContentType()) == "application/json" {
				reqBodyJson = utils.ToPtr(string(bodyBytes))
			} else {
				nonJsonMap := map[string]interface{}{}
				b64Str := base64.StdEncoding.EncodeToString(bodyBytes)
				nonJsonMap["requestType"] = string(c.Request().Header.ContentType())
				nonJsonMap["base64"] = b64Str
				if jsonBytes, err := json.Marshal(nonJsonMap); err != nil {
					logger.Errorf("failed to marshal nonJsonMap, err: %+v", err.Error())
				} else {
					reqBodyJson = utils.ToPtr(string(jsonBytes))
				}
			}
		}

		reqHeader, _ := json.Marshal(c.GetReqHeaders())
		// log.Printf("reqHeader: %+v \n", string(reqHeader))

		var userId interface{}
		claims, err := auth.ParseJwt(c.Get("Authorization"))
		if err == nil {
			userId = claims["userId"]
		}
		// log.Println("JWT userId:", userId)
		// log.Println("created by:", claims["userId"], claims["username"])

		start := time.Now()
		defer func() {
			/* write to database or send to monitor service */
			ip := c.IP()
			// log.Println("from IP:", ip)
			if len(string(c.Response().Body())) > 0 {
				if string(c.Response().Header.ContentType()) == "application/json" {
					respBodyJson = utils.ToPtr(string(c.Response().Body()))
				} else {
					nonJsonMap := map[string]interface{}{}
					b64Str := base64.StdEncoding.EncodeToString(c.Response().Body())
					nonJsonMap["responseType"] = string(c.Response().Header.ContentType())
					nonJsonMap["base64"] = b64Str
					if jsonBytes, err := json.Marshal(nonJsonMap); err != nil {
						logger.Errorf("failed to marshal nonJsonMap, err: %+v", err.Error())
					} else {
						respBodyJson = utils.ToPtr(string(jsonBytes))
					}
				}
			}

			/* insert into logs table */
			if slices.Contains(cfg.Logging.Type, "database") {
				logData := []*customLog.Log{{
					UserId:        userId,
					IpAddress:     ip,
					HttpMethod:    c.Method(),
					Route:         c.Request().URI().String(),
					UserAgent:     string(c.Request().Header.UserAgent()),
					RequestHeader: string(reqHeader),
					RequestBody:   reqBodyJson,
					ResponseBody:  respBodyJson,
					Status:        int64(c.Response().StatusCode()),
					Duration:      time.Since(start).Milliseconds(),
					CreatedAt:     &helper.CustomDatetime{&start, utils.ToPtr(time.RFC3339)},
				}}
				// log.Printf("%+v\n", logData)

				go QueueLog(logData...)

				// create log to database,
				// WARN: this will slower the performance as one more database operation
				// customLog.Srvc.Create(logData)
			}

			// create log to files
			if slices.Contains(cfg.Logging.Type, "zap") {
				logger.SysLog("FIBER REQ LOG",
					logger.GetField("UserId", userId),
					logger.GetField("IpAddress", ip),
					logger.GetField("HttpMethod", c.Method()),
					logger.GetField("Route", c.Request().URI().String()),
					logger.GetField("UserAgent", (c.Request().Header.UserAgent())),
					logger.GetField("RequestHeader", (reqHeader)),
					logger.GetField("RequestBody", reqBodyJson),
					logger.GetField("ResponseBody", respBodyJson),
					logger.GetField("Status", int64(c.Response().StatusCode())),
					logger.GetField("Duration", time.Since(start).Milliseconds()),
					logger.GetField("CreatedAt", &helper.CustomDatetime{&start, utils.ToPtr(time.RFC3339)}),
				)
			}
		}()

		return c.Next()
	}
}

func QueueLog(logs ...*customLog.Log) error {
	url:= rabbitmq.GetUrl()
	rabbitMQ, err := rabbitmq.NewRabbitMQ(url, "log_queue")
	if err != nil {
		return logger.Errorf(err.Error())
	}
	defer rabbitMQ.Close()

	for _, log := range logs {
		logDataBytes, err := json.Marshal(log)
		if err != nil {
			return logger.Errorf("failed to json marshal log, err: %+v", err)
		}

		if err := rabbitMQ.Publish(logDataBytes); err != nil {
			logger.Errorf("rabbit failed to publish error:", err)
		}
	}

	return err
}

// DecodeB64ToFormData for decode the base64 encoded multipart/form-data's raw baody.
// it is useless for now, put it here just in case we need to view the body from logs in the future.
func DecodeB64ToFormData(b64, reqContentType string) {
	/* SAMPLE CODE TO USE THIS DecodeB64ToFormData for convert base64's req Body back into multipart/form-data
		// var testMap map[string]interface{}
		// if err := json.Unmarshal([]byte(*reqBodyJson), &testMap); err != nil {
		// 	logger.Errorf("failed to unmarshal: %+v", err.Error())
		// }
		//
		// DecodeB64ToFormData(testMap["base64"].(string), testMap["requestType"].(string))
  */

	// decode the base64 string back to the original byte slice
	bodyBytes, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		logger.Errorf("err: %+v", err.Error())
		return
	}

	// create a new reader from the decoded byte slice
	reader := bytes.NewReader(bodyBytes)

	// create a new http.Request object
	mr := &http.Request{
		Header: make(http.Header),
		Body:   io.NopCloser(reader),
	}

	// set the Content-Type header to multipart/form-data
	mr.Header.Set("Content-Type", "multipart/form-data")
	// mr.Header.Set("Content-Type", reqContentType)

	// parse the multipart request
	err = mr.ParseMultipartForm(200 << 20) // 200MB max memory
	if err != nil {
		logger.Errorf("err: %+v", err.Error())
		return
	}

	// access the request values
	for key, values := range mr.Form {
		for _, value := range values {
			logger.Debugf("key: %s, value: %s\n", key, value)
		}
	}

	// access the uploaded files
	for _, file := range mr.MultipartForm.File {
		for _, f := range file {
			logger.Debugf("file: %s, size: %d\n", f.Filename, f.Size)
		}
	}
}
