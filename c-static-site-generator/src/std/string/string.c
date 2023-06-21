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

void string_from_slice(string *s, str src)
{
    string_init(s);
    string_push_slice(s, src);
}

void string_from_raw(string *s, char *data, size_t len)
{
    string_init(s);
    string_push_raw(s, data, len);
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

void string_borrow(string *s, str *out)
{
    out->data = s->data;
    out->len = s->len;
}

void string_slice(string *s, str *out, size_t start, size_t end)
{
    string_borrow(s, out);
    str_slice(*out, out, start, end);
}

void string_copy_to_c(char *dst, string *s, size_t len)
{
    str tmp;
    str_copy_to_c(dst, tmp, len);
}