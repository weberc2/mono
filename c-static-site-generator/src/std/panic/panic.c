#include <stdio.h>
#include <stdarg.h>
#include <stdlib.h>

void panic(char *fmt, ...)
{
    va_list args;
    va_start(args, fmt);
    vfprintf(stderr, fmt, args);
    va_end(args);
    abort();
}