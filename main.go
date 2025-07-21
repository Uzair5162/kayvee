package main

import (
	"bufio"
	"fmt"
	"kayvee/persistence"
	"kayvee/store"
	"os"
	"strconv"
	"strings"
)

func main() {
	fp := persistence.NewFilePersister("data.json")
	s, err := store.New(store.Config{
		Persister: fp,
	})
	if err != nil {
		fmt.Println(err)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)

	for {
		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				fmt.Println(err)
			}
			s.Shutdown()
			return
		}

		words := strings.Fields(scanner.Text())
		if len(words) == 0 {
			fmt.Println("invalid command")
			continue
		}

		cmd := strings.ToUpper(words[0])
		switch cmd {
		case "SET":
			if len(words) < 3 || len(words) > 4 {
				fmt.Println("usage: SET key value [ttlSec]")
				continue
			}

			ttl := 0
			if len(words) == 4 {
				v, err := strconv.Atoi(words[3])
				if err != nil {
					fmt.Println("invalid ttl")
					continue
				}
				ttl = v
			}
			if err := s.Set(words[1], words[2], ttl); err != nil {
				fmt.Println(err)
			}
		case "GET":
			if len(words) != 2 {
				fmt.Println("usage: GET key")
				continue
			}

			if v, ok := s.Get(words[1]); ok {
				fmt.Println(v)
			} else {
				fmt.Println("no such key")
			}
		case "DEL":
			if len(words) != 2 {
				fmt.Println("usage: DEL key")
			}

			if err := s.Del(words[1]); err != nil {
				fmt.Println(err)
			}
		case "OUT":
			for _, line := range s.Snapshot() {
				fmt.Println(line)
			}
		case "STOP":
			s.Shutdown()
			return
		default:
			fmt.Println("invalid command")
		}
	}
}
