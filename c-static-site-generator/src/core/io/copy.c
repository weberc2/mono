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
        io_result write_res = writer_write(
            dst,
            str_slice(buf, 0, read_res.size));
        written += write_res.size;

        if (read_res.size != write_res.size)
        {
            *res = result_err(ERR_SHORT_WRITE);
            break;
        }

        if (io_result_is_err(write_res))
        {
            res->err = write_res.err;
            res->ok = false;
            break;
        }

        if (io_result_is_err(read_res))
        {
            res->err = read_res.err;
            res->ok = false;
            break;
        }
    }

    return written;
}