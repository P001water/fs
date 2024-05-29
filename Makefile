BUILD_ENV = CGO_ENABLED=0
OPTIONS = -trimpath -ldflags "-w -s"
NAME = fs

.PHONY: all linux windows macos mips arm clean

all:
	${BUILD_ENV} GOOS=linux GOARCH=386 go build ${OPTIONS} -o release/${NAME}_86 main.go
	${BUILD_ENV} GOOS=linux GOARCH=amd64 go build ${OPTIONS} -o release/${NAME}_64 main.go
	${BUILD_ENV} GOOS=windows GOARCH=amd64 go build ${OPTIONS} -o release/${NAME}_64.exe main.go
	${BUILD_ENV} GOOS=windows GOARCH=386 go build ${OPTIONS} -o release/${NAME}_86.exe main.go
	${BUILD_ENV} GOOS=darwin GOARCH=amd64 go build ${OPTIONS} -o release/${NAME}_darwin64 main.go
	${BUILD_ENV} GOOS=darwin GOARCH=arm64 go build ${OPTIONS} -o release/${NAME}_darwinarm64 main.go
	${BUILD_ENV} GOOS=linux GOARCH=mipsle go build ${OPTIONS} -o release/${NAME}_mipsel main.go
	${BUILD_ENV} GOOS=linux GOARCH=arm64 go build ${OPTIONS} -o release/${NAME}_arm64 main.go

linux:
	${BUILD_ENV} GOOS=linux GOARCH=386 go build ${OPTIONS} -o release/${NAME}_86 main.go
	${BUILD_ENV} GOOS=linux GOARCH=amd64 go build ${OPTIONS} -o release/${NAME}_64 main.go
	${BUILD_ENV} GOOS=linux GOARCH=arm64 go build ${OPTIONS} -o release/${NAME}_arm64 main.go

windows:
	${BUILD_ENV} GOOS=windows GOARCH=amd64 go build ${OPTIONS} -o release/${NAME}_64.exe main.go
	${BUILD_ENV} GOOS=windows GOARCH=386 go build ${OPTIONS} -o release/${NAME}_86.exe main.go

macos:
	${BUILD_ENV} GOOS=darwin GOARCH=amd64 go build ${OPTIONS} -o release/${NAME}_darwin64 main.go
	${BUILD_ENV} GOOS=darwin GOARCH=arm64 go build ${OPTIONS} -o release/${NAME}_darwinarm64 main.go

arm:
	${BUILD_ENV} GOOS=linux GOARCH=arm GOARM=5 go build ${OPTIONS} -o release/${NAME}_arm64 main.go

mips:
	${BUILD_ENV} GOOS=linux GOARCH=mipsle go build ${OPTIONS} -o release/${NAME}_mipsel main.go




# Here is a special situation
# You can see Stowaway get the params passed by the user through console by default
# But if you define the params in the program(instead of passing them by the console),you can just run Stowaway agent by double-click
# Sounds great? Right?
# But it is slightly weird on Windows since double-clicking Stowaway agent or entering "shell" command in Stowaway admin will spawn a cmd window
# That makes Stowaway pretty hard to hide itself
# To solve this,here is my solution
# First, check the detail in "agent/shell.go", follow my instruction and change some codes
# Then, run `make windows_nogui` and get your bonus!

clean:
	@rm release/*