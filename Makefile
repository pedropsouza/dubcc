build: build_assembler build_gui build_debug

build_assembler: ./assembler/*.go
	go build -C ./assembler -v

build_gui: ./VirtualMachine/*.go
	go build -C ./VirtualMachine -v

build_debug: ./debug/*.go
	go build -C ./debug -v 
