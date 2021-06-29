package shared

import (
	"fmt"
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/go-resty/resty/v2"
	"net/http"
	"reflect"
	"testing"
)

func TestBaseClient(t *testing.T) {
	type args struct {
		config *configuration.Config
	}
	tests := []struct {
		name     string
		args     args
		selector func(client *resty.Client) interface{}
		want     interface{}
	}{
		{
			name: "Uses config redirects",
			args: args{config: &configuration.Config{
				Client: struct {
					Redirects int
					Retries   int
				}{Retries: 2},
			}},
			selector: func(client *resty.Client) interface{} {
				return client.RetryCount
			},
			want: 2,
		},
		{
			name: "Sets redirection function when redirects is passed",
			args: args{config: &configuration.Config{
				Client: struct {
					Redirects int
					Retries   int
				}{Redirects: 2},
			}},
			selector: func(client *resty.Client) interface{} {
				return client.GetClient().CheckRedirect != nil
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.selector(BaseClient(tt.args.config)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("BaseClient() has %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckResponse(t *testing.T) {
	type args struct {
		response *resty.Response
		err      error
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Error passed through correctly",
			args: args{
				response: nil,
				err:      fmt.Errorf("some error"),
			},
			wantErr: true,
		},
		{
			name: "Bad status is error",
			args: args{
				response: &resty.Response{
					Request: &resty.Request{
						URL: "URL",
					},
					RawResponse: &http.Response{
						StatusCode: 404,
					},
				},
				err: nil,
			},
			wantErr: true,
		},
		{
			name: "200-series status doesn't error",
			args: args{
				response: &resty.Response{
					Request: &resty.Request{
						URL: "URL",
					},
					RawResponse: &http.Response{
						StatusCode: 201,
					},
				},
				err: nil,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := CheckResponse(tt.args.response, tt.args.err); (err != nil) != tt.wantErr {
				t.Errorf("CheckResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
