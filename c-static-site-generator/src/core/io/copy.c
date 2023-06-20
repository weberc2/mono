#include "core/io/copy.h"
#include "core/result/result.h"
#include "core/str/str.h"

const char *const ERR_SHORT_WRITE = "short write";
const char *const ERR_INVALID_WRITE = "invalid write";

size_t copy(writer dst, reader src, result *res)
{
    char buffer[256];
    str buf;
    str_init(&buf, buffer, sizeof(buffer));
    buf.len = sizeof(buffer) - 1;

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
                    error_const(&err, ERR_INVALID_WRITE);
                    result_err(res, err);
                }
            }
            written += nw;

            if (!res->ok)
            {
                break;
            }

            if (nr != nw)
            {
                error_const(&err, ERR_SHORT_WRITE);
                result_err(res, err);
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