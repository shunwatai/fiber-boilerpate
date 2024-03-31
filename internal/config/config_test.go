package config

import (
	"fmt"
	"reflect"
	"testing"
)

type configTests struct {
	name  string
	setConfig func() Config
	want  string
}

var cfg = Cfg

func setServerWithHostPort() Config {
	cfg.LoadEnvVariables()
	cfg.Vpr.Set("server.host", "localhost")
	cfg.Vpr.Set("server.port", "2345")
	if err := cfg.Vpr.Unmarshal(cfg); err != nil {
		fmt.Printf("failed loading conf, err: %+v\n", err.Error())
	}

	return *cfg
}

func setServerWithHost() Config {
	cfg.LoadEnvVariables()
	cfg.Vpr.Set("server.host", "tld.com.hk")
	cfg.Vpr.Set("server.port", "")
	if err := cfg.Vpr.Unmarshal(cfg); err != nil {
		fmt.Printf("failed loading conf, err: %+v\n", err.Error())
	}

	return *cfg
}

func TestGetServerUrl(t *testing.T) {
	tests := []configTests{
		{name: "test server url for both host & port", setConfig: setServerWithHostPort, want: "http://localhost:2345"},
		{name: "test server url with host only", setConfig: setServerWithHost, want: "http://tld.com.hk"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testCase.setConfig()
			got := cfg.GetServerUrl()

			eq := reflect.DeepEqual(testCase.want, got)

			if !eq {
				t.Errorf("got %q want %q", got, testCase.want)
			}
		})
	}
}
