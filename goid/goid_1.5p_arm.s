// +build arm
// +build go1.5

#include "textflag.h"

// func getg() uintptr
TEXT ·getg(SB),NOSPLIT,$0-8
	MOVW g, ret+0(FP)
	RET
