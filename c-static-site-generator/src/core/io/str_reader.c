#include "core/io/err_eof.h"
#include "core/io/str_reader.h"

str_reader str_reader_new(str buffer)
{
    return (str_reader){
        .buffer = buffer,
        .cursor = 0,
    };
}

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

io_result str_reader_io_read(str_reader *sr, str buffer)
{
    size_t nr = str_reader_read(sr, buffer);
    return IO_RESULT(
        nr,
        sr->cursor < sr->buffer.len ? ERROR_NULL : ERR_EOF);
}

reader str_reader_to_reader(str_reader *sr)
{
    return reader_new((void *)sr, (read_func)str_reader_io_read);
}