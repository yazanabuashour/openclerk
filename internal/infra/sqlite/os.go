package sqlite

import "os"

var (
	osMkdirAll   = os.MkdirAll
	osRemove     = os.Remove
	osReadFile   = os.ReadFile
	osStat       = os.Stat
	osWriteBytes = func(name string, data []byte) error {
		return os.WriteFile(name, data, 0o644)
	}
	osWriteFile = func(name string, data string) error {
		return os.WriteFile(name, []byte(data), 0o644)
	}
)
