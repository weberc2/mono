#include "io/buffered_reader.h"

void buffered_reader_init(buffered_reader *br, reader source, str buf)
{
    br->source = source;
    br->buffer = buf;
    br->cursor = 0;
}

size_t buffered_reader_read(buffered_reader *br, str buf, errors *errs)
{
    str remaining;
    str_slice(br->buffer, &remaining, br->cursor, br->buffer.len);
    size_t n = str_copy(buf, remaining);

    // if n >= buf.len, it means we had at least a whole `buf` left in the
    // buffer. If n == 0, it means we've reached the end of the file. In either
    // case, return.
    if (n >= buf.len || n < 1)
    {
        return n;
    }

    // otherwise, we only partially filled the buffer and we need to reload the
    // buffer.
    while (n < buf.len)
    {
        br->cursor = 0;
        size_t nr = reader_read(br->source, br->buffer, errs);

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
        size_t copied = str_copy_at(buf, br->buffer, n);
        size_t unwritten = buf.len - n;
        n += copied;

        // if we filled the output buffer OR encountered errors, return.
        if (copied >= unwritten || errors_len(errs) > 0)
        {
            break;
        }

        // otherwise loop around and refil the buffer
    }
    return n;
}