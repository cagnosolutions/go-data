// +build amd64

#include "textflag.h"

TEXT ·Compare16AVX2(SB), NOSPLIT, $0
    VMOVDQU a+0(FP), X0   // Load 16 bytes from 'a'
    VMOVDQU b+16(FP), X1  // Load 16 bytes from 'b'
    VPCMPEQB X0, X1, X2   // Compare each byte (set bits if equal)
    VPMOVMSKB X2, CX      // Move result mask to CX
    CMPQ CX, $0xFFFF      // If all bits match, keys are equal
    JE equal

    VPCMPUB X0, X1, X2, 2 // Compare unsigned bytes (a < b)
    VPMOVMSKB X2, CX
    TESTQ CX, CX
    JNZ less

greater:
    MOVB $0, ret+32(FP)   // a > b
    RET

less:
    MOVB $1, ret+32(FP)   // a < b
    RET

equal:
    MOVB $0, ret+32(FP)   // Equal case
    RET
