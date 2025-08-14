br start
a: const 0xdead
b: const 0xbeef
c: const 0o31337
d: const 0b1001011101

start: load a
add b
push ACC
sub c
push ACC
push d
pop ACC
pop R1
pop R0
stop
