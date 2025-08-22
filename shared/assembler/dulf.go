package assembler

import (
	"dubcc"
	"encoding/binary"
	"io"
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
	SHT_NOBITS                        // uninitialized data (BSS) (not using yet)
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
	Magic         [4]byte	// magic number "DULF"
	SectionCount  uint16 	// number of sections
	SymbolCount   uint16 	// number of symbols
	RelocCount    uint16 	// number of relocations
	SectionOffset uint32 	// offset to section headers
	SymbolOffset  uint32 	// offset to symbol table
	RelocOffset   uint32 	// offset to relocation table
	StringOffset  uint32 	// offset to string table
}

type SectionHeader struct {
	NameOffset uint32           		// offset in string table
	Type       DulfSection  				// section type
	Flags      uint32       				// section flags
	Address    dubcc.MachineAddress	// virtual address
	Offset     uint32       				// file offset
	Size       uint32       				// size in bytes
	Link       uint32       				// link to another section
	Info       uint32       				// additional info
	Alignment  uint32       				// address alignment
}

type Symbol struct {
	NameOffset uint32      					// offset in string table
	Value      dubcc.MachineAddress	// symbol value
	Size       uint32       				// symbol size
	Info       uint8        				// symbol type and binding
	Other      uint8        				// reserved
	Section    uint16       				// section index
}

type Relocation struct {
	Offset     dubcc.MachineAddress // location to relocate
	Info       uint32               // relocation type and symbol index
	Addend     int64                // for relocation
	SymbolName string
}

type ObjectFile struct {
	Header      DulfHeader				// header
	Sections    []Section					// sections
	Symbols     []Symbol					// symbols
	Relocations []Relocation			// relocations
	StringTable []byte						// string table
	stringMap   map[string]uint32	// string map
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

func (info *Info) GenerateObjectFile() (*ObjectFile, error) {
	obj := &ObjectFile{
		// why is this a string to uint32
		stringMap: make(map[string]uint32),
	}
	
	obj.addString("")
	
	textSection := Section{
		Name: ".text",
		Header: SectionHeader{
			Type:  SHT_PROGBITS,
			Flags: 0x6, // allocatable + executable
			Size:  uint32(len(info.GetOutput()) * 2), // 2 bytes per word
		},
		Data: info.GetOutput(),
	}
	textSection.Header.NameOffset = obj.addString(".text")
	obj.Sections = []Section{textSection}
	
	obj.buildSymbolTable(info)
	obj.buildRelocationTable(info)
	
	obj.Header.Magic = [4]byte{'D', 'U', 'L', 'F'}
	obj.Header.SectionCount = uint16(len(obj.Sections))
	obj.Header.SymbolCount = uint16(len(obj.Symbols))
	obj.Header.RelocCount = uint16(len(obj.Relocations))
	
	return obj, nil
}

func (obj *ObjectFile) addString(s string) uint32 {
	if offset, exists := obj.stringMap[s]; exists {
		return offset
	}
	offset := uint32(len(obj.StringTable))
	obj.StringTable = append(obj.StringTable, []byte(s)...)
	obj.StringTable = append(obj.StringTable, 0) // null terminator
	obj.stringMap[s] = offset
	return offset
}

func (obj *ObjectFile) buildSymbolTable(info *Info) {
	// defined symbols
	for name, addr := range info.symbols {
		symbol := Symbol{
			NameOffset: obj.addString(name),
			Value:      addr,
			Size:       8, // default size
			Section:    0, // all symbols in .text section
		}
		
		// check if symbol is global
		if IsGlobalSymbol(name) {
			symbol.SetInfo(STB_GLOBAL, STT_NOTYPE)
		} else {
			symbol.SetInfo(STB_LOCAL, STT_NOTYPE)
		}
		
		obj.Symbols = append(obj.Symbols, symbol)
	}
	
	// external symbols as undefined
	for externSym := range externSymbols {
		if _, exists := info.symbols[externSym]; !exists {
			symbol := Symbol{
				NameOffset: obj.addString(externSym),
				Value:      0,     // Undefined
				Size:       0,
				Section:    0xFFF1, // SHN_UNDEF
			}
			symbol.SetInfo(STB_GLOBAL, STT_NOTYPE)
			obj.Symbols = append(obj.Symbols, symbol)
		}
	}
}

func (obj *ObjectFile) buildRelocationTable(info *Info) {
	for _, link := range info.undefSyms.links {
		reloc := Relocation{
			Offset:     link.from,
			Addend:     0,
			SymbolName: link.name,
		}

		for i, sym := range obj.Symbols {
			if obj.getString(sym.NameOffset) == link.name {
				reloc.SetInfo(uint32(i), R_ABSOLUTE)
				break
			}
		}
		obj.Relocations = append(obj.Relocations, reloc)
	}
}

func (obj *ObjectFile) getString(offset uint32) string {
	if offset >= uint32(len(obj.StringTable)) {
		return ""
	}
	end := offset
	for end < uint32(len(obj.StringTable)) && obj.StringTable[end] != 0 {
		end++
	}
	return string(obj.StringTable[offset:end])
}

func (obj *ObjectFile) Write(w io.Writer) error {
	headerSize := uint32(40) // fixed header size??
	sectionHeaderSize := uint32(len(obj.Sections) * 36) // 36 bytes per section header
	
	obj.Header.SectionOffset = headerSize
	obj.Header.SymbolOffset = obj.Header.SectionOffset + sectionHeaderSize
	obj.Header.RelocOffset = obj.Header.SymbolOffset + uint32(len(obj.Symbols)*20) // 20 bytes per symbol
	obj.Header.StringOffset = obj.Header.RelocOffset + uint32(len(obj.Relocations)*24) // 24 bytes per relocation
	
	// header
	if err := binary.Write(w, binary.BigEndian, obj.Header); err != nil {
		return err
	}
	// section headers
	for _, section := range obj.Sections {
		if err := binary.Write(w, binary.BigEndian, section.Header); err != nil {
			return err
		}
	}
	// symbols
	for _, symbol := range obj.Symbols {
		if err := binary.Write(w, binary.BigEndian, symbol); err != nil {
			return err
		}
	}
	// relocations
	for _, reloc := range obj.Relocations {
		if err := binary.Write(w, binary.BigEndian, reloc.Offset); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, reloc.Info); err != nil {
			return err
		}
		if err := binary.Write(w, binary.BigEndian, reloc.Addend); err != nil {
			return err
		}
	}
	// string table
	if _, err := w.Write(obj.StringTable); err != nil {
		return err
	}
	// section data
	for _, section := range obj.Sections {
		for _, word := range section.Data {
			if err := binary.Write(w, binary.BigEndian, word); err != nil {
				return err
			}
		}
	}
	
	return nil
}

func (obj *ObjectFile) ToMachineWordSlice() []dubcc.MachineWord {
    var words []dubcc.MachineWord
    for _, section := range obj.Sections {
        if section.Header.Flags == 0x6 {
            words = append(words, section.Data...)
        }
    }
    return words
}
