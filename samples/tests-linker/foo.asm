; foo.assm
br main

extdef main
extdef bar

main:
    load 10   
    call bar
    stop
end
    

