package main

import (
	"bufio"
	"fmt"
	"kayvee/store"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	s := store.New(time.Duration(1) * time.Second)

	scanner := bufio.NewScanner(os.Stdin)

	for {
		scanner.Scan()
		words := strings.Fields(scanner.Text())

		if len(words) == 0 {
			fmt.Println("invalid command")
			continue
		}

		if words[0] == "SET" && len(words) == 3 {
			s.Set(words[1], words[2], 0)
		} else if words[0] == "SET" && len(words) == 4 {
			if ttl, err := strconv.Atoi(words[3]); err == nil {
				s.Set(words[1], words[2], ttl)
			} else {
				fmt.Println("invalid ttl")
			}
		} else if words[0] == "GET" && len(words) == 2 {
			if v, ok := s.Get(words[1]); ok {
				fmt.Println(v)
			} else {
				fmt.Println("no such key")
			}
		} else if words[0] == "DEL" && len(words) == 2 {
			if ok := s.Del(words[1]); !ok {
				fmt.Println("no such key")
			}
		} else if words[0] == "OUT" {
			s.Display()
		} else {
			fmt.Println("invalid command")
		}
	}
}
