#include "parse_field_value.h"

parse_field_value_result parse_field_value(
    reader r,
    writer w,
    str buf,
    size_t cursor,
    size_t last_read_end)
{
    size_t size = 0;
    str_find_result find_res = (str_find_result){.found = false, .index = 0};
    while (true)
    {
        str gooddata = str_slice(buf, cursor, last_read_end);
        find_res = str_find_char(gooddata, '\n');
        str value = find_res.found
                        ? str_slice(gooddata, 0, find_res.index)
                        : gooddata;
        result res = result_new();
        size_t nw = writer_write(w, value, &res);
        size += nw;

        if (nw != value.len)
        {
            return PARSE_FIELD_VALUE_IO_ERROR(ERR_SHORT_WRITE);
        }

        if (!res.ok)
        {
            return PARSE_FIELD_VALUE_IO_ERROR(res.err);
        }

        if (find_res.found)
        {
            break;
        }

        last_read_end = reader_read(r, buf, &res);
        cursor = 0;
        if (last_read_end < 1)
        {
            return PARSE_FIELD_VALUE_OK(0);
        }

        if (!res.ok)
        {
            return PARSE_FIELD_VALUE_IO_ERROR(res.err);
        }
    }

    return PARSE_FIELD_VALUE_OK(find_res.index);
}