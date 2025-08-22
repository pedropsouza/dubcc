package linker

import (
	"dubcc"
	"dubcc/assembler"
	"fmt"
	"sort"
	"github.com/k0kubun/pp/v3"
)

type ObjectFile = assembler.ObjectFile
type MachineAddress = dubcc.MachineAddress
type MachineWord = dubcc.MachineWord
type Section = assembler.Section
type SectionHeader = assembler.SectionHeader
type Symbol = assembler.Symbol
type SymbolBinding = assembler.SymbolBinding
type Relocation = assembler.Relocation

const STT_FUNC = assembler.STT_FUNC
const SHT_PROGBITS = assembler.SHT_PROGBITS

type LinkerMode int

const (
	Relocator LinkerMode = iota
	Absolute
)

const (
	R_ABSOLUTE = 1
)

type Linker struct {
	Mode					LinkerMode
	LoadAddress		MachineAddress
	Objects       []*ObjectFile
	Executable    *ObjectFile
	SectionMap    map[string]*LinkedSection // section key -> linked section
	SymbolMap     map[string]*LinkedSymbol  // symbol name -> resolved symbol
	SectionLayout []SectionInfo             // ordered list of sections with addresses
}

type LinkedSection struct {
	Section      *Section
	ObjectIndex  int
	BaseAddress  MachineAddress // relative to start of executable
	AbsAddress   MachineAddress // absolute address if in absolute mode
	Size         uint32
	SectionIndex int 						// index in final executable
}

type LinkedSymbol struct {
	Symbol      *Symbol
	ObjectIndex int
	RelAddress  MachineAddress // relative address within executable
	AbsAddress  MachineAddress // absolute address if in absolute mode
	Section     string
}

type SectionInfo struct {
	Name        string
	RelAddress  MachineAddress
	AbsAddress  MachineAddress
	Size        uint32
	ObjectIndex int
}

func MakeRelocatorLinker() *Linker {
	return &Linker{
		Mode:	Relocator,
	}
}

func MakeAbsoluteLinker(loadAddress MachineAddress) *Linker {
	return &Linker{
		Mode:        Absolute,
		LoadAddress: loadAddress,
	}
}

func (linker *Linker) GenerateExecutable(objects []*ObjectFile) (*ObjectFile, error) {
	// layout sections and build symbol table
	linker.Objects = objects
	linker.SectionMap = make(map[string]*LinkedSection)
	linker.SymbolMap = make(map[string]*LinkedSymbol)

	if err := linker.firstPass(); err != nil {
		return nil, fmt.Errorf("failed at pass 1: %v", err)
	}

	if err := linker.secondPass(); err != nil {
		return nil, fmt.Errorf("failed at pass 2: %v", err)
	}

	return linker.Executable, nil
}

func (linker *Linker) firstPass() error {
	// calculate section layout
	if err := linker.calculateSectionLayout(); err != nil {
		return err
	}
	// build global symbol table with addresses
	if err := linker.buildGlobalSymbolTable(); err != nil {
		return err
	}
	// verify all external references can be resolved
	if err := linker.verifySymbolResolution(); err != nil {
		return err
	}

	return nil
}

func (linker *Linker) calculateSectionLayout() error {
	var currentRelAddress MachineAddress = 0
	var currentAbsAddress MachineAddress = linker.LoadAddress

	// process .text sections in order
	for objIdx, obj := range linker.Objects {
		for _, section := range obj.Sections {
			if section.Name == ".text" {
				sectionInfo := SectionInfo{
					Name:					section.Name,
					RelAddress: 	currentRelAddress,
					AbsAddress: 	currentAbsAddress,
					Size:					section.Header.Size,
					ObjectIndex: 	objIdx,
				}
				linker.SectionLayout = append(linker.SectionLayout, sectionInfo)

				// section mapping
				sectionKey := fmt.Sprintf("%d:.text", objIdx)
				linker.SectionMap[sectionKey] = &LinkedSection{
					Section:			&section,
					ObjectIndex:  objIdx,
					BaseAddress:  currentRelAddress, 						// relative to start of executable
					AbsAddress:   currentAbsAddress, 						// absolute address if in absolute mode
					Size:        	section.Header.Size,
					SectionIndex:	len(linker.SectionLayout) - 1,	// index in final executable
				}

				// advance addresses
				sectionSizeWords := (section.Header.Size + 1) / 2 * 2 // word boundary
				currentRelAddress += MachineAddress(sectionSizeWords)
				currentAbsAddress += MachineAddress(sectionSizeWords)
			}
		}
	}

	return nil
}

