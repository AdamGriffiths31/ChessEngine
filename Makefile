test:
	go test ./... -v
bench:
	go test ./... -bench=
run:
	go run main.go
testperft:
	 go test ./... -v -run TestMoveGenPerft