#ifndef STR_H
#define STR_H

#include <stddef.h>
#include <stdbool.h>

typedef struct
{
    char *data;
    size_t len;
} str;

void str_init(str *str, char *src);
size_t str_copy(str dst, str src);
void str_slice(str *out, str str, size_t start, size_t end);
bool str_eq(str lhs, str rhs);

#endif // STR_H