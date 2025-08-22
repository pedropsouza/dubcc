; a forma mais orégano de se usar macros

MACRO
DATA_LIST
  X 1234 4
  X 67 2
  X 34 2
  X 0 1
  X 31337 5
  X 7 1
  X 888888 6
MEND

MACRO
  KEYS
  MACRO
    X K V
    const K
  MEND
  DATA_LIST
MEND

MACRO
  VALUES
  MACRO
    X K V
    const V
  MEND
  DATA_LIST
MEND

br begin

list_keys: KEYS
list_values: VALUES

MACRO
  XCOUNT
  MACRO
    X K V
    add 1
  MEND
  DATA_LIST
MEND

begin: XCOUNT
  sub list_values
  add list_keys
stop
