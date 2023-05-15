package env_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/davidmdm/env"
	"github.com/stretchr/testify/require"
)

type Custom struct {
	Value any
}

func (c *Custom) UnmarshalText(data []byte) error {
	return json.Unmarshal(data, &c.Value)
}

func TestVar(t *testing.T) {
	var cfg struct {
		int         int
		uint        uint
		float64     float64
		boolean     bool
		duration    time.Duration
		string      string
		stringslice []string
		custom      Custom
	}

	environment := env.MakeEnvSet(func() env.LookupFunc {
		e := map[string]string{
			"i":      "-1",
			"ui":     "1",
			"f":      "3.14",
			"b":      "T",
			"d":      "5m",
			"s":      "hello",
			"ss":     "hello,world",
			"custom": `[1,2,3]`,
		}
		return func(s string) (string, bool) {
			value, ok := e[s]
			return value, ok
		}
	}())

	env.FlagVar(environment, &cfg.int, "i")
	env.FlagVar(environment, &cfg.uint, "ui")
	env.FlagVar(environment, &cfg.float64, "f")
	env.FlagVar(environment, &cfg.boolean, "b")
	env.FlagVar(environment, &cfg.string, "s")
	env.FlagVar(environment, &cfg.duration, "d")
	env.FlagVar(environment, &cfg.stringslice, "ss")
	env.FlagVar(environment, &cfg.custom, "custom")

	require.NoError(t, environment.Parse())

	require.Equal(t, -1, cfg.int)
	require.Equal(t, uint(1), cfg.uint)
	require.Equal(t, 3.14, cfg.float64)
	require.Equal(t, true, cfg.boolean)
	require.Equal(t, "hello", cfg.string)
	require.Equal(t, []string{"hello", "world"}, cfg.stringslice)
	require.Equal(t, 5*time.Minute, cfg.duration)
	require.EqualValues(t, []any{1.0, 2.0, 3.0}, cfg.custom.Value)
}
