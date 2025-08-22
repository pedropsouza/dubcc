br start

fac: sub 1
brzero fac.base
add 1
push ACC
sub 1
call fac
pop R1
mult R1
ret
fac.base: add 1
ret

start: load 5
call fac
