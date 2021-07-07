CFLAGS := -I/usr/local/include -fPIC
LDFLAGS :=  -shared 
DIST := ../../../../dist

all: mosqauth

mosqauth:
	env CGO_CFLAGS="$(CFLAGS)" go build -buildmode=c-archive main.go
	env CGO_LDFLAGS="$(LDFLAGS)" go build -buildmode=c-shared -o mosqauth.so
	rm *.a
	mv mosqauth.so $(DIST)/bin
	#rm mosqauth.h

clean:
	rm *.a *.so
	rm mosqauth.h
	