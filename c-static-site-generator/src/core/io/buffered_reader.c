#include "core/io/buffered_reader.h"
#include "core/result/result.h"
#include "core/math/math.h"

void buffered_reader_init(buffered_reader *br, reader source, str buf)
{
    br->source = source;
    br->buffer = buf;
    br->cursor = 0;
}

size_t buffered_reader_read(buffered_reader *br, str buf, result *res)
{
    size_t n = 0;
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
        // reset the cursor back to the beginning of the buffer.
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
    buffered_reader *src,
    writer dst,
    result *res,
    str match)
{
    char buf_[256];
    str buf;
    str_init(&buf, buf_, sizeof(buf_));

    size_t match_cursor = 0;
    while (true)
    {
        size_t nr;
    READ_MORE:

        nr = buffered_reader_read(src, buf, res);

        str read_slice;
        str_slice(buf, &read_slice, 0, nr);

        // if we read 0 bytes, it means we reached `eof` without finding a
        // match.
        if (nr == 0)
        {
            return false;
        }

        // loop over the buffer to try and find a match--it's possible that we
        // began matching but couldn't complete the match because the buffer
        // ran out before the match was completed--`match_cursor` holds the
        // index into `match` where we left off during the last iteration.
        for (size_t i = 0; i < read_slice.len; i++)
        {
            for (size_t j = 0; j + match_cursor < match.len; j++)
            {
                // check to see if the current character is out-of-bounds of
                // the read_slice--if so, then update the match_cursor and jump
                // to the outermost loop.
                if (i + j >= read_slice.len)
                {
                    match_cursor += read_slice.len;
                    goto READ_MORE;
                }

                // if we got here, we're still in bounds of the previously-read
                // data--check to see if we are still matching; if not, reset
                // the match_cursor and jump to the next iteration of the match
                // loop.
                if (read_slice.data[i + j] != match.data[j + match_cursor])
                {
                    match_cursor = 0;
                    goto RESET;
                }
            }

            // If we got here, we completed a match, so we'll finish copying,
            // reverse the `buffered_reader` cursor to the end of the match,
            // and return `true`.
            size_t match_end = i + (match.len - match_cursor);
            size_t rewind = read_slice.len - match_end;
            src->cursor -= rewind;

            // slice the buffer up to the start of the match
            str slice;
            str_slice(read_slice, &slice, 0, i);

            // write the new slice to the writer
            result write_res;
            result_init(&write_res);
            size_t nw = writer_write(dst, slice, &write_res);

            // handle write errors
            if (!write_res.ok)
            {
                *res = write_res;
                return true;
            }

            // at this point, `res` may or may not be `ok` (it may be
            // in an `err` state from the previous read operation), but
            // in either case we will return with the number of bytes
            // copied.
            return true;

        RESET:
            continue;
        }

        // we've scanned the slice of the buffer containing data from the
        // last read and we never found a match--copy that slice of the buffer
        // to the writer and loop back around, refreshing the buffer.
        result write_res;
        result_init(&write_res);
        size_t nw = writer_write(dst, read_slice, &write_res);

        // if there was a problem writing, return the error.
        if (!write_res.ok)
        {
            *res = write_res;
            return false;
        }

        // if we didn't write as many bytes as we read, return an error.
        if (nw != nr)
        {
            result_err(res, ERR_SHORT_WRITE);
            return false;
        }
    }
}

void buffered_reader_to_reader(buffered_reader *br, reader *r)
{
    reader_init(r, (void *)br, (read_func)buffered_reader_read);
}