#ifndef STR_H
#define STR_H

#include <stddef.h>
#include <stdbool.h>

typedef struct
{
    char *data;
    size_t len;
} str;

str str_new(char *data, size_t len);
void str_init(str *s, char *data, size_t len);
str str_slice(str s, size_t start, size_t end);
size_t str_copy(str dst, str src);
size_t str_copy_at(str dst, str src, size_t start);
bool str_eq(str lhs, str rhs);
size_t str_copy_to_c(char *dst, str src, size_t len);

#define STR_NEW_ARR(s) str_new(s, sizeof(s))

#endif // STR_H