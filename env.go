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

func MakeEnvSet(funcs ...LookupFunc) EnvSet {
	lookupFuncs := make([]LookupFunc, 0, len(funcs))
	for _, fn := range funcs {
		if fn == nil {
			continue
		}
		lookupFuncs = append(lookupFuncs, fn)
	}

	lookup := os.LookupEnv
	if len(lookupFuncs) > 0 {
		lookup = func(key string) (value string, ok bool) {
			for _, fn := range lookupFuncs {
				value, ok = fn(key)
				if ok {
					return
				}
			}
			return
		}
	}

	return EnvSet{
		flags:  make(map[string]flag),
		lookup: lookup,
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

// MustParse is like Parse but panics if an error occurs
func (env EnvSet) MustParse() {
	if err := env.Parse(); err != nil {
		panic(err)
	}
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
