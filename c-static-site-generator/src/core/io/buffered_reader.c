#include "core/io/buffered_reader.h"
#include "core/result/result.h"
#include "core/math/math.h"

void buffered_reader_init(buffered_reader *br, reader source, str buf)
{
    br->source = source;
    br->buffer = buf;
    br->cursor = 0;
    br->read_end = 0;
}

size_t buffered_reader_read(buffered_reader *br, str buf, result *res)
{
    size_t n = 0;
    if (br->cursor > 0 && br->cursor < br->read_end)
    {
        str remaining;
        str_slice(br->buffer, &remaining, br->cursor, br->read_end);
        n = str_copy(buf, remaining);
        br->cursor += n;

        // if n >= buf.len, it means we had at least a whole `buf` left in the
        // buffer. If n == 0, it means we've reached the end of the file. In
        // either case, return.
        if (n >= buf.len || n < 1)
        {
            result_ok(res);
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

bool buffered_reader_find(
    buffered_reader *br,
    writer w,
    result *res,
    str match)
{
    char buf_[256] = {0};
    str buf;
    str_init(&buf, buf_, sizeof(buf_));

    size_t match_cursor = 0;
    while (true)
    {
        size_t nr = buffered_reader_read(br, buf, res);
        if (nr < 1)
        {
            return false;
        }

        str read;
        str_slice(buf, &read, 0, nr);

        // loop over each character in the valid portion of the buffer (the
        // slice that was populated by the last read).
        for (size_t i = 0; i < read.len; i++)
        {
            // check to see if there is a match beginning with the character in
            // the `i`th position of the buffer (taking into account that we
            // may be in the middle of a match that spans multiple buffers).
            for (size_t j = 0; j < match.len - match_cursor; j++)
            {
                // check to see if we've reached the end of the portion of the
                // buffer containg data from the last read. If so, update the
                // match cursor and jump to the outer loop so we can pull in
                // more data and continue matching.
                if (i + j >= read.len)
                {
                    match_cursor += j;
                    goto REFILL_BUFFER;
                }

                // we haven't reached the end of the valid portion of the
                // buffer yet, so check to see if we're still matching. If not,
                // then reset the `match_cursor` and start trying to match from
                // the next character.
                if (read.data[i + j] != match.data[match_cursor + j])
                {
                    match_cursor = 0;
                    goto RESET;
                }
            }

            // we've matched the remainder of the valid portion of the buffer
            // (from `i` to `i+match.len`); write the data up to the match,
            // rewind the cursor to the end of the match (`read.len - i +
            // (match.len - match_cursor`), and return.
            str prelude;
            str_slice(read, &prelude, 0, i);
            result write_res;
            result_init(&write_res);
            size_t nw = writer_write(w, prelude, &write_res);

            br->cursor -= read.len - (i + (match.len - match_cursor));

            if (!write_res.ok)
            {
                *res = write_res;
            }
            return true;

        RESET:
            continue;
        }

        result write_res;
    REFILL_BUFFER:
        result_init(&write_res);
        size_t nw = writer_write(w, read, &write_res);

        if (!write_res.ok)
        {
            *res = write_res;
            return false;
        }
        if (nw != nr)
        {
            result_err(res, ERR_SHORT_WRITE);
            return false;
        }
        if (!res->ok)
        {
            return false;
        }
    }
}

void buffered_reader_to_reader(buffered_reader *br, reader *r)
{
    r->data = (void *)br;
    r->read = (read_func)buffered_reader_read;
}