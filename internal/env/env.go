package env

import "os"

type EnvKey string

func (key EnvKey) GetValue() string {
	return os.Getenv(string(key))
}

const (
	CiToken       EnvKey = "CI_TOKEN"
	ReadOnlyToken EnvKey = "READ_ONLY_TOKEN"
)
