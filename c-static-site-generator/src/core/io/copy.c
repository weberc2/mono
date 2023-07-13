#include "core/io/copy.h"
#include "core/result/result.h"
#include "core/str/str.h"

size_t copy(writer dst, reader src, result *res)
{
    char buffer[256];
    str buf = str_new(buffer, sizeof(buffer));
    return copy_buf(dst, src, buf, res);
}

size_t copy_buf(writer dst, reader src, str buf, result *res)
{
    size_t written = 0;
    while (true)
    {
        io_result read_res = reader_read(src, buf);
        res->err = read_res.err;
        res->ok = io_result_is_ok(read_res);
        if (read_res.size < 1)
        {
            break;
        }
        size_t nw = writer_write(dst, str_slice(buf, 0, read_res.size), res);
        written += nw;

        if (read_res.size != nw)
        {
            *res = result_err(ERR_SHORT_WRITE);
            break;
        }
        if (io_result_is_err(read_res))
        {
            break;
        }
    }
    return written;
}