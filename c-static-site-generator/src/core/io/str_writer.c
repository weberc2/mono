#include "core/io/str_writer.h"
#include "core/io/err_eof.h"

io_result str_writer_io_write(str_writer *sw, str buf)
{
    str dst = str_slice(sw->buffer, sw->cursor, sw->buffer.len);
    size_t nc = str_copy(dst, buf);
    sw->cursor += nc;
    return IO_RESULT(nc, nc < 1 ? ERR_EOF : ERROR_NULL);
}

writer str_writer_to_writer(str_writer *sw)
{
    return WRITER(sw, str_writer_io_write);
}