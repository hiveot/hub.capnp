# Build for mosquitto authentication & authorization plugin for WoST Hub
CFLAGS := -I/usr/local/include -fPIC
LDFLAGS :=  -shared 
DIST := ../../dist

all: mosqauth

#test:
#env CGO_LDFLAGS="$(LDFLAGS)" go test 


mosqauth:
	env CGO_CFLAGS="$(CFLAGS)" go build -buildmode=c-archive main.go
	env CGO_LDFLAGS="$(LDFLAGS)" go build -buildmode=c-shared -o mosqauth.so
	rm *.a
	mv mosqauth.so $(DIST)/bin

clean:
	rm *.a *.so
	rm mosqauth.h
