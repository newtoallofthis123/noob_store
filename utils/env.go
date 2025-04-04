package utils

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Env struct {
	ConnString string
	ListenAddr string
	BucketPath string
	CacheConn  string
}

// Reads the .env file and returns an Env struct.
func ReadEnv() Env {
	if os.Getenv("GIN_MODE") == "release" {
		godotenv.Load(".env.prod")
	} else {
		godotenv.Load(".env")
	}

	return Env{
		ConnString: constructDbString(),
		ListenAddr: getEnv("LISTEN_ADDR"),
		BucketPath: getEnv("BUCKET_PATH"),
		CacheConn:  getEnv("CACHE_CONN"),
	}
}

func constructDbString() string {
	return fmt.Sprintf("user=%s password=%s host=%s dbname=%s port=%s sslmode=disable", getEnv("DB_USER"), getEnv("DB_PASS"), getEnv("DB_HOST"), getEnv("DB_NAME"), getEnv("DB_PORT"))
}

// Returns the value of the given env var name.
func getEnv(name string) string {
	val, ok := os.LookupEnv(name)
	if !ok {
		panic(fmt.Sprintf("Env var %s not found", name))
	}
	return val
}
