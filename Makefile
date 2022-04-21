default: 
	go build cmd/client/main.go
	go build cmd/worker/main.go
	go build cmd/supervisor/main.go
 
clean:
	rm -rf cmd/client/client
	rm -rf cmd/worker/worker
	rm -rf cmd/supervisor/supervisor

.PHONY: default clean

