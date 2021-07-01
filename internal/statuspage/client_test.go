package statuspage

import (
	"github.com/broadinstitute/revere/internal/configuration"
	"github.com/go-resty/resty/v2"
	"reflect"
	"testing"
)

func TestClient(t *testing.T) {
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
			name: "Sets OAuth Scheme",
			args: args{
				config: &configuration.Config{
					Statuspage: struct {
						ApiKey     string `validate:"required"`
						PageID     string `validate:"required"`
						ApiRoot    string
						Components []configuration.Component
					}{},
				},
			},
			selector: func(client *resty.Client) interface{} {
				return client.AuthScheme
			},
			want: "OAuth",
		},
		{
			name: "Sets OAuth Key",
			args: args{
				config: &configuration.Config{
					Statuspage: struct {
						ApiKey     string `validate:"required"`
						PageID     string `validate:"required"`
						ApiRoot    string
						Components []configuration.Component
					}{ApiKey: "foo"},
				},
			},
			selector: func(client *resty.Client) interface{} {
				return client.Token
			},
			want: "foo",
		},
		{
			name: "Sets API Root",
			args: args{
				config: &configuration.Config{
					Statuspage: struct {
						ApiKey     string `validate:"required"`
						PageID     string `validate:"required"`
						ApiRoot    string
						Components []configuration.Component
					}{ApiRoot: "https://example.com"},
				},
			},
			selector: func(client *resty.Client) interface{} {
				return client.HostURL
			},
			want: "https://example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.selector(Client(tt.args.config)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Client() = %v, want %v", got, tt.want)
			}
		})
	}
}
