#include "io/byteslice_reader.h"

void byteslice_reader_init(byteslice_reader *br, byteslice buffer)
{
    br->buffer = buffer;
    br->cursor = 0;
}

size_t byteslice_reader_read(byteslice_reader *br, byteslice buffer)
{
    size_t n = byteslice_copy_at(buffer, br->buffer, br->cursor);
    br->cursor += n;
    return n;
}

size_t byteslice_reader_io_read(
    byteslice_reader *br,
    byteslice buffer,
    errors *errs)
{
    return byteslice_reader_read(br, buffer);
}

void byteslice_reader_to_reader(byteslice_reader *br, reader *out)
{
    out->data = br;
    out->read = (read_func)byteslice_reader_io_read;
}