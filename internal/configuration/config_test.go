package configuration

import (
	"github.com/google/go-cmp/cmp"
	"github.com/spf13/viper"
	"os"
	"reflect"
	"strconv"
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
				v.Set("Pubsub.ProjectID", "test-project")
				v.Set("Pubsub.SubscriptionID", "test-subscription")
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
					Components []Component      `validate:"unique=Name,dive"`
					Groups     []ComponentGroup `validate:"unique=Name,dive"`
				}{
					ApiKey:  "foo",
					PageID:  "bar",
					ApiRoot: "https://api.statuspage.io/v1",
				},
				Pubsub: struct {
					ProjectID      string `validate:"required"`
					SubscriptionID string `validate:"required"`
				}{ProjectID: "test-project", SubscriptionID: "test-subscription"},
				Api: struct {
					Port   int
					Debug  bool
					Silent bool
				}{Port: 8080, Debug: false, Silent: false},
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
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("AssembleConfig() mismatch (-want +got):\n%s", diff)
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
					Components []Component      `validate:"unique=Name,dive"`
					Groups     []ComponentGroup `validate:"unique=Name,dive"`
				}{
					ApiRoot: "https://api.statuspage.io/v1",
				},
				Api: struct {
					Port   int
					Debug  bool
					Silent bool
				}{Port: 8080},
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
	type args struct {
		config *Config
	}
	tests := []struct {
		name         string
		args         args
		envVal       string
		envKey       string
		configAccess func(config *Config) string
		wantErr      bool
	}{
		{
			name:   "Reads Statuspage API key",
			args:   args{config: &Config{}},
			envVal: "foobar",
			envKey: "REVERE_STATUSPAGE_APIKEY",
			configAccess: func(config *Config) string {
				return config.Statuspage.ApiKey
			},
		},
		{
			name:   "Reads API port",
			args:   args{config: &Config{}},
			envVal: "123",
			envKey: "REVERE_API_PORT",
			configAccess: func(config *Config) string {
				return strconv.Itoa(config.Api.Port)
			},
		},
		{
			name:   "Errors on bad port",
			args:   args{config: &Config{}},
			envVal: "foobar",
			envKey: "REVERE_API_PORT",
			configAccess: func(config *Config) string {
				return strconv.Itoa(config.Api.Port)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			existingVal, present := os.LookupEnv(tt.envKey)
			if err := os.Setenv(tt.envKey, tt.envVal); err != nil {
				t.Errorf("env error setting %v", err)
			}
			if err := readEnvironmentVariables(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("readEnvironmentVariables had err %s, wantErr %t", err, tt.wantErr)
			}
			if got := tt.configAccess(tt.args.config); !tt.wantErr && (got != tt.envVal) {
				t.Errorf("readEnvironmentVariables() got %s for %s, want %s", got, tt.envKey, tt.envVal)
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

func Test_secondaryConfigValidation(t *testing.T) {
	type args struct {
		config *Config
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "allows correct mappings",
			args: args{config: &Config{
				Statuspage: struct {
					ApiKey     string `validate:"required"`
					PageID     string `validate:"required"`
					ApiRoot    string
					Components []Component      `validate:"unique=Name,dive"`
					Groups     []ComponentGroup `validate:"unique=Name,dive"`
				}{
					Components: []Component{
						{Name: "notebooks"},
						{Name: "ui"},
					},
				},
				ServiceToComponentMapping: []ServiceToComponentMapping{
					{ServiceName: "leonardo", AffectsComponentsNamed: []string{"notebooks"}},
					{ServiceName: "sam", AffectsComponentsNamed: []string{"notebooks", "ui"}},
					{ServiceName: "sherlock"},
				},
			}},
		},
		{
			name: "rejects bad mappings",
			args: args{config: &Config{
				Statuspage: struct {
					ApiKey     string `validate:"required"`
					PageID     string `validate:"required"`
					ApiRoot    string
					Components []Component      `validate:"unique=Name,dive"`
					Groups     []ComponentGroup `validate:"unique=Name,dive"`
				}{
					Components: []Component{
						{Name: "notebooks"},
						{Name: "ui"},
					},
				},
				ServiceToComponentMapping: []ServiceToComponentMapping{
					{ServiceName: "leonardo", AffectsComponentsNamed: []string{"notebooks"}},
					{ServiceName: "sam", AffectsComponentsNamed: []string{"notebooks", "ui"}},
					{ServiceName: "sherlock", AffectsComponentsNamed: []string{"preview-environments"}},
				},
			}},
			wantErr: true,
		},
		{
			name: "rejects bad mappings where there's no components",
			args: args{config: &Config{
				ServiceToComponentMapping: []ServiceToComponentMapping{
					{ServiceName: "leonardo", AffectsComponentsNamed: []string{"notebooks"}},
				},
			}},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := secondaryConfigValidation(tt.args.config); (err != nil) != tt.wantErr {
				t.Errorf("secondaryConfigValidation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
