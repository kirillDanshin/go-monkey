#include <stdio.h>
#include "monkey.h"
#include "_cgo_export.h"

/* File name for evaluate script. */
const char* eval_filename = "Eval()";

JSClass global_class = {
    "global", JSCLASS_GLOBAL_FLAGS,
    JS_PropertyStub, JS_PropertyStub, JS_PropertyStub, JS_StrictPropertyStub,
    JS_EnumerateStub, JS_ResolveStub, JS_ConvertStub, JS_FinalizeStub,
    JSCLASS_NO_OPTIONAL_MEMBERS
};

/* The error reporter callback. */
void error_callback(JSContext *cx, const char *message, JSErrorReport *report) {
	call_error_func(JS_GetRuntimePrivate(JS_GetRuntime(cx)), (char*)message, report);
}

/* The function callback. */
JSBool go_func_callback(JSContext *cx, uintN argc, jsval *vp) {
	JSObject *callee = JSVAL_TO_OBJECT(JS_CALLEE(cx, vp));

	jsval name;
	JS_GetProperty(cx, callee, "name", &name);

	char* cname = JS_EncodeString(cx, JS_ValueToString(cx, name));

	JSBool result = call_go_func(JS_GetRuntimePrivate(JS_GetRuntime(cx)), cname, argc, vp);

	JS_free(cx, (void*)cname);

	return result;
}

/* The object function callback. */
JSBool go_obj_func_callback(JSContext *cx, uintN argc, jsval *vp) {
	JSObject *callee = JSVAL_TO_OBJECT(JS_CALLEE(cx, vp));

	jsval name;
	JS_GetProperty(cx, callee, "name", &name);

	char* cname = JS_EncodeString(cx, JS_ValueToString(cx, name));

	JSBool result = call_go_obj_func(JS_GetPrivate(cx, JS_THIS_OBJECT(cx, vp)), cname, argc, vp);

	JS_free(cx, (void*)cname);

	return result;
}

/* The property getter callback */
JSBool go_getter_callback(JSContext *cx, JSObject *obj, jsid id, jsval *vp) {
	char* cname = JS_EncodeString(cx, JSID_TO_STRING(id));

	JSBool result = call_go_getter(JS_GetPrivate(cx, obj), cname, vp);

	JS_free(cx, (void*)cname);

	return result;
}

/* The property setter callback */
JSBool go_setter_callback(JSContext *cx, JSObject *obj, jsid id, JSBool strict, jsval *vp) {
	char* cname = JS_EncodeString(cx, JSID_TO_STRING(id));

	JSBool result = call_go_setter(JS_GetPrivate(cx, obj), cname, vp);

	JS_free(cx, (void*)cname);

	return result;
}

/* Fix CGO marco problem */
void SET_RVAL(JSContext *cx, jsval* vp, jsval v) {
	JS_SET_RVAL(cx, vp, v);
}

jsval GET_ARGV(JSContext *cx, jsval* vp, int n) {
	return JS_ARGV(cx, vp)[n];
}

jsval GET_NULL() {
	return JSVAL_NULL;
}

jsval GET_VOID() {
	return JSVAL_VOID;
}

/* Function pointers to avoid CGO warnning. */
JSErrorReporter    the_error_callback = &error_callback;
JSNative           the_go_func_callback = &go_func_callback;
JSNative           the_go_obj_func_callback = &go_obj_func_callback;
JSPropertyOp       the_go_getter_callback = &go_getter_callback;
JSStrictPropertyOp the_go_setter_callback = &go_setter_callback;
