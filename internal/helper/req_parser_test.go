package helper

import (
	zlog "golang-api-starter/internal/helper/logger/zap_log"
	"reflect"
	"testing"
)

type getQueryStrTest struct {
	name  string
	input string
	want  map[string]interface{}
}

func TestGetQueryString(t *testing.T) {
	cfg.LoadEnvVariables()
	zlog.NewZlog()
	tests := []getQueryStrTest{
		{name: "test nothing", input: "", want: map[string]interface{}{}},
		{name: "test normal", input: "id=1&name=something", want: map[string]interface{}{"id": "1", "name": "something"}},
		{name: "test list []", input: "id=1&name=something&name=nothing", want: map[string]interface{}{"id": "1", "name": []string{"something", "nothing"}}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := GetQueryString([]byte(testCase.input))

			eq := reflect.DeepEqual(testCase.want, got)

			if !eq {
				t.Errorf("got %q want %q", got, testCase.want)
			}
		})
	}
}
