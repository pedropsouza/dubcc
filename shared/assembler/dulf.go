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
	StringTabSize uint32  // string table size in bytes
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
	Info       SymbolBinding        // symbol type and binding
	Other      uint8        				// reserved
	Section    uint16       				// section index
}

type Relocation struct {
	Offset     dubcc.MachineAddress // location to relocate
	Info       uint32               // relocation type and symbol index
	Addend     int64                // for relocation
}

type ObjectFile struct {
	Header      DulfHeader				// header
	Sections    []Section					// sections
	Symbols     []Symbol					// symbols
	Relocations []Relocation			// relocations
	StringTable []byte						// string table
	StringMap  	map[string]uint32	// string map
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
	s.Info = binding << 4 | SymbolBinding(symType)
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

func (r *Relocation) GetSymbolIndex() uint32 {
	return r.Info >> 8
}

func (r *Relocation) GetType() uint32 {
	return r.Info & 0xF
}

func (s *Symbol) GetBinding() SymbolBinding {
	return s.Info >> 4
}

func (s *Symbol) GetType() uint8 {
	return uint8(s.Info & 0xF)
}

func (info *Info) GenerateObjectFile() (*ObjectFile, error) {
	obj := &ObjectFile{
		// why is this a string to uint32
		StringMap: make(map[string]uint32),
	}
	
	obj.AddString("")
	
	textSection := Section{
		Name: ".text",
		Header: SectionHeader{
			Type:  SHT_PROGBITS,
			Flags: 0x6, // allocatable + executable
			Size:  uint32(len(info.GetOutput()) * 2), // 2 bytes per word
		},
		Data: info.GetOutput(),
	}
	textSection.Header.NameOffset = obj.AddString(".text")
	obj.Sections = []Section{textSection}
	
	obj.buildSymbolTable(info)
	obj.buildRelocationTable(info)
	
	obj.Header.Magic = [4]byte{'D', 'U', 'L', 'F'}
	obj.Header.SectionCount = uint16(len(obj.Sections))
	obj.Header.SymbolCount = uint16(len(obj.Symbols))
	obj.Header.RelocCount = uint16(len(obj.Relocations))
	obj.Header.StringTabSize = uint32(len(obj.StringTable))
	
	return obj, nil
}

func (obj *ObjectFile) AddString(s string) uint32 {
	if offset, exists := obj.StringMap[s]; exists {
		return offset
	}
	offset := uint32(len(obj.StringTable))
	obj.StringTable = append(obj.StringTable, []byte(s)...)
	obj.StringTable = append(obj.StringTable, 0) // null terminator
	obj.StringMap[s] = offset
	return offset
}

func (obj *ObjectFile) buildSymbolTable(info *Info) {
	// defined symbols
	for name, addr := range info.symbols {
		symbol := Symbol{
			NameOffset: obj.AddString(name),
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
				NameOffset: obj.AddString(externSym),
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
		}

		for i, sym := range obj.Symbols {
			if obj.GetString(sym.NameOffset) == link.name {
				reloc.SetInfo(uint32(i), R_ABSOLUTE)
				break
			}
		}
		obj.Relocations = append(obj.Relocations, reloc)
	}
}

func (obj *ObjectFile) GetString(offset uint32) string {
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
	headerSize := uint32(30) // fixed header size??
	sectionHeaderSize := uint32(len(obj.Sections) * 34) // 34 bytes per section header
	
	obj.Header.SectionOffset = headerSize
	obj.Header.SymbolOffset = obj.Header.SectionOffset + sectionHeaderSize
	obj.Header.RelocOffset = obj.Header.SymbolOffset + uint32(len(obj.Symbols)*14) // 20 bytes per symbol
	obj.Header.StringOffset = obj.Header.RelocOffset + uint32(len(obj.Relocations)*14) // 24 bytes per relocation
	
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

func Read(r io.Reader) (obj *ObjectFile, err error) {
	obj = &ObjectFile{}

	// header
	if err := binary.Read(r, binary.BigEndian, &obj.Header); err != nil {
		return nil, err
	}
	obj.StringTable = make([]byte, obj.Header.StringTabSize)
	// section headers
	for range obj.Header.SectionCount {
		section := Section{}
		if err := binary.Read(r, binary.BigEndian, &section.Header); err != nil {
			return nil, err
		}
		obj.Sections = append(obj.Sections, section)
	}
	// symbols
	for range obj.Header.SymbolCount {
		symbol := Symbol{}
		if err := binary.Read(r, binary.BigEndian, &symbol); err != nil {
			return nil, err
		}
		obj.Symbols = append(obj.Symbols, symbol)
	}
	// relocations
	for range obj.Header.RelocCount {
		reloc := Relocation{}
		if err := binary.Read(r, binary.BigEndian, &reloc.Offset); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.BigEndian, &reloc.Info); err != nil {
			return nil, err
		}
		if err := binary.Read(r, binary.BigEndian, &reloc.Addend); err != nil {
			return nil, err
		}
		obj.Relocations = append(obj.Relocations, reloc)
	}
	// string table
	if _, err := r.Read(obj.StringTable); err != nil {
		return nil, err
	}
	// section data
	for idx, section := range obj.Sections {
		for range (section.Header.Size/2) {
			var word dubcc.MachineWord
			if err := binary.Read(r, binary.BigEndian, &word); err != nil {
				return nil, err
			}
			obj.Sections[idx].Data = append(obj.Sections[idx].Data, word)
		}
	}

	return obj, nil
}
