#include "io/buffered_reader.h"
#include "io/io_result.h"
#include "math/math.h"

void buffered_reader_init(buffered_reader *br, reader source, str buf)
{
    br->source = source;
    br->buffer = buf;
    br->cursor = 0;
}

size_t buffered_reader_read(buffered_reader *br, str buf, io_result *res)
{
    size_t n;
    if (br->cursor > 0 && br->cursor < br->buffer.len)
    {
        str remaining;
        str_slice(br->buffer, &remaining, br->cursor, br->buffer.len);
        n = str_copy(buf, remaining);
        br->cursor += n;

        // if n >= buf.len, it means we had at least a whole `buf` left in the
        // buffer. If n == 0, it means we've reached the end of the file. In
        // either case, return.
        if (n >= buf.len || n < 1)
        {
            io_result_ok(res);
            return n;
        }
    }
    else
    {
        n = 0;
    }

    // otherwise, we only partially filled the output buffer and we need to
    // reload the internal buffer.
    while (n < buf.len)
    {
        br->cursor = 0;
        size_t nr = reader_read(br->source, br->buffer, res);

        // NB: we are deliberately *NOT* handling errors at this point--we
        // first want to copy anything we successfully read into the output
        // buffer.

        // if we don't read anything, it means we've reached eof--return with
        // a partially-written output buffer.
        if (nr < 1)
        {
            break;
        }

        // otherwise we read something; let's copy it to the unwritten portion
        // of the output buffer.
        str target;
        str_slice(buf, &target, n, min(buf.len, nr));
        size_t copied = str_copy(target, br->buffer);
        size_t unwritten = buf.len - n;
        n += copied;
        br->cursor += copied;

        // if we filled the output buffer OR encountered errors, return.
        if (copied >= unwritten || !res->ok)
        {
            break;
        }

        // otherwise loop around and refill the buffer
    }
    return n;
}