func (linker *Linker) buildGlobalSymbolTable() error {
	// collect all defined symbols
	for objIdx, obj := range linker.Objects {
		// find the section base address for the object
		sectionKey := fmt.Sprintf("%d:.text", objIdx)
		linkedSection, exists := linker.SectionMap[sectionKey]
		if !exists {
			continue // has no .text section
		}
		for _, symbol := range obj.Symbols {
			symbolName := obj.GetString(symbol.NameOffset)

			// skip empty symbol names
			if symbolName == "" {
				continue
			}
			// only process defined symbols (not external/undefined)
			if symbol.Section != 0xFFF1 { // TODO: what is the undefsym code?
				// collect all defined symbols with their resolved address
				if existing, exists := linker.SymbolMap[symbolName]; exists {
					return fmt.Errorf("symbol '%s' defined in multiple objects (obj %d and obj %d)", 
						symbolName, existing.ObjectIndex, objIdx)
				}

				relAddress := linkedSection.BaseAddress/2 + symbol.Value
				absAddress := linkedSection.AbsAddress/2 + symbol.Value
        
				fmt.Printf(
`Symbol %s got
relative address = %d + %d = %d and 
absolute address = %d + %d = %d
`,
				symbolName,
				linkedSection.BaseAddress/2, symbol.Value, 
				linkedSection.BaseAddress/2 + symbol.Value,
        linkedSection.AbsAddress/2, symbol.Value,
				linkedSection.AbsAddress/2 + symbol.Value,
			  )

				linker.SymbolMap[symbolName] = &LinkedSymbol{
					Symbol:      &symbol,
					ObjectIndex: objIdx,
					RelAddress:  relAddress,
					AbsAddress:  absAddress,
					Section:     ".text",
				}
			}
		}
	}

	return nil
}

func (linker *Linker) verifySymbolResolution() error {
	// check that all undefined symbols can be resolved
	for objIdx, obj := range linker.Objects {
		for _, symbol := range obj.Symbols {
			symbolName := obj.GetString(symbol.NameOffset)

			// skip empty symbol names
			if symbolName == "" {
				continue
			}

			// check undefined symbols
			if symbol.Section == 0xFFF1 { // TODO: what is the undefsym code?
				if _, exists := linker.SymbolMap[symbolName]; !exists {
					return fmt.Errorf("undefined symbol '%s' referenced in object %d", symbolName, objIdx)
				}
			}
		}
	}

	return nil
}

func (linker *Linker) secondPass() error {
	// merge structures
	if  err := linker.createExecutableStructure(); err != nil {
		return err
	}

	pp.Print(linker.SectionMap)
	// apply relocations
	if  err := linker.applyRelocations(); err != nil {
		return err
	}
	// build final symbol table and header
	if  err := linker.finalizeBinary(); err != nil {
		return err
	}

	return nil
}

func (linker *Linker) createExecutableStructure() error {
	linker.Executable = &ObjectFile{
		StringMap: make(map[string]uint32),
	}
	linker.Executable.AddString("") // empty string at offset 0

	// merge all .text sections
	var mergedTextData []MachineWord
	totalSize := uint32(0)

	for _, sectionInfo := range linker.SectionLayout {
		obj := linker.Objects[sectionInfo.ObjectIndex]
		for _, section := range obj.Sections {
			if section.Name == ".text" {
				mergedTextData = append(mergedTextData, section.Data...)
				totalSize += section.Header.Size
				break
			}
		}
	}

	// create merged .text section
	mergedSection := Section{
		Name: ".text",
		Header: SectionHeader{
			Type:    SHT_PROGBITS,
			Flags:   0x6, // allocatable + executable
			Size:    totalSize,
			Address: 0,
			Offset:  0,
		},
		Data: mergedTextData,
	}

	// set section address based on linking mode
	switch linker.Mode {
	case Absolute:
		mergedSection.Header.Address = linker.LoadAddress
	case Relocator:
		mergedSection.Header.Address = 0 // relocatable
	}

	mergedSection.Header.NameOffset = linker.Executable.AddString(".text")
	linker.Executable.Sections = []Section{mergedSection}

	return nil

}

