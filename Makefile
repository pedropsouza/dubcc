build: assembler linker vm debug

assembler: ./assembler/*.go
	go build -C ./assembler -v

vm: ./VirtualMachine/*.go
	go build -C ./VirtualMachine -v

debug: ./debug/*.go
	go build -C ./debug -v 

linker: ./linker/*.go
	go build -C ./linker -v 

