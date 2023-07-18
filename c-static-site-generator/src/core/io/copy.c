#include "core/io/err_eof.h"
#include "core/io/copy.h"
#include "core/str/str.h"

io_result copy(writer dst, reader src)
{
    char buffer[256];
    str buf = str_new(buffer, sizeof(buffer));
    return copy_buf(dst, src, buf);
}

io_result copy_buf(writer dst, reader src, str buf)
{
    size_t written = 0;
    while (true)
    {
        io_result read_res = reader_read(src, buf);
        if (read_res.size < 1)
        {
            return IO_RESULT(written, read_res.err);
        }

        io_result write_res = writer_write(
            dst,
            str_slice(buf, 0, read_res.size));
        written += write_res.size;

        if (read_res.size != write_res.size)
        {
            return IO_RESULT(written, ERR_SHORT_WRITE);
        }

        if (io_result_is_err(write_res))
        {
            if (error_is_eof(write_res.err))
            {
                return IO_RESULT_OK(written);
            }
            return IO_RESULT(written, write_res.err);
        }

        if (io_result_is_err(read_res))
        {
            if (error_is_eof(read_res.err))
            {
                return IO_RESULT_OK(written);
            }
            return IO_RESULT(written, read_res.err);
        }
    }

    return IO_RESULT_OK(written);
}