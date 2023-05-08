package env

import (
	"errors"
	"fmt"
	"os"
)

type EnvSet struct {
	flags map[string]flag
}

var Environment = EnvSet{flags: make(map[string]flag)}

type flag struct {
	value value
	opts  flagOptions
}

func Var[T any](p *T, name string, opts ...Options[T]) {
	Environment.flags[name] = flag{
		value: genericValue[T]{p},
		opts:  multiOpts[T](opts).toFlagOptions(),
	}
}

func Parse() error {
	errs := make([]error, 0, len(Environment.flags))
	for name, flag := range Environment.flags {
		envvar, ok := os.LookupEnv(name)
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

func MustParse() {
	if err := Parse(); err != nil {
		panic(err)
	}
}
