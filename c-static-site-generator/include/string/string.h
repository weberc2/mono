#ifndef STRING_H
#define STRING_H

#include <stdlib.h>
#include "str.h"

#define STRING_DEFAULT_CAP 100

typedef struct
{
    char *data;
    size_t cap;
    size_t len;
} string;

void string_init(string *s);
void string_init_with_cap(string *s, size_t cap);
void string_from_str(string *s, str str);
void string_from_raw(string *s, char *src);
void string_drop(string *s);
void string_borrow(string *lhs, str *out);
void string_slice(string *s, str *out, size_t start, size_t end);
bool string_eq(string *lhs, string *rhs);
void string_append_raw(string *lhs, char *rhs, size_t len);
void string_append_str(string *lhs, str rhs);

#endif // STRING_H