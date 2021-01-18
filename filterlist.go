package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
)

func main() {
	input := "wordlist.txt"
	output := "filtered-%d.txt"

	allowed := "ACDEGOPRSUV"
	notAllowed := ""
	for _, c := range "ABCDEFGHIJKLMNOPQRSTUVWXYZ" {
		if strings.Contains(allowed, string(c)) {
			continue
		}
		notAllowed = notAllowed + string(c)
	}

	data, err := ioutil.ReadFile(input)
	if err != nil {
		log.Fatalf("can't read input: %v", err)
	}

	out := make(map[int][]string)
	dist := [9]int{}
	for _, l := range strings.Split(string(data), "\n") {
		if len(l) > 8 {
			continue
		}
		good := true
		for _, c := range notAllowed {
			if strings.Contains(l, string(c)) {
				good = false
				break
			}
		}

		palindrom := true
		for i := range l {
			if l[i] != l[len(l)-i-1] {
				palindrom = false
				break
			}
		}

		if good && !palindrom {
			dist[len(l)]++
			if out[len(l)] == nil {
				out[len(l)] = []string{}
			}
			out[len(l)] = append(out[len(l)], l)
		}
	}

	for i := range dist {
		if dist[i] == 0 {
			continue
		}
		log.Printf("%d: %d", i, dist[i])
		err = ioutil.WriteFile(fmt.Sprintf(output, i), []byte(strings.Join(out[i], "\n")), 0644)
		if err != nil {
			log.Fatalf("Can't write: %v", err)
		}
	}
}
