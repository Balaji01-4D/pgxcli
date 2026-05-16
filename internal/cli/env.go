package cli

import (
	"os"
	"strconv"
)

func firstEnv(keys ...string) string {
	for _, key := range keys {
		if val := os.Getenv(key); val != "" {
			return val
		}
	}

	return ""
}

func getUserFromEnv() string {
	return firstEnv("PGXUSER", "PGUSER")
}

func getPasswordFromEnv() string {
	return firstEnv("PGXPASSWORD", "PGPASSWORD")
}

func getHostFromEnv() string {
	return firstEnv("PGXHOST", "PGHOST")
}

func getDatabaseFromEnv() string {
	return firstEnv("PGXDATABASE", "PGDATABASE")
}

func getPortFromEnv() uint16 {
	portStr := firstEnv("PGXPORT", "PGPORT")
	if portStr == "" {
		return 0
	}

	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0
	}

	return uint16(port)
}