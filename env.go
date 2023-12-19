package env

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
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
		lookup = joinLookupFuncs(lookupFuncs...)
	}

	return EnvSet{
		flags:  make(map[string]flag),
		lookup: lookup,
	}
}

func (env *EnvSet) SetLookupFunc(fns ...LookupFunc) {
	env.lookup = joinLookupFuncs(fns...)
}

func (env EnvSet) Parse() error {
	errs := make([]error, 0, len(Environment.flags))
	for name, flag := range env.flags {
		envvar, ok := env.lookup(name)
		if !ok && flag.opts.required {
			errs = append(errs, fmt.Errorf("%q is required but not found", name))
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

// CommandLineArgs returns a lookup function that will search the provided args for flags.
// Since we often want our EnvironmentVariable name declarations to be reusable for command line args
// the lookup is case-insensitive and all underscores are changes to dashes.
// For example, a variable mapped to DATABASE_URL can be found using the --database-url flag when working with CommandLineArgs.
func CommandLineArgs(args ...string) LookupFunc {
	if len(args) == 0 {
		args = os.Args[1:]
	}

	var (
		m    = map[string][]string{}
		flag = ""
	)

	for _, arg := range args {
		switch {
		case strings.HasPrefix(arg, "-"):
			if flag != "" && len(m[flag]) == 0 {
				m[flag] = []string{"true"}
			}
			flag = strings.ToLower(strings.TrimLeft(arg, "-"))
			if key, value, ok := strings.Cut(flag, "="); ok {
				m[key] = append(m[key], value)
				flag = ""
			}
		case flag == "":
			// skip positional args
		default:
			m[flag] = append(m[flag], arg)
			flag = ""
		}
	}

	return func(name string) (string, bool) {
		name = strings.ReplaceAll(strings.ToLower(name), "_", "-")
		value, ok := m[name]
		return strings.Join(value, ","), ok
	}
}

type FSLookupOpts struct {
	Base string
}

func FileSystem(opts FSLookupOpts) LookupFunc {
	if opts.Base == "" {
		opts.Base = "."
	}
	return func(path string) (string, bool) {
		if !filepath.IsAbs(path) {
			path = filepath.Join(opts.Base, path)
		}

		data, err := os.ReadFile(path)
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				panic(err)
			}
			return "", false
		}

		return string(data), true
	}
}

func joinLookupFuncs(fns ...LookupFunc) LookupFunc {
	return func(key string) (value string, ok bool) {
		for _, fn := range fns {
			value, ok = fn(key)
			if ok {
				return
			}
		}
		return
	}
}
