package main

import (
	"bufio"
	"strings"
	"testing"
)

var testNginxStandardStatus = `Active connections: 291
server accepts handled requests
 16630948 16630948 31070465
Reading: 6 Writing: 179 Waiting: 106
`
var testBadNginxStandardStatus = `Active connections: 291
server accepts handled requests
this is an extra line that makes the parser fail
 16630948 16630948 31070465
Reading: 6 Writing: 179 Waiting: 106
`

var testNginxPlusStatus = `{
  "timestamp": 1490347905131,
  "connections": {
    "accepted": 4112716,
    "dropped": 0,
    "active": 6,
    "idle": 41
  },
  "requests": {
    "total": 9353067,
    "current": 5
  },
  "nginx_version": "1.0"
}
`
var testBadNginxPlusStatus = `{`

func TestGetPlusMetrics(t *testing.T) {
	rawMetrics, err := getPlusMetrics(bufio.NewReader(strings.NewReader(testNginxPlusStatus)))
	if err != nil {
		t.Fatal()
	}

	if len(rawMetrics) != 8 {
		t.Error()
	}
	if rawMetrics["connections.accepted"] != 4112716 {
		t.Error()
	}
	if rawMetrics["connections.dropped"] != 0 {
		t.Error()
	}
	if rawMetrics["connections.active"] != 6 {
		t.Error()
	}
	if rawMetrics["connections.idle"] != 41 {
		t.Error()
	}
	if rawMetrics["requests.total"] != 9353067 {
		t.Error()
	}
	if rawMetrics["version"] != "1.0" {
		t.Error()
	}
	if rawMetrics["edition"] != "plus" {
		t.Error()
	}
}

func TestGetPlusMetricsWithInvalidData(t *testing.T) {
	rawMetrics, err := getPlusMetrics(bufio.NewReader(strings.NewReader(testBadNginxPlusStatus)))

	if rawMetrics != nil {
		t.Error()
	}
	if err == nil {
		t.Error()
	}
}

func TestGetStandardMetrics(t *testing.T) {
	rawMetrics, err := getStandardMetrics(bufio.NewReader(strings.NewReader(testNginxStandardStatus)))
	if err != nil {
		t.Fatal()
	}
	if len(rawMetrics) != 9 {
		t.Error()
	}
	if rawMetrics["active"] != 291 {
		t.Error()
	}
	if rawMetrics["reading"] != 6 {
		t.Error()
	}
	if rawMetrics["waiting"] != 106 {
		t.Error()
	}
	if rawMetrics["writing"] != 179 {
		t.Error()
	}
	if rawMetrics["requests"] != 31070465 {
		t.Error()
	}
	if rawMetrics["accepted"] != 16630948 {
		t.Error()
	}
	if rawMetrics["handled"] != 16630948 {
		t.Error()
	}
}

func TestGetStandardMetricsWithInvalidData(t *testing.T) {
	rawMetrics, err := getStandardMetrics(bufio.NewReader(strings.NewReader(testBadNginxStandardStatus)))

	if rawMetrics != nil {
		t.Error()
	}
	if err == nil {
		t.Error()
	}
}

func Test_pathToPrefix(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		prefix string
	}{
		{"Single", "nginx", "nginx."},
		{"Single, prefix", "/nginx", "nginx."},
		{"Single, suffix", "nginx/", "nginx."},
		{"Single,  bracketed", "/nginx/", "nginx."},
		{"Multi", "nginx/version", "nginx.version."},
		{"Multi, prefix", "/nginx/version", "nginx.version."},
		{"Multi, suffix", "nginx/version/", "nginx.version."},
		{"Multi, bracketed", "/nginx/version/", "nginx.version."},
		{"Empty", "", ""},
		{"Single slash", "/", ""},
		{"Double slash", "//", ""},
		{"Triple slash", "///", ""},
		{"Quad slash", "////", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotPrefix := pathToPrefix(tt.path); gotPrefix != tt.prefix {
				t.Errorf("pathToPrefix() = %v, want %v", gotPrefix, tt.prefix)
			}
		})
	}
}
