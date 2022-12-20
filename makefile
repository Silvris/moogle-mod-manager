.DEFAULT_GOAL := build
BINARY_NAME=moogle-mod-manager
windows:
	go-winres make
	go build -ldflags="-s -H=windowsgui"  -o moogle-mod-manager.exe
	upx -9 -k moogle-mod-manager.exe
	rm moogle-mod-manager.ex~
	mv moogle-mod-manager.exe ./bin/moogle-mod-manager.exe
	#7z a -tzip moogle-mod-manager.zip  moogle-mod-manager.exe

wsl:
	wsl echo Linux; GOOS=linux GOARCH=amd64 go build -o ${BINARY_NAME}-linux.x86_64 main.go ;./upx -9 -k ${BINARY_NAME}-linux.x86_64; rm ${BINARY_NAME}-linux.x86_6~

#7z a -tzip moogle-mod-manager.zip  moogle-mod-manager.exe

linux:
	CGO_ENABLED=1 \
	GOOS=linux \
	GOARCH=amd64 \
	go build -o ${BINARY_NAME}-linux.x86_64 main.go
	./upx -9 -k ${BINARY_NAME}-linux.x86_64
	-rm ${BINARY_NAME}-linux.x86_6~

build:
	make wsl
#7z a -tzip moogle-mod-manager.zip  moogle-mod-manager.exe
