package cfg

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type LookupFunc func(string) (string, bool)

type field struct {
	value value
	opts  options
}

type Parser struct {
	fields map[string]field
	lookup LookupFunc
}

func MakeParser(funcs ...LookupFunc) Parser {
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

	return Parser{
		fields: make(map[string]field),
		lookup: lookup,
	}
}

func (parser Parser) Parse() error {
	errs := make([]error, 0, len(parser.fields))
	for name, field := range parser.fields {
		envvar, ok := parser.lookup(name)
		if !ok && field.opts.required {
			errs = append(errs, fmt.Errorf("%q is required but not found", name))
			continue
		}
		if !ok {
			field.value.Set(field.opts.fallback)
			continue
		}

		if err := field.value.Parse(envvar); err != nil {
			errs = append(errs, fmt.Errorf("failed to parse %s: %v", name, err))
			continue
		}
	}
	return errors.Join(errs...)
}

// MustParse is like Parse but panics if an error occurs
func (parser Parser) MustParse() {
	if err := parser.Parse(); err != nil {
		panic(err)
	}
}

func Var[T any](parser Parser, p *T, name string, opts ...Options[T]) {
	parser.fields[name] = field{
		value: genericValue[T]{p},
		opts:  multiOpts[T](opts).toOptions(),
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

type FileSystemOptions struct {
	Base string
}

func FileSystem(opts FileSystemOptions) LookupFunc {
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
