#ifndef STRING_H
#define STRING_H

#include "core/str/str.h"
#include <stddef.h>

typedef struct
{
    char *data;
    size_t len;
    size_t cap;
} string;

void string_init(string *s);
string string_new();
void string_drop(string *s);
void string_push_raw(string *s, char *data, size_t len);
void string_push_slice(string *s, str src);
str string_borrow(string *s);
str string_slice(string *s, size_t start, size_t end);
void string_copy_to_c(char *dst, string *s, size_t len);

#endif // STRING_H