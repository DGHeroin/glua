all:
	go build -buildmode c-shared -o glua.so lua.go

clean:
	@-rm *.so