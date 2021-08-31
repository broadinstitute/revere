package api

import (
	"encoding/json"
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/broadinstitute/revere/internal/version"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Squelch Gin's normal logging output in favor of test logs
var testConfig = configuration.Config{
	Api: struct {
		Port   int
		Debug  bool
		Silent bool
	}{Debug: false, Silent: true},
}

// Alias the abstract fields needed to test a route
type routeTest = struct {
	name      string
	reqMethod string
	reqUrl    string
	reqBody   io.Reader
	wantCode  int
	wantJson  interface{}
}

func Test_getStatus(t *testing.T) {
	tests := []routeTest{
		{
			name:      "Status returns static output",
			reqMethod: "GET",
			reqUrl:    "/status",
			reqBody:   nil,
			wantCode:  200,
			wantJson:  gin.H{"status": "ok"},
		},
		{
			name:      "Status is available at api root",
			reqMethod: "GET",
			reqUrl:    "/api/v1/status",
			reqBody:   nil,
			wantCode:  200,
			wantJson:  gin.H{"status": "ok"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderedWantJson, err := json.Marshal(tt.wantJson)
			if err != nil {
				t.Errorf("wantJson %v could not be rendered: %v", tt.wantJson, err)
				return
			}
			router := NewRouter(&testConfig)
			got := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.reqMethod, tt.reqUrl, tt.reqBody)
			router.ServeHTTP(got, req)
			if got.Code != tt.wantCode {
				t.Errorf("%s %s -> code %d, want %d", tt.reqMethod, tt.reqUrl, got.Code, tt.wantCode)
			}
			if got.Body.String() != string(renderedWantJson) {
				t.Errorf("%s %s -> body %v, want %v", tt.reqMethod, tt.reqUrl, got.Body.String(), string(renderedWantJson))
			}
		})
	}
}

func Test_getVersion(t *testing.T) {
	tests := []routeTest{
		{
			name:      "Version returns stored version",
			reqMethod: "GET",
			reqUrl:    "/version",
			reqBody:   nil,
			wantCode:  200,
			wantJson:  gin.H{"version": version.BuildVersion},
		},
		{
			name:      "Version is available at api root",
			reqMethod: "GET",
			reqUrl:    "/api/v1/version",
			reqBody:   nil,
			wantCode:  200,
			wantJson:  gin.H{"version": version.BuildVersion},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			renderedWantJson, err := json.Marshal(tt.wantJson)
			if err != nil {
				t.Errorf("wantJson %v could not be rendered: %v", tt.wantJson, err)
				return
			}
			router := NewRouter(&testConfig)
			got := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.reqMethod, tt.reqUrl, tt.reqBody)
			router.ServeHTTP(got, req)
			if got.Code != tt.wantCode {
				t.Errorf("%s %s -> code %d, want %d", tt.reqMethod, tt.reqUrl, got.Code, tt.wantCode)
			}
			if got.Body.String() != string(renderedWantJson) {
				t.Errorf("%s %s -> body %v, want %v", tt.reqMethod, tt.reqUrl, got.Body.String(), string(renderedWantJson))
			}
		})
	}
}
