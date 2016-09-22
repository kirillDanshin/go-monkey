// +build !go1.4
#include <runtime.h>

void ·Get(int64 ret) {
	ret = g->goid;
	USED(&ret);
}
