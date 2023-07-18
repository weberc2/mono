#ifndef STR_FORMATTER_H
#define STR_FORMATTER_H

#include "formatter.h"

typedef struct str_formatter
{
    str buffer;
    size_t cursor;
} str_formatter;

#define STR_FORMATTER(buf) \
    (str_formatter) { .buffer = buf, .cursor = 0 }

#define STR_FORMATTER_WITH_CAP(cap) \
    STR_FORMATTER(STR_ARR((char[cap]){0}))

formatter str_formatter_to_formatter(str_formatter *sf);

static inline str str_formatter_data(str_formatter *sf)
{
    return str_slice(sf->buffer, 0, sf->cursor);
}

static inline void str_formatter_reset(str_formatter *sf)
{
    sf->cursor = 0;
}

#endif // STR_FORMATTER_H