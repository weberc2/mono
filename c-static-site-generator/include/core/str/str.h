#ifndef STR_H
#define STR_H

#include <stddef.h>
#include <stdbool.h>

typedef struct
{
    char *data;
    size_t len;
} str;

void str_init(str *s, char *data, size_t len);
void str_slice(str s, str *out, size_t start, size_t end);
size_t str_copy(str dst, str src);
size_t str_copy_at(str dst, str src, size_t start);
bool str_eq(str lhs, str rhs);
size_t str_copy_to_c(char *dst, str src, size_t len);

typedef struct
{
    bool found;
    size_t index;
} str_find_result;

str_find_result str_find(str s, str search);

#define STR_NEW(var, data)       \
    char __src_##var[] = (data); \
    str(var);                    \
    str_init(&(var), __src_##var, sizeof(__src_##var) - 1);

#endif // STR_H