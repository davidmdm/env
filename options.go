package env

type flagOptions struct {
	required bool
	fallback any
}

type Options[T any] struct {
	Required     bool
	DefaultValue T
}

func (opts Options[T]) toFlagOptions() flagOptions {
	return flagOptions{
		required: opts.Required,
		fallback: opts.DefaultValue,
	}
}

type multiOpts[T any] []Options[T]

func (opts multiOpts[T]) toFlagOptions() flagOptions {
	if len(opts) == 0 {
		var zero T
		return flagOptions{
			required: false,
			fallback: zero,
		}
	}
	return opts[0].toFlagOptions()
}
