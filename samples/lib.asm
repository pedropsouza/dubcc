; acc = acc % memsize
bound: brneg bound.ndec
sub 4096
brpos bound.pdec
add 4096
ret
bound.ndec: sub 4096
bound.pdec: br bound

; cons arg1 arg2 clobbers ACC
cons: pop ACC

