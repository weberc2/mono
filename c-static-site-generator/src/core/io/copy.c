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
        size_t nr = reader_read(src, buf, res);
        if (nr < 1)
        {
            break;
        }
        size_t nw = writer_write(dst, str_slice(buf, 0, nr), res);
        written += nw;

        if (nr != nw)
        {
            *res = result_err(ERR_SHORT_WRITE);
            break;
        }
        if (!res->ok)
        {
            break;
        }
    }
    return written;
}