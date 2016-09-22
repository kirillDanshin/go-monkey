// +build amd64 amd64p32 arm 386
// +build go1.4,!go1.5

#include "textflag.h"

#ifdef GOARCH_arm
#define JMP B
#endif

TEXT ·getg(SB),NOSPLIT,$0-0
	JMP	runtime·getg(SB)
