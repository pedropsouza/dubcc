build: assembler gui debug

assembler: ./assembler/*.go
	go build -C ./assembler -v

vm: ./VirtualMachine/*.go
	go build -C ./VirtualMachine -v

debug: ./debug/*.go
	go build -C ./debug -v 

linker: ./assemler/linker/*.go
	go build -C ./linker -v 

