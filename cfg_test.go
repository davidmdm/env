package cfg_test

import (
	"encoding"
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/davidmdm/cfg"
	"github.com/stretchr/testify/require"
)

type Custom struct {
	Value any
}

func (c *Custom) UnmarshalText(data []byte) error {
	return json.Unmarshal(data, &c.Value)
}

func TestVar(t *testing.T) {
	var config struct {
		int         int
		uint        uint
		float64     float64
		boolean     bool
		duration    time.Duration
		string      string
		stringslice []string
		custom      Custom
		mapint      map[string]int
	}

	environment := cfg.MakeParser(func() cfg.LookupFunc {
		e := map[string]string{
			"i":      "-1",
			"ui":     "1",
			"f":      "3.14",
			"b":      "T",
			"d":      "5m",
			"s":      "hello",
			"ss":     "hello,world",
			"custom": `[1,2,3]`,
			"mapint": `x=3,y=1`,
		}
		return func(s string) (string, bool) {
			value, ok := e[s]
			return value, ok
		}
	}())

	cfg.Var(environment, &config.int, "i")
	cfg.Var(environment, &config.uint, "ui")
	cfg.Var(environment, &config.float64, "f")
	cfg.Var(environment, &config.boolean, "b")
	cfg.Var(environment, &config.string, "s")
	cfg.Var(environment, &config.duration, "d")
	cfg.Var(environment, &config.stringslice, "ss")
	cfg.Var(environment, &config.custom, "custom")
	cfg.Var(environment, &config.mapint, "mapint")

	require.NoError(t, environment.Parse())

	require.Equal(t, -1, config.int)
	require.Equal(t, uint(1), config.uint)
	require.Equal(t, 3.14, config.float64)
	require.Equal(t, true, config.boolean)
	require.Equal(t, "hello", config.string)
	require.Equal(t, []string{"hello", "world"}, config.stringslice)
	require.Equal(t, 5*time.Minute, config.duration)
	require.EqualValues(t, []any{1.0, 2.0, 3.0}, config.custom.Value)
	require.EqualValues(t, map[string]int{"x": 3, "y": 1}, config.mapint)
}

func TestMapParsingErrors(t *testing.T) {
	e1 := cfg.MakeParser(func(name string) (string, bool) {
		return "3=4", true
	})

	var boolint map[bool]int
	cfg.Var(e1, &boolint, "BOOLINT")

	var intbool map[int]bool
	cfg.Var(e1, &intbool, "INTBOOL")

	errText := e1.Parse().Error()
	require.Contains(
		t,
		errText,
		`failed to parse BOOLINT: failed to parse key: 3: strconv.ParseBool: parsing "3": invalid syntax`,
	)

	require.Contains(
		t,
		errText,
		`failed to parse INTBOOL: failed to parse value at key: 3: strconv.ParseBool: parsing "4": invalid syntax`,
	)
}

type CapText string

var _ encoding.TextUnmarshaler = new(CapText)

func (text *CapText) UnmarshalText(data []byte) error {
	*text = CapText(strings.ToUpper(string(data)))
	return nil
}

func TestTextUnmarshaler(t *testing.T) {
	var text CapText

	environment := cfg.MakeParser(func(s string) (string, bool) { return "value", true })
	cfg.Var(environment, &text, "VAR")

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

	environment := cfg.MakeParser(func(s string) (string, bool) { return "aGVsbG8gd29ybGQK", true })
	cfg.Var(environment, &text, "VAR")

	require.NoError(t, environment.Parse())
	require.Equal(t, "hello world\n", string(text))
}

func TestCommandLineArgsSource(t *testing.T) {
	environment := cfg.MakeParser(
		cfg.CommandLineArgs("--database-url", "db", "-force", "--filter=*.sql", "-count", "42", "-input", "a", "--input", "b"),
	)

	var (
		databaseURL string
		force       bool
		filter      string
		count       int
		other       string
		input       []string
	)

	cfg.Var(environment, &databaseURL, "DATABASE_URL")
	cfg.Var(environment, &force, "FORCE")
	cfg.Var(environment, &filter, "FILTER")
	cfg.Var(environment, &count, "COUNT")
	cfg.Var(environment, &other, "OTHER")
	cfg.Var(environment, &input, "INPUT")

	require.NoError(t, environment.Parse())

	require.True(t, force)
	require.Equal(t, "db", databaseURL)
	require.Equal(t, "*.sql", filter)
	require.Equal(t, 42, count)
	require.Equal(t, "", other)
	require.Equal(t, []string{"a", "b"}, input)
}

func TestMultipleLookups(t *testing.T) {
	environment := cfg.MakeParser(
		cfg.CommandLineArgs("--max=42"),
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

	cfg.Var(environment, &max, "MAX")
	cfg.Var(environment, &name, "NAME")

	require.NoError(t, environment.Parse())

	require.Equal(t, 42, max)
	require.Equal(t, "bob", name)
}
