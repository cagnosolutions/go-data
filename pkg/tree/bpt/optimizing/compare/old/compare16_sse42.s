// +build amd64

#include "textflag.h"

TEXT ·Compare16SSE42(SB), NOSPLIT, $0
    MOVDQU a+0(FP), X0   // Load 'a'
    MOVDQU b+16(FP), X1  // Load 'b'
    PCMPEQB X0, X1       // Compare each byte
    PMOVMSKB X0, CX      // Move result mask to CX
    CMPQ CX, $0xFFFF     // Check if all match
    JE equal

    PCMPGTB X1, X0       // Compare (a < b)
    PMOVMSKB X0, CX
    TESTQ CX, CX
    JNZ less

greater:
    MOVB $0, ret+32(FP)  // a > b
    RET

less:
    MOVB $1, ret+32(FP)  // a < b
    RET

equal:
    MOVB $0, ret+32(FP)  // Equal case
    RET
