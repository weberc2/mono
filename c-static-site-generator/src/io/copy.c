#include "io/copy.h"
#include "str/str.h"

const char *const ERR_SHORT_WRITE = "short write";
const char *const ERR_INVALID_WRITE = "invalid write";

size_t copy(writer dst, reader src, errors *errs)
{
    char buffer[256];
    str buf;
    str_init(&buf, buffer, sizeof(buffer));
    buf.len = sizeof(buffer) - 1;

    size_t written;
    while (true)
    {
        size_t nr = reader_read(src, buf, errs);
        if (nr > 0)
        {

            str tmp;
            str_slice(buf, &tmp, 0, nr);

            size_t nw = writer_write(dst, tmp, errs);

            error err;
            if (nw < nr || nw < 0)
            {
                nw = 0;
                if (errors_len(errs) < 0)
                {
                    error_const(&err, ERR_INVALID_WRITE);
                    errors_push(errs, err);
                }
            }
            written += nw;

            if (errors_len(errs) > 0)
            {
                break;
            }

            if (nr != nw)
            {
                error_const(&err, ERR_SHORT_WRITE);
                errors_push(errs, err);
                break;
            }
            if (errors_len(errs) > 0)
            {
                break;
            }

            continue;
        }
        break;
    }
    return written;
}