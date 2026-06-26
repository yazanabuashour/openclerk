package sqlite

import "os"

var (
	osMkdirAll   = os.MkdirAll
	osLstat      = os.Lstat
	osRemove     = os.Remove
	osReadFile   = os.ReadFile
	osStat       = os.Stat
	osWriteBytes = func(name string, data []byte) error {
		return os.WriteFile(name, data, 0o600)
	}
)
