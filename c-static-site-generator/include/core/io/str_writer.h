#ifndef STR_WRITER_H
#define STR_WRITER_H

#include "core/str/str.h"
#include "writer.h"

typedef struct str_writer
{
    str buffer;
    size_t cursor;
} str_writer;

#define STR_WRITER(b) \
    (str_writer) { .buffer = (b), .cursor = 0 }
#define STR_WRITER_WITH_CAP(cap) STR_WRITER(STR_ARR((char[cap]){0}))

#define STR_WRITER_TO_WRITER(sw) WRITER(sw, str_writer_io_write)

io_result str_writer_io_write(str_writer *sw, str buf);
writer str_writer_to_writer(str_writer *sw);

static inline str str_writer_data(str_writer *sw)
{
    return str_slice(sw->buffer, 0, sw->cursor);
}

#endif // STR_WRITER_H