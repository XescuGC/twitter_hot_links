build:
	go build -o hot-links -v *.go

run:
	go run *.go

clean:
	rm hot-links
