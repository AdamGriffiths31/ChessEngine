package fen

// func TestValidFen(t *testing.T) {
// 	var tests = []string{
// 		"rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
// 		"rnbqkbnr/pp1ppppp/8/2p5/4P3/8/PPPP1PPP/RNBQKBNR w KQkq c6 0 2",
// 		"rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2",
// 	}

// 	for _, tt := range tests {
// 		testname := fmt.Sprintf("%v", tt)
// 		fmt.Printf("%v", testname)
// 		t.Run(testname, func(t *testing.T) {
// 			ans := IsValid(tt)
// 			if ans != true {
// 				t.Errorf("got %v, want %v", ans, true)
// 			}
// 		})
// 	}
// }

// func TestInvalidFen(t *testing.T) {
// 	var tests = []string{
// 		"Invalid string",
// 		"",
// 		"test rnbqkbnr/pp1ppppp/8/2p5/4P3/5N2/PPPP1PPP/RNBQKB1R b KQkq - 1 2",
// 	}

// 	for _, tt := range tests {
// 		testname := fmt.Sprintf("%v", tt)
// 		fmt.Printf("%v", testname)
// 		t.Run(testname, func(t *testing.T) {
// 			ans := IsValid(tt)
// 			if ans != false {
// 				t.Errorf("got %v, want %v", ans, false)
// 			}
// 		})
// 	}
// }
