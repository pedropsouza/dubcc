package dubcc

import (
	"errors"
	"fmt"
)

// fazer uma (lista encadeada, pilha, fila) de erros, onde após o assemble terminar, chamar os erros que foram encontrados
type ErrorList struct {
	ErrName []error
}

var EmptyLineErr = errors.New("empty line")
var InvalidCharacter = errors.New("invalid character")
var Oi = errors.New("AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")

func PrintError(err error) {
	fmt.Println(err)

}
