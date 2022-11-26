package edison

type OptionType int

const (
	OptRestPort OptionType = iota
	OptGrcpPort
	OptGracefullShutdown
)

type Option interface {
	Type() OptionType

	Value() interface{}
}

type (
	optRestPort          string
	optGrcpPort          string
	optGracefullShutdown bool
)

type option struct {
	restPort          string
	grpcPort          string
	gracefullShutdown bool
}

func RestPort(port string) Option {
	return optRestPort(port)
}

func (v optRestPort) Type() OptionType   { return OptRestPort }
func (v optRestPort) Value() interface{} { return v }

func GrpcPort(port string) Option {
	return optGrcpPort(port)
}

func (v optGrcpPort) Type() OptionType   { return OptGrcpPort }
func (v optGrcpPort) Value() interface{} { return v }

func GracefullShutdown() Option {
	return optGracefullShutdown(true)
}

func (v optGracefullShutdown) Type() OptionType   { return OptGracefullShutdown }
func (v optGracefullShutdown) Value() interface{} { return v }

func composeOptions(opts ...Option) (option, error) {
	res := option{
		restPort: "8080",
		grpcPort: "9090",
	}

	for _, opt := range opts {
		switch opt := opt.(type) {
		case optRestPort:
			res.restPort = string(opt)
		case optGrcpPort:
			res.grpcPort = string(opt)
		case optGracefullShutdown:
			res.gracefullShutdown = bool(opt)
		default:
		}
	}

	return res, nil
}
