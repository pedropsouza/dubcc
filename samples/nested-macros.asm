MACRO
  ASSOCIATED U
  space
  const U
MEND

MACRO
OUTER
  MACRO
    INNER A B C
    const A
    const B
    const C
  MEND
  const 2
  ASSOCIATED 6
MEND

OUTER
INNER 1 2 3

