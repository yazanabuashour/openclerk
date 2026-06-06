package sqlite

import "os"

var (
	osMkdirAll   = os.MkdirAll
	osRename     = os.Rename
	osRemove     = os.Remove
	osReadFile   = os.ReadFile
	osStat       = os.Stat
	osWriteBytes = func(name string, data []byte) error {
		return os.WriteFile(name, data, 0o600)
	}
	osWriteFile = func(name string, data string) error {
		return os.WriteFile(name, []byte(data), 0o600)
	}
)
