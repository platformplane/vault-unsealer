package main

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/vault/api"
)

func main() {
	pathToken := os.Getenv("UNSEALER_PATH_TOKEN")
	pathShards := os.Getenv("UNSEALER_PATH_SHARDS")

	if pathToken == "" {
		pathToken = "vault.token"
	}

	if pathShards == "" {
		pathShards = "vault.shards"
	}

	cfg := api.DefaultConfig()

	client, err := api.NewClient(cfg)

	if err != nil {
		log.Fatal(err)
	}

	for {
		health, err := client.Sys().Health()

		if err != nil {
			log.Println("vault not ready", err)

			time.Sleep(2 * time.Second)
			continue
		}

		if !health.Initialized {
			init, err := client.Sys().Init(&api.InitRequest{
				SecretShares:    5,
				SecretThreshold: 3,
			})

			if err != nil {
				log.Fatal("unable to init vault", err)
			}

			if err := writeShards(pathShards, init.Keys); err != nil {
				log.Fatal("unable to write vault secrets", err)
			}

			if err := writeToken(pathToken, init.RootToken); err != nil {
				log.Fatal("unable to write vault secrets", err)
			}
		}

		if health.Sealed {
			shards, err := readShards(pathShards)

			if err != nil {
				log.Fatal("unable to read vault secrets", err)
			}

			sealed := true

			for _, shard := range shards {
				seal, err := client.Sys().Unseal(shard)

				if err != nil {
					log.Fatal("unable to unseal with shard", err)
				}

				sealed = seal.Sealed

				if !sealed {
					break
				}
			}

			if sealed {
				log.Fatal("unable to unseal vault")
			}
		}

		time.Sleep(10 * time.Second)
	}
}

func writeShards(path string, lines []string) error {
	data := []byte(strings.Join(lines, "\n"))
	return os.WriteFile(path, data, 0600)
}

func readShards(path string) ([]string, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	if len(data) == 0 {
		return nil, errors.New("no shards found")
	}

	shards := strings.Split(string(data), "\n")
	return shards, nil
}

func writeToken(path string, token string) error {
	data := []byte(token)
	return os.WriteFile(path, data, 0600)
}

func readToken(path string) (string, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return "", err
	}

	if len(data) == 0 {
		return "", errors.New("no token found")
	}

	token := string(data)
	return token, nil
}
