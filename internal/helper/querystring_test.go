package helper

import (
	"log"
	"reflect"
	"testing"
)

type orderByTests struct {
	name  string
	input string
	want  map[string]string
}

func TestParseOrderBy(t *testing.T) {
	cfg.LoadEnvVariables()
	cfg.Vpr.Set("database.engine", "postgres")
	if err := cfg.Vpr.Unmarshal(cfg); err != nil {
		log.Printf("failed loading conf, err: %+v\n", err.Error())
	}
	tests := []orderByTests{
		{name: "orderBy is empty string", input: "", want: map[string]string{"key": "id", "by": "desc"}},
		{name: "orderBy is name.asc", input: "name.asc", want: map[string]string{"key": "name", "by": "asc"}},
		{name: "orderBy is name", input: "name", want: map[string]string{"key": "name", "by": "desc"}},
		{name: "orderBy is id.asc", input: "id.asc", want: map[string]string{"key": "id", "by": "asc"}},
		{name: "orderBy is userId.asc", input: "userId.asc", want: map[string]string{"key": "user_id", "by": "asc"}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := parseOrderBy(testCase.input)

			eq := reflect.DeepEqual(testCase.want, got)

			if !eq {
				t.Errorf("got %q want %q", got, testCase.want)
			}
		})
	}
}

func TestMongoParseOrderBy(t *testing.T) {
	cfg.LoadEnvVariables()
	cfg.Vpr.Set("database.engine", "mongodb")
	if err := cfg.Vpr.Unmarshal(cfg); err != nil {
		log.Printf("failed loading conf, err: %+v\n", err.Error())
	}
	tests := []orderByTests{
		{name: "orderBy is empty string", input: "", want: map[string]string{"key": "createdAt", "by": "desc"}},
		{name: "orderBy is name.asc", input: "name.asc", want: map[string]string{"key": "name", "by": "asc"}},
		{name: "orderBy is name", input: "name", want: map[string]string{"key": "name", "by": "desc"}},
		{name: "orderBy is userId.asc", input: "userId.asc", want: map[string]string{"key": "user_id", "by": "asc"}},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := parseOrderBy(testCase.input)

			eq := reflect.DeepEqual(testCase.want, got)

			if !eq {
				t.Errorf("got %q want %q", got, testCase.want)
			}
		})
	}
}
