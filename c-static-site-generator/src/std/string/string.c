#include "core/str/str.h"
#include "std/string/string.h"
#include "core/math/math.h"

#include <stdlib.h> // free()
#include <string.h> // memcpy()

void string_init(string *s)
{
    s->data = NULL;
    s->len = 0;
    s->cap = 0;
}

string string_new()
{
    string s;
    string_init(&s);
    return s;
}

void string_drop(string *s)
{
    free(s->data);
    s->data = NULL;
    s->len = 0;
    s->cap = 0;
}

void string_grow(string *s, size_t at_least)
{
    char *old = s->data;
    s->cap = max(2 * s->cap, at_least);
    s->data = calloc(s->cap, 1);
    memcpy(s->data, old, s->len);
    free(old);
}

void string_push_raw(string *s, char *data, size_t len)
{
    if (s->cap - s->len < len)
    {
        string_grow(s, len);
    }
    memcpy(s->data + s->len, data, len);
    s->len += len;
}

void string_push_slice(string *s, str src)
{
    string_push_raw(s, src.data, src.len);
}

void string_reset(string *s)
{
    s->len = 0;
}

str string_borrow(string *s)
{
    return str_new(s->data, s->len);
}

str string_slice(string *s, size_t start, size_t end)
{
    return str_slice(string_borrow(s), start, end);
}

void string_copy_to_c(char *dst, string *s, size_t len)
{
    str_copy_to_c(dst, string_borrow(s), len);
}