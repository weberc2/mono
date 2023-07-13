#ifndef FMT_ARG_H
#define FMT_ARG_H

#include "core/str/str.h"

typedef size_t (*format_func)(void *, str);

typedef struct
{
    void *data;
    format_func format;
} fmt_arg;

size_t fmt_arg_format(fmt_arg arg, str buf);
static inline bool fmt_arg_is_null(fmt_arg arg)
{
    return arg.data == NULL && arg.format == NULL;
}

typedef struct
{
    fmt_arg *data;
    size_t len;
} fmt_args;

#define FMT_ARGS(...)                                              \
    (fmt_args)                                                     \
    {                                                              \
        .data = (fmt_arg[]){__VA_ARGS__},                          \
        .len = sizeof((fmt_arg[]){__VA_ARGS__}) / sizeof(fmt_arg), \
    }

fmt_arg fmt_args_first(fmt_args *args);
fmt_arg fmt_args_pop(fmt_args *args);

typedef struct
{
    str buffer;
    size_t cursor;
} fmt_arg_str;

size_t fmt_arg_str_format(fmt_arg_str *fas, str buf);

#define FMT_STR(s)                                          \
    (fmt_arg)                                               \
    {                                                       \
        .data = &(fmt_arg_str){.buffer = (s), .cursor = 0}, \
        .format = (format_func)fmt_arg_str_format,          \
    }

#define FMT_STR_LIT(s) FMT_STR(STR(s))

#endif // FMT_ARG_H