package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"strings"

	"github.com/AdamGriffiths31/ChessEngine/search"
	"github.com/AdamGriffiths31/ChessEngine/uci"
)

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	reader := bufio.NewReader(os.Stdin)
	for {
		input, err := reader.ReadString('\n')
		if err != nil {
			panic(fmt.Errorf("main reader loop: %v", err))
		}
		input = strings.TrimSpace(input)

		if input == "uci" {
			uci := uci.NewUCI()
			uci.UCIMode()
			continue
		}

		if input == "b" {
			search.RunBenchmark()
		}

		if input == "quit" {
			break
		}
	}
}
