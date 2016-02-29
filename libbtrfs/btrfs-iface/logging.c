#define _GNU_SOURCE
#include <stdio.h>
#include <stdarg.h>

typedef void (*log_callback_func)(char *);

static log_callback_func log_error_callback;
static log_callback_func log_info_callback;

void set_log_info_callback(log_callback_func func)
{
	log_info_callback = func;
}

void set_log_error_callback(log_callback_func func)
{
	log_error_callback = func;
}

int __wrap_printf(const char *__restrict format, ...)
{
	va_list argptr;
	va_start(argptr, format);
	char *msg = NULL;
	vasprintf(&msg, format, argptr);
	log_info_callback(msg);
}

int __wrap_sprintf(char *__restrict s, const char *__restrict format, ...)
{
	va_list argptr;
	va_start(argptr, format);
	char *msg = NULL;

	//vasprintf();
}