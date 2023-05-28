# ENV

**Description**: A Go library for parsing environment variables and setting corresponding flags.

## Table of Contents

- [Why](#why)
- [Usage](#usage)
- [Example](#example)

## Why

Why another package for parsing the environment? Currently, most popular environment parsing libraries depend on struct tags to map the environment to a structure and provide options like flag requirement or default values when absent.

With `env` and the use of Go Generics, we can now have a type-safe API that doesn't depend on struct tags and can take advantage of strong typing.

Let's contrast `davidmdm/env` with `github.com/kelseyhightower/envconfig`:

The envconfig approach is convenient but very sensitive to typos, and the defaults need to be encoded in their string format, which can be error-prone.

```go
package main

import (
    "time"
    "github.com/kelseyhightower/envconfig"
)

type Config struct {
    DatabaseURL string        `envconfig:"DATABASE_URL" required:"true"`
    Timeout     time.Duration `envconfig:"TIMEMOUT" default:"5m"`
}

func main() {
    var cfg Config
    envconfig.Process("", &cfg)
}
```

On the other hand, `davidmdm/env` does not suffer from these problems. It also has the added benefit of being programmatic instead of static. If we need, environment variable names and options could be determined at runtime instead of statically typed into a struct definition.

```go
package main

import (
    "time"
    "github.com/davidmdm/env"
)

type Config struct {
    DatabaseURL string
    Timeout     time.Duration
}

func main() {
    var cfg Config
    env.Var(&cfg.DatabaseURL, "DATABASE_URL", env.Options[string]{Required: true})
    env.Var(&cfg.Timeout, "TIMEOUT", env.Options[time.Duration]{Default: 5 * time.Minute})
    env.Parse()
}
```

## Usage

The library provides a convenient way to parse environment variables and set corresponding flags in your Go application.

### Creating an EnvSet

The `EnvSet` struct is used to define a set of environment variables and their corresponding flags. You can create an `EnvSet` by calling the `MakeEnvSet` function:

```go
envset := env.MakeEnvSet()
```

By default, an envset will use `os.Lookup` to find environment variables. However, when instantiating, you can pass a variadic number of lookup funcs that will be used one after the other while searching for the flag variable. This can be useful for testing or when you want your data to come from other sources like `hashicorp/vault`, `AWS Secret Manager`, or anything you desire.

```go
envset := env.MakeEnvSet(func(envvar string) (string, bool) {
  // lookup from some source.
})
```

### Setting Flags on an EnvSet

```go
envset := env.MakeEnvSet()

var max int
env.FlagVar(envset, &max, "MAX")
```

### Custom Decoding

`env` will use reflection to figure out how to parse the environment string into the target variable, and this works well for common types. However, if a type implements `encoding.TextUnmarshaler` or `encoding.BinaryUnmarshaler`, then those methods will be used.

Example taken from the tests:

```go
type Base64Text string

var _ encoding.BinaryUnmarshaler = new(Base64Text)

func (text *Base64Text) UnmarshalBinary(data []byte) error {
	result, err := base64.Raw

StdEncoding.DecodeString(string(data))
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
```

### Parsing Environment Variables

To parse the environment variables and set the corresponding flags, call the `Parse` method on an `EnvSet`:

```go
err := envset.Parse()
if err != nil {
    // Handle error
}
```

### Convenience Variables and Functions

- `env.Environment`: The default envset exposed by the package, which looks up envvars using `os.Lookup`
- `env.Var()`: Convenience function to set flags on the default environment, equivalent to: `env.FlagVar(env.Environment, ...)`
- `env.Parse()`: Convenience function to parse the default Environment envset.
- `env.MustParse()`: Same as `env.Parse`, but panics if an error is encountered

## Example

```go
import "github.com/davidmdm/env"

type Config struct {
    DatabaseURL string
    Timeout     time.Duration
}

func ParseConfig() (*Config, error) {
    var cfg Config

    env.Var(&cfg.DatabaseURL, "DATABASE_URL", env.Options[string]{Required: true})
    env.Var(&cfg.Timeout, "TIMEOUT", env.Options[time.Duration]{DefaultValue: 5 * time.Minute})

    return &cfg, env.Parse()
}
```
