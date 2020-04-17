build4linux:
	CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-w" -o jenkins-pipeline-go main.go

build4windows:
	CGO_ENABLED=0 GOOS=windows go build -a -ldflags "-w" -o jenkins-pipeline-go.exe main.go
