#ifndef _MONKEY_H_
#define _MONKEY_H_

#include "js/jsapi.h"

/* Function pointers to avoid CGO warnning. */
extern JSClass            global_class;
extern JSErrorReporter    the_error_callback;
extern JSNative           the_go_func_callback;
extern JSNative           the_go_obj_func_callback;
extern JSPropertyOp       the_go_getter_callback;
extern JSStrictPropertyOp the_go_setter_callback;

/* File name for evaluate script. */
extern const char* eval_filename;

/* Fix CGO marco problem */
extern void  SET_RVAL(JSContext *cx, jsval* vp, jsval v);
extern jsval GET_ARGV(JSContext *cx, jsval* vp, int n);
extern jsval GET_NULL();
extern jsval GET_VOID();

#endif
