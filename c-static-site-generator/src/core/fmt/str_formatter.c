#include "core/fmt/str_formatter.h"

bool str_formatter_write_str(str_formatter *sf, str src)
{
    size_t nc = str_copy(sf->buffer, src);
    sf->cursor += nc;
    return nc == src.len;
}

formatter str_formatter_to_formatter(str_formatter *sf)
{
    return (formatter){
        .data = (void *)sf,
        .write_str = (formatter_write_str)str_formatter_write_str,
    };
}