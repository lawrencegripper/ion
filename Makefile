dispatcher:
	make -f build/dispatcher/Makefile.Docker
	
sidecar:
	make -f build/sidecar/Makefile.Docker
	
all: dispatcher sidecar