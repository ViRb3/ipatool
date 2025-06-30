package keychain

//go:generate go run go.uber.org/mock/mockgen -source=keychain.go -destination=keychain_mock.go -package keychain
type Keychain interface {
	Get(key string) ([]byte, error)
	Set(key string, data []byte) error
	Remove(key string) error
}

type keychain struct {
	directory    string
	passwordFunc func(string) (string, error)
	password     string
}

type Args struct {
	Directory    string
	PasswordFunc func(string) (string, error)
}

func New(args Args) Keychain {
	return &keychain{
		directory:    args.Directory,
		passwordFunc: args.PasswordFunc,
	}
}
