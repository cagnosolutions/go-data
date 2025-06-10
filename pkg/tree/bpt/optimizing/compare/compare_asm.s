#include "textflag.h"

// func CompareASM(a, b [16]byte) int
TEXT ·CompareASM(SB), NOSPLIT, $0
    MOVQ a+0(FP), R8    // Load first 8 bytes of a
    MOVQ a+8(FP), R9    // Load second 8 bytes of a
    MOVQ b+0(FP), R10   // Load first 8 bytes of b
    MOVQ b+8(FP), R11   // Load second 8 bytes of b

    CMPQ R8, R10        // Compare first 8 bytes
    JL   less           // Jump if a < b
    JG   greater        // Jump if a > b

    CMPQ R9, R11        // Compare second 8 bytes
    JL   less           // Jump if a < b
    JG   greater        // Jump if a > b

    MOVQ $0, AX         // If equal, return 0
    RET

less:
    MOVQ $-1, AX        // If a < b, return -1
    RET

greater:
    MOVQ $1, AX         // If a > b, return 1
    RET
