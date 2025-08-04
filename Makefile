build: build_assembler build_gui

build_assembler: ./assembler/*.go
	go build -C ./assembler -v

build_gui: ./VirtualMachine/*.go
	go build -C ./VirtualMachine -v
