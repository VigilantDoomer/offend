PLATFORMS := linux/amd64/ windows/amd64/.exe /linux/386/32 /windows/386/32.exe

temp = $(subst /, ,$@)
os = $(word 1, $(temp))
arch = $(word 2, $(temp))
suffix = $(word 3, $(temp))

release: $(PLATFORMS)

clean:
	go clean
	rm -f offend
	rm -f offend.exe
	rm -f offend32
	rm -f offend32.exe

$(PLATFORMS):
	GOOS=$(os) GOARCH=$(arch) go build -ldflags="-s -w -buildid=" -trimpath -o 'offend$(suffix)'

