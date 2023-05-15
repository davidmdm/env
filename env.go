package env

import (
	"errors"
	"fmt"
	"os"
)

type (
	LookupFunc func(string) (string, bool)
	EnvSet     struct {
		flags  map[string]flag
		lookup LookupFunc
	}
)

func MakeEnvSet(fn LookupFunc) EnvSet {
	if fn == nil {
		fn = os.LookupEnv
	}
	return EnvSet{
		flags:  make(map[string]flag),
		lookup: fn,
	}
}

func (env EnvSet) Parse() error {
	errs := make([]error, 0, len(Environment.flags))
	for name, flag := range env.flags {
		envvar, ok := env.lookup(name)
		if !ok && flag.opts.required {
			errs = append(errs, fmt.Errorf("environment variable is required: %s", name))
			continue
		}
		if !ok {
			flag.value.Set(flag.opts.fallback)
			continue
		}

		if err := flag.value.Parse(envvar); err != nil {
			errs = append(errs, fmt.Errorf("failed to parse %s: %v", name, err))
			continue
		}
	}
	return errors.Join(errs...)
}

func FlagVar[T any](envset EnvSet, p *T, name string, opts ...Options[T]) {
	envset.flags[name] = flag{
		value: genericValue[T]{p},
		opts:  multiOpts[T](opts).toFlagOptions(),
	}
}

type flag struct {
	value value
	opts  flagOptions
}

var Environment = EnvSet{
	flags:  make(map[string]flag),
	lookup: os.LookupEnv,
}

func Var[T any](p *T, name string, opts ...Options[T]) {
	FlagVar(Environment, p, name, opts...)
}

func Parse() error {
	return Environment.Parse()
}

func MustParse() {
	if err := Parse(); err != nil {
		panic(err)
	}
}
