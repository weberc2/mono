#include "core/io/copy.h"
#include "core/result/result.h"
#include "core/str/str.h"

size_t copy(writer dst, reader src, result *res)
{
    char buffer[256];
    str buf;
    str_init(&buf, buffer, sizeof(buffer));
    buf.len = sizeof(buffer);

    size_t written;
    while (true)
    {
        size_t nr = reader_read(src, buf, res);
        if (nr > 0)
        {

            str tmp;
            str_slice(buf, &tmp, 0, nr);

            size_t nw = writer_write(dst, tmp, res);

            error err;
            if (nw < nr || nw < 0)
            {
                nw = 0;
                if (!res->ok)
                {
                    result_err(res, ERR_INVALID_WRITE);
                }
            }
            written += nw;

            if (!res->ok)
            {
                break;
            }

            if (nr != nw)
            {
                result_err(res, ERR_SHORT_WRITE);
                break;
            }
            if (!res->ok)
            {
                break;
            }

            continue;
        }
        break;
    }
    return written;
}