package env

import "os"

// GetEnvString checks is environment variable set and if yes
// returns its value. If not, it returns provided default value.
func GetEnvString(key, defval string) string {
	v := os.Getenv(key)
	if v == "" {
		return defval
	}
	return v
}