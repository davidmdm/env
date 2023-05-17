# ENV

**Description**: A Go library for parsing environment variables and setting corresponding flags.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Example](#example)
- [API Reference](#api-reference)

## Installation

To use this library in your Go project, you need to have Go installed and set up on your machine. Once you have Go installed, you can install the library by running the following command:

```shell
go get github.com/davidmdm/env
```

## Usage

The library provides a convenient way to parse environment variables and set corresponding flags in your Go application.

To get started, import the library in your Go code:

```go
import "github.com/davidmdm/env"
```

### Creating an EnvSet

The `EnvSet` struct is used to define a set of environment variables and their corresponding flags. You can create an `EnvSet` by calling the `MakeEnvSet` function:

```go
envset := env.MakeEnvSet()
```

### Parsing Environment Variables

To parse the environment variables and set the corresponding flags, call the `Parse` method on an `EnvSet`:

```go
err := envset.Parse()
if err != nil {
    // Handle error
}
```

## Example

```go
import "github.com/davidmdm/env"

type Config struct {
    DatabaseURL string
    Timeout     time.Duration
}

func ParseConfig() (*Config, error) {
    var cfg Config

    env.Var(&cfg.DatabaseURL, "DATABASE_URL", env.Option[string]{Required: true})
    env.Var(&cfg.Timeout, "TIMEOUT", env.Option[time.Duration]{DefaultValue: 5 * time.Minute})

    return &cfg, env.Parse()
}
```

### Defining Flags with Options

You can define flags for environment variables using the `FlagVar` function, which supports options to specify additional behavior. The `Options` struct allows you to define options such as whether the environment variable is required and a default value.

Here's an example of defining a flag with options:

```go
var myFlag int
env.FlagVar(envset, &myFlag, "MY_ENV_VAR", env.Options{
    Required:     true,
    DefaultValue: 42,
})
```

### Convenience Functions

The library also provides convenience functions to work with a default `EnvSet` named `Environment`. These functions allow you to define flags and parse the environment variables without explicitly passing an `EnvSet` object.

#### Defining Flags with Options using `Var`

To define a flag using the `Var` function with options, you can directly call it with the variable, environment variable name, and options:

```go
var myFlag int
env.Var(&myFlag, "MY_ENV_VAR", env.Options{
    Required:     true,
    DefaultValue: 42,
})
```

#### Parsing Environment Variables with `Parse`

To parse the environment variables using the default `EnvSet`, call the `Parse` function:

```go
err := env.Parse()
if err != nil {
    // Handle error
}
```

#### Parsing Environment Variables with `MustParse`

The `MustParse` function is similar to `Parse` but panics if an error occurs during parsing:

```go
env.MustParse()
```

## API Reference

### Types

- `LookupFunc`: A function type that takes a string as input and returns a string and a boolean value indicating whether the lookup was successful or not.
- `EnvSet`: A struct representing a set of environment variables and their corresponding flags.
- `Options[T]`: A struct representing options for defining flags.

### Functions

- `MakeEnvSet(funcs ...LookupFunc) EnvSet`: Creates a new `EnvSet` with the specified lookup functions.
- `FlagVar[T any](envset EnvSet, p *T, name string, opts ...Options[T])`: Defines a flag for an environment variable with options.
- `Var[T any](p *T, name string, opts ...Options[T])`: Defines a flag using the default `EnvSet` with options
