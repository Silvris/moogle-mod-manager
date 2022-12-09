<<<<<<< Updated upstream
.DEFAULT_GOAL := build
build:
	go-winres make
	go build -ldflags="-s -H=windowsgui"  -o moogle-mod-manager.exe
	upx -9 -k moogle-mod-manager.exe
	rm moogle-mod-manager.ex~
	mv moogle-mod-manager.exe ./bin/moogle-mod-manager.exe
	#7z a -tzip moogle-mod-manager.zip  moogle-mod-manager.exe
=======
BINARY_NAME=moogle-mod-manager

linux:
	CGO_ENABLED=1 \
	GOOS=linux \
	GOARCH=amd64 \
	go build -o ${BINARY_NAME}-linux.x86_64 main.go
	./upx -9 -k ${BINARY_NAME}-linux.x86_64
	-rm ${BINARY_NAME}-linux.x86_6~

build:
	make linux
#7z a -tzip moogle-mod-manager.zip  moogle-mod-manager.exe
>>>>>>> Stashed changes
