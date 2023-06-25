#include "core/math/math.h"
#include "core/result/result.h"
#include "core/io/buffered_reader.h"
#include "core/io/match_reader.h"

void buffered_reader_init(buffered_reader *br, reader source, str buf)
{
    br->source = source;
    br->buffer = buf;
    br->cursor = 0;
    br->read_end = 0;
}

buffered_reader buffered_reader_new(reader source, str buf)
{
    buffered_reader br;
    buffered_reader_init(&br, source, buf);
    return br;
}

size_t buffered_reader_read(buffered_reader *br, str buf, result *res)
{
    size_t n = 0;
    if (br->cursor > 0 && br->cursor < br->read_end)
    {
        str remaining = str_slice(br->buffer, br->cursor, br->read_end);
        n = str_copy(buf, remaining);
        br->cursor += n;

        // if n >= buf.len, it means we had at least a whole `buf` left in the
        // buffer. If n == 0, it means we've reached the end of the file. In
        // either case, return.
        if (n >= buf.len || n < 1)
        {
            *res = result_ok();
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

        // if we didn't read anything because we reached eof, we don't want to
        // reset the `read_end` back to the beginning of the buffer.
        br->read_end = nr;

        // if we didn't read anything because we reached eof, we don't want to
        // reset the `cursor` back to the beginning of the buffer.
        br->cursor = 0;

        // otherwise we read something; let's copy it to the unwritten portion
        // of the output buffer.
        str target = str_slice(buf, n, min(buf.len, nr));
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

size_t buffered_reader_refresh(buffered_reader *br, result *res)
{
}

bool buffered_reader_find(
    buffered_reader *br,
    writer w,
    str match,
    result *res)
{
    char buf_[256] = {0};
    str buf = str_new(buf_, sizeof(buf_));

    match_reader mr = match_reader_new(br, match);
    while (true)
    {
        size_t nr = match_reader_read(&mr, buf, res);
        if (nr < 1)
        {
            return true;
        }

        size_t nw = writer_write(w, str_slice(buf, 0, nr), res);
        if (nr != nw)
        {
            *res = result_err(ERR_SHORT_WRITE);
        }

        // if there was an error, return early
        if (!res->ok)
        {
            return false;
        }
    }
}