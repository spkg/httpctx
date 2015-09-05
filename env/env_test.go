package env_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"sp.com.au/exp/env"
)

func TestEnv(t *testing.T) {
	testCases := []struct {
		Name          string
		IsValid       bool
		IsProduction  bool
		IsTest        bool
		IsDevelopment bool
	}{
		{"production", true, true, false, false},
		{"test", true, false, true, false},
		{"development", true, false, false, true},
		{"production1", true, true, false, false},
		{"test1", true, false, true, false},
		{"development1", true, false, false, true},
		{"DEVELOPMENT", false, false, false, false},
		{"something", false, false, false, false},
	}

	for _, tc := range testCases {
		err := env.SetName(tc.Name)
		if tc.IsValid {
			assert.NoError(t, err)
			assert.Equal(t, tc.IsProduction, env.IsProduction())
			assert.Equal(t, tc.IsTest, env.IsTest())
			assert.Equal(t, tc.IsDevelopment, env.IsDevelopment())
			assert.Equal(t, tc.Name, env.Name())
			if tc.IsTest {
				assert.NoError(t, env.CheckTest())
			} else {
				assert.Equal(t, env.ErrNotTest, env.CheckTest())
			}
		} else {
			assert.Error(t, err)
		}
	}
}

func TestNameFor(t *testing.T) {
	prevFormat := env.FormatName
	defer func() { env.FormatName = prevFormat }()
	env.FormatName = func(envname, separator, name string) string {
		return "oslo" + separator + envname + separator + name
	}

	testCases := []struct {
		EnvName     string
		TableName   string
		AlteredName string
	}{
		// production is special -- the name is not changed
		{"production", "queue-name", "oslo-production-queue-name"},
		{"production1", "queue-name", "oslo-production1-queue-name"},
		{"test", "active.users", "oslo.test.active.users"},
		{"development", "queue-name", "oslo-development-queue-name"},
		{"development", "tablename", "oslo_development_tablename"},
		{"development", "table_name", "oslo_development_table_name"},
		{"development", "table_name.1", "oslo_development_table_name.1"},
		{"development", "table-name.1", "oslo-development-table-name.1"},
	}

	for _, tc := range testCases {
		env.MustSetName(tc.EnvName)
		assert.Equal(t, tc.AlteredName, env.NameFor(tc.TableName))
	}
}
