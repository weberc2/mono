#include "io/str_reader.h"

void str_reader_init(str_reader *sr, str buffer)
{
    sr->buffer = buffer;
    sr->cursor = 0;
}

size_t str_reader_read(str_reader *sr, str buffer)
{
    size_t n = str_copy_at(buffer, sr->buffer, sr->cursor);
    sr->cursor += n;
    return n;
}

size_t str_reader_io_read(
    str_reader *sr,
    str buffer,
    errors *errs)
{
    return str_reader_read(sr, buffer);
}

void str_reader_to_reader(str_reader *sr, reader *out)
{
    out->data = sr;
    out->read = (read_func)str_reader_io_read;
}