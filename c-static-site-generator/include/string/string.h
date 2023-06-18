#ifndef STRING_H
#define STRING_H

#include "str/str.h"
#include <stddef.h>

typedef struct
{
    char *data;
    size_t len;
    size_t cap;
} string;

void string_init(string *s);
void string_drop(string *s);
void string_push_raw(string *s, char *data, size_t len);
void string_push_slice(string *s, str src);
void string_borrow(string *s, str *out);
void string_slice(string *s, str *out, size_t start, size_t len);

#endif // STRING_H