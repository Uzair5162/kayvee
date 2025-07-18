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
		scanner.Scan()
		words := strings.Fields(scanner.Text())

		if len(words) == 0 {
			fmt.Println("invalid command")
			continue
		}

		if words[0] == "SET" && len(words) == 3 {
			if err := s.Set(words[1], words[2], 0); err != nil {
				fmt.Println(err)
			}
		} else if words[0] == "SET" && len(words) == 4 {
			ttl, err := strconv.Atoi(words[3])
			if err != nil {
				fmt.Println("invalid ttl")
			}

			if err := s.Set(words[1], words[2], ttl); err != nil {
				fmt.Println(err)
			}
		} else if words[0] == "GET" && len(words) == 2 {
			if v, ok := s.Get(words[1]); ok {
				fmt.Println(v)
			} else {
				fmt.Println("no such key")
			}
		} else if words[0] == "DEL" && len(words) == 2 {
			if err := s.Del(words[1]); err != nil {
				fmt.Println(err)
			}
		} else if words[0] == "OUT" {
			for _, line := range s.Snapshot() {
				fmt.Println(line)
			}
		} else if words[0] == "STOP" {
			s.StopEvictionLoop()
			return
		} else {
			fmt.Println("invalid command")
		}
	}
}
