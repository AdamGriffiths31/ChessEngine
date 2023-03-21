package perft2

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"testing"
)

func TestMoveGenPerft(t *testing.T) {
	// Open the file
	//startTime := time.Now()

	file, err := os.Open("perft.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	counter := 0
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ",")
		fen := parts[0]
		counter++
		for i := 1; i < len(parts) && i <= 3; i++ {
			ans := PerftTest(i, fen)

			expected, _ := strconv.Atoi(parts[i])
			if ans != int64(expected) {
				t.Logf("Read line: %s", line)
				t.Errorf("got %d, want %d", ans, expected)
			}
		}

		// if counter == 1000 {
		// 	duration := time.Since(startTime)
		// 	fmt.Println("\n\nTime elapsed:", duration)
		// 	break
		// }
	}

	if err := scanner.Err(); err != nil {
		t.Fatal(err)
	}
}
