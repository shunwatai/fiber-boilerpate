package helper

import (
	"fmt"
	logger "golang-api-starter/internal/helper/logger/zap_log"
	"strconv"
)

func ConvertNumberSliceToString(i interface{}) ([]string, error) {
	stringSlice := []string{}

	switch i.(type) {
	case []interface{}:
		for _, v := range i.([]interface{}) {
			stringSlice = append(stringSlice, strconv.Itoa(int(v.(int))))
		}
	case []int64:
		for _, v := range i.([]int64) {
			stringSlice = append(stringSlice, strconv.Itoa(int(v)))
		}
	case []int:
		for _, v := range i.([]int) {
			stringSlice = append(stringSlice, strconv.Itoa(int(v)))
		}
	case []float64:
		for _, v := range i.([]float64) {
			stringSlice = append(stringSlice, strconv.Itoa(int(v)))
		}
	case []FlexInt:
		for _, v := range i.([]FlexInt) {
			stringSlice = append(stringSlice, strconv.Itoa(int(v)))
		}
	default:
		msg := "failed to handle %+v (%T)\n"
		logger.Errorf(msg, i, i)
		return nil, fmt.Errorf(msg, i, i)
	}

	return stringSlice, nil
}

func ConvertStringToInt(numstr string) (int64, error) {
	var (
		num int64
		err error
	)
	if num, err = strconv.ParseInt(numstr, 10, 64); err != nil {
		logger.Errorf("ConvertStringToInt err:  %+v", err)
		return num, err
	}
	return num, nil
}
