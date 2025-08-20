package assembler

import (
	"dubcc"
)

type DulfSection    uint32
type SymbolBinding  uint8
type SymbolType     uint8
type RelocationType uint32

const (
	SHT_PROGBITS  DulfSection = iota  // code
	SHT_SYMTAB                        // symbol table (not using yet)
	SHT_STRTAB                        // string table (not using yet)
	SHT_RELA                          // relocation entries (not using yet)
	SHT_NOBITS                       // uninitialized data (BSS) (not using yet)
)

const (
	STB_LOCAL  	SymbolBinding = iota // local symbols
	STB_GLOBAL                       // global symbols
)

const (
	STT_NOTYPE  SymbolType = iota // no type specified
	STT_DATA	                    // data
	STT_FUNC                      // function
	STT_SECTION                   // section
)

const (
	R_ABSOLUTE RelocationType = 1 // direct reference
	R_RELATIVE RelocationType = 2 // PC relative reference
)

type DulfHeader struct {
	Magic         [4]byte              // magic number "DULF"
	SectionCount  uint16               // number of sections
	SymbolCount   uint16               // number of symbols
	RelocCount    uint16               // number of relocations
	EntryPoint    dubcc.MachineAddress // entry point address
	SectionOffset uint32               // offset to section headers
	SymbolOffset  uint32               // offset to symbol table
	RelocOffset   uint32               // offset to relocation table
	StringOffset  uint32               // offset to string table
}

type SectionHeader struct {
	NameOffset uint32                   // offset in string table
	Type       DulfSection              // section type
	Flags      uint32                   // section flags
	Address    dubcc.MachineAddress     // virtual address
	Offset     uint32                   // file offset
	Size       uint32                   // size in bytes
	Link       uint32                   // link to another section
	Info       uint32                   // additional info
	Alignment  uint32                   // address alignment
}

type Symbol struct {
	NameOffset uint32                   // offset in string table
	Value      dubcc.MachineAddress     // symbol value
	Size       uint32                   // symbol size
	Info       uint8                    // symbol type and binding
	Other      uint8                    // reserved
	Section    uint16                   // section index
}

type Relocation struct {
	Offset     dubcc.MachineAddress // location to relocate
	Info       uint32               // relocation type and symbol index
	Addend     int64                // for relocation
	SymbolName string
}

type ObjectFile struct {
	Header      DulfHeader
	Sections    []Section
	Symbols     []Symbol
	Relocations []Relocation
	StringTable []byte
	stringMap   map[string]uint32
}

type Section struct {
	Header SectionHeader
	Data   []dubcc.MachineWord
	Name   string
}


func (s *Symbol) Type() SymbolType {
	return SymbolType(s.Info & 0xf)
}

func (s *Symbol) SetInfo(binding SymbolBinding, symType SymbolType) {
	s.Info = uint8(binding)<<4 | uint8(symType)
}

func (r *Relocation) SymbolIndex() uint32 {
	return r.Info >> 8
}

func (r *Relocation) RelocType() RelocationType {
	return RelocationType(r.Info & 0xff)
}

func (r *Relocation) SetInfo(symbolIndex uint32, relocType RelocationType) {
	r.Info = (symbolIndex << 8) | uint32(relocType)
}
