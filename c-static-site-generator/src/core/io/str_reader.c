#include "core/io/str_reader.h"
#include "core/result/result.h"

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

size_t str_reader_io_read(str_reader *sr, str buffer, result *res)
{
    *res = result_ok();
    return str_reader_read(sr, buffer);
}

reader str_reader_to_reader(str_reader *sr)
{
    return reader_new((void *)sr, (read_func)str_reader_io_read);
}