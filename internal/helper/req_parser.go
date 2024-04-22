package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/iancoleman/strcase"
)

type FlexInt int64

type IReqPayload interface {
	ParseJsonToStruct(interface{}, interface{}) (error, error)
	ValidateJson() error
}

type ReqContext struct {
	Payload IReqPayload
}

type FiberCtx struct {
	Fctx *fiber.Ctx
}

// ref: https://docs.bitnami.com/tutorials/dealing-with-json-with-non-homogeneous-types-in-go
func (fi *FlexInt) UnmarshalJSON(b []byte) error {
	if b[0] != '"' {
		return json.Unmarshal(b, (*int64)(fi))
	}
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*fi = FlexInt(i)
	return nil
}

func GetQueryString(queryString []byte) map[string]interface{} {
	decodedQuerystring, err := url.QueryUnescape(string(queryString))
	if err != nil {
		logger.Errorf("decodedQuerystring err: %+v", err)
	}
	logger.Debugf("decodedQuerystring: %+v", decodedQuerystring)
	params, err := url.ParseQuery(decodedQuerystring)
	if err != nil {
		log.Printf("ParseQuery err: %+v\n", err.Error())
	}

	var paramsMap = make(map[string]interface{}, 0)

	for key, value := range params {
		// fmt.Printf("-->  %v = %v\n", key, value)
		snakeCase := strcase.ToSnake(key)
		if strings.Contains(snakeCase, "date") || strings.Contains(snakeCase, "_at") {
			paramsMap["withDateFilter"] = true
		}

		if len(value) == 1 {
			paramsMap[snakeCase] = value[0]
			continue
		}
		paramsMap[snakeCase] = value
	}

	// fmt.Printf("paramsMap: %+v\n", paramsMap)
	return paramsMap
}

func (c *FiberCtx) ValidateJson() error {
	if !json.Valid(c.Fctx.BodyRaw()) {
		return fmt.Errorf("request JSON not valid...")
	}

	return nil
}

func (c *FiberCtx) ParseJsonToStruct(single interface{}, plural interface{}) (error, error) {
	singleErr := c.Fctx.BodyParser(single)
	pluralErr := c.Fctx.BodyParser(plural)

	if pluralErr != nil {
		log.Printf("pluralErr err: %+v\n", pluralErr.Error())
	}

	if singleErr != nil {
		log.Printf("singleErr err: %+v\n", singleErr.Error())
	}

	var allFailed error
	if singleErr != nil && pluralErr != nil {
		allFailed = errors.Join(fmt.Errorf("failed to parse given json into struct. "), singleErr, pluralErr)
	}

	return singleErr, allFailed
}
