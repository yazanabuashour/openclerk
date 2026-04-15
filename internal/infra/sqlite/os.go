package sqlite

import "os"

var (
	osMkdirAll  = os.MkdirAll
	osReadFile  = os.ReadFile
	osStat      = os.Stat
	osWriteFile = func(name string, data string) error {
		return os.WriteFile(name, []byte(data), 0o644)
	}
)
