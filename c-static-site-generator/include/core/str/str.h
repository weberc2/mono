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
bool str_has_prefix(str s, str prefix);
size_t str_copy_to_c(char *dst, str src, size_t len);
str str_trim_left(str s, str cutset);
str str_trim_right(str s, str cutset);
str str_trim(str s, str cutset);
str str_trim_space_left(str s);
str str_trim_space_right(str s);
str str_trim_space(str s);
str str_put_int(str s, int i);

typedef struct
{
    bool found;
    size_t index;
} str_find_result;

str_find_result str_find(str src, str match);
str_find_result str_find_char(str src, char match);

#define STR_ARR(s) \
    (str) { .data = (s), .len = sizeof(s) }
#define STR_LIT(s) \
    (str) { .data = (s), .len = sizeof(s) - 1 }

#endif // STR_H