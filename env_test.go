package env_test

import (
	"encoding"
	"encoding/base64"
	"encoding/json"
	"strings"
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

type CapText string

var _ encoding.TextUnmarshaler = new(CapText)

func (text *CapText) UnmarshalText(data []byte) error {
	*text = CapText(strings.ToUpper(string(data)))
	return nil
}

func TestTextUnmarshaler(t *testing.T) {
	var text CapText

	environment := env.MakeEnvSet(func(s string) (string, bool) { return "value", true })
	env.FlagVar(environment, &text, "VAR")

	require.NoError(t, environment.Parse())
	require.Equal(t, "VALUE", string(text))
}

type Base64Text string

var _ encoding.BinaryUnmarshaler = new(Base64Text)

func (text *Base64Text) UnmarshalBinary(data []byte) error {
	result, err := base64.RawStdEncoding.DecodeString(string(data))
	if err != nil {
		return err
	}
	*text = Base64Text(result)
	return nil
}

func TestBinaryUnmarshaler(t *testing.T) {
	var text Base64Text

	environment := env.MakeEnvSet(func(s string) (string, bool) { return "aGVsbG8gd29ybGQK", true })
	env.FlagVar(environment, &text, "VAR")

	require.NoError(t, environment.Parse())
	require.Equal(t, "hello world\n", string(text))
}

func TestCommandLineArgsSource(t *testing.T) {
	environment := env.MakeEnvSet(
		env.CommandLineArgs("--database-url", "db", "-force", "--filter=*.sql", "-count", "42", "-input", "a", "--input", "b"),
	)

	var (
		databaseURL string
		force       bool
		filter      string
		count       int
		other       string
		input       []string
	)

	env.FlagVar(environment, &databaseURL, "DATABASE_URL")
	env.FlagVar(environment, &force, "FORCE")
	env.FlagVar(environment, &filter, "FILTER")
	env.FlagVar(environment, &count, "COUNT")
	env.FlagVar(environment, &other, "OTHER")
	env.FlagVar(environment, &input, "INPUT")

	require.NoError(t, environment.Parse())

	require.True(t, force)
	require.Equal(t, "db", databaseURL)
	require.Equal(t, "*.sql", filter)
	require.Equal(t, 42, count)
	require.Equal(t, "", other)
	require.Equal(t, []string{"a", "b"}, input)
}

func TestMultipleLookups(t *testing.T) {
	environment := env.MakeEnvSet(
		env.CommandLineArgs("--max=42"),
		func(name string) (string, bool) {
			value, ok := map[string]string{
				"MAX":  "10",
				"NAME": "bob",
			}[name]
			return value, ok
		},
	)

	var (
		max  int
		name string
	)

	env.FlagVar(environment, &max, "MAX")
	env.FlagVar(environment, &name, "NAME")

	require.NoError(t, environment.Parse())

	require.Equal(t, 42, max)
	require.Equal(t, "bob", name)
}