func (linker *Linker) applyRelocations() error {
	// process relocations from each object file
	var currentDataOffset uint32 = 0

	for objIdx, obj := range linker.Objects {
		// calculate where this object's data starts in the merged section
			if objIdx > 0 {
				for i := range objIdx {
				for _, section := range linker.Objects[i].Sections {
					currentDataOffset += section.Header.Size / 2
				}
			}
		}

		for _, reloc := range obj.Relocations {
			// find target symbol
			symbi := reloc.GetSymbolIndex()
			symbName := obj.GetString(obj.Symbols[symbi].NameOffset)
			linkedSymbol, exists := linker.SymbolMap[symbName]
			if !exists {
				return fmt.Errorf("cannot resolve relocation for symbol '%s'", symbName)
			}

			// calculate relocation position in merged data
			relocPosition := MachineAddress(currentDataOffset) + reloc.Offset
			
			if relocPosition >= MachineAddress(len(linker.Executable.Sections[0].Data)) {
				return fmt.Errorf("relocation position %d out of bounds (max %d)", 
					relocPosition, len(linker.Executable.Sections[0].Data))
			}

			// apply relocation based on linking mode
			var targetAddress MachineAddress
			switch linker.Mode {
			case Absolute:
				targetAddress = linkedSymbol.AbsAddress
			case Relocator:
				targetAddress = linkedSymbol.RelAddress
			}

			// apply the relocation
			switch reloc.GetType() {
			case R_ABSOLUTE:
				linker.Executable.Sections[0].Data[relocPosition] = uint16(targetAddress & 0xFFFF)
			default:
				return fmt.Errorf("unsupported relocation type: %d", reloc.GetType())
			}
		}
	}

	return nil
}


func (linker *Linker) finalizeBinary() error {
	// build final symbol table
	var finalSymbols []Symbol
	
	// sort symbols by address
	var sortedSymbols []*LinkedSymbol
	for _, linkedSym := range linker.SymbolMap {
		sortedSymbols = append(sortedSymbols, linkedSym)
	}
	
	sort.Slice(sortedSymbols, func(i, j int) bool {
		return sortedSymbols[i].RelAddress < sortedSymbols[j].RelAddress
	})

	for _, linkedSym := range sortedSymbols {
		symbolName := linker.Objects[linkedSym.ObjectIndex].GetString(linkedSym.Symbol.NameOffset)
		
		// include all defined symbols
		var symbolValue MachineAddress
		switch linker.Mode {
		case Absolute:
			symbolValue = linkedSym.AbsAddress
		case Relocator:
			symbolValue = linkedSym.RelAddress
		}

		finalSym := Symbol{
			NameOffset: linker.Executable.AddString(symbolName),
			Value:      symbolValue,
			Size:       linkedSym.Symbol.Size,
			Section:    0, // everything in .text
		}

		binding := linkedSym.Symbol.GetBinding()
		finalSym.SetInfo(binding, STT_FUNC)
		finalSymbols = append(finalSymbols, finalSym)
	}

	linker.Executable.Symbols = finalSymbols

	if linker.Mode == Relocator {
		linker.Executable.Relocations = []Relocation{}
	} else {
		linker.Executable.Relocations = []Relocation{}
	}

	linker.Executable.Header.Magic = [4]byte{'D', 'U', 'L', 'F'}
	linker.Executable.Header.SectionCount = uint16(len(linker.Executable.Sections))
	linker.Executable.Header.SymbolCount = uint16(len(linker.Executable.Symbols))
	linker.Executable.Header.RelocCount = uint16(len(linker.Executable.Relocations))

	return nil
}
