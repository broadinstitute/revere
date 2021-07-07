package configuration

import (
	"github.com/spf13/viper"
	"os"
	"reflect"
	"testing"
)

func TestAssembleConfig(t *testing.T) {
	tests := []struct {
		name           string
		configureViper func(v *viper.Viper)
		want           *Config
		wantErr        bool
	}{
		{
			name:           "Errors on empty config",
			configureViper: func(v *viper.Viper) {},
			want:           nil,
			wantErr:        true,
		},
		{
			name: "Correctly parses minimal config",
			configureViper: func(v *viper.Viper) {
				v.Set("Statuspage.ApiKey", "foo")
				v.Set("Statuspage.PageID", "bar")
			},
			want: &Config{
				Verbose: false,
				Client: struct {
					Redirects int
					Retries   int
				}{
					Redirects: 3,
					Retries:   3,
				},
				Statuspage: struct {
					ApiKey     string `validate:"required"`
					PageID     string `validate:"required"`
					ApiRoot    string
					Components []Component
					Groups     []ComponentGroup
				}{
					ApiKey:  "foo",
					PageID:  "bar",
					ApiRoot: "https://api.statuspage.io/v1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			tt.configureViper(v)
			got, err := AssembleConfig(v)
			if (err != nil) != tt.wantErr {
				t.Errorf("AssembleConfig() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AssembleConfig() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_newDefaultConfig(t *testing.T) {
	tests := []struct {
		name string
		want *Config
	}{
		{
			name: "Default config values",
			want: &Config{
				Verbose: false,
				Client: struct {
					Redirects int
					Retries   int
				}{
					Redirects: 3,
					Retries:   3,
				},
				Statuspage: struct {
					ApiKey     string `validate:"required"`
					PageID     string `validate:"required"`
					ApiRoot    string
					Components []Component
					Groups     []ComponentGroup
				}{
					ApiRoot: "https://api.statuspage.io/v1",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := newDefaultConfig(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("newDefaultConfig() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_readEnvironmentVariables(t *testing.T) {
	envVal := "foobar"
	type args struct {
		config *Config
	}
	tests := []struct {
		name         string
		args         args
		envKey       string
		configAccess func(config *Config) string
	}{
		{
			name:   "Reads Statuspage API Key",
			args:   args{config: &Config{}},
			envKey: "REVERE_STATUSPAGE_APIKEY",
			configAccess: func(config *Config) string {
				return config.Statuspage.ApiKey
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existingVal, present := os.LookupEnv(tt.envKey)
			err := os.Setenv(tt.envKey, envVal)
			if err != nil {
				t.Errorf("env error setting %v", err)
			}
			readEnvironmentVariables(tt.args.config)
			if got := tt.configAccess(tt.args.config); got != envVal {
				t.Errorf("readEnvironmentVariables() got %s for %s, want %s", got, tt.envKey, envVal)
			}
			if present {
				err := os.Setenv(tt.envKey, existingVal)
				if err != nil {
					t.Errorf("env error resetting %v", err)
				}
			} else {
				err := os.Unsetenv(tt.envKey)
				if err != nil {
					t.Errorf("env error unsetting %v", err)
				}
			}

		})
	}
}
