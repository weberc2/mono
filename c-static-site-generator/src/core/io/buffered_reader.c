#include "core/math/math.h"
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

io_result buffered_reader_read(buffered_reader *br, str buf)
{
    io_result ret = IO_RESULT_OK(0);
    if (br->cursor > 0 && br->cursor < br->read_end)
    {
        str remaining = str_slice(br->buffer, br->cursor, br->read_end);
        ret.size = str_copy(buf, remaining);
        br->cursor += ret.size;

        // if `ret.size >= buf.len`, it means we had at least a whole `buf`
        // left in the buffer. If `ret.size == 0`, it means we've reached the
        // end of the file. In either case, return.
        if (ret.size >= buf.len || ret.size < 1)
        {
            return ret;
        }
    }
    else
    {
        ret.size = 0;
    }

    // otherwise, we only partially filled the output buffer and we need to
    // reload the internal buffer.
    while (ret.size < buf.len)
    {
        io_result res = reader_read(br->source, br->buffer);
        ret.err = res.err;

        // NB: we are deliberately *NOT* handling errors at this point--we
        // first want to copy anything we successfully read into the output
        // buffer.

        // if we don't read anything, it means we've reached eof--return with
        // a partially-written output buffer.
        if (res.size < 1)
        {
            break;
        }

        // if we didn't read anything because we reached eof, we don't want to
        // reset the `read_end` back to the beginning of the buffer.
        br->read_end = res.size;

        // if we didn't read anything because we reached eof, we don't want to
        // reset the `cursor` back to the beginning of the buffer.
        br->cursor = 0;

        // otherwise we read something; let's copy it to the unwritten portion
        // of the output buffer.
        str target = str_slice(buf, ret.size, min(buf.len, res.size));
        size_t copied = str_copy(target, br->buffer);
        size_t unwritten = buf.len - ret.size;
        ret.size += copied;
        br->cursor += copied;

        // if we filled the output buffer OR encountered errors, return.
        if (copied >= unwritten || io_result_is_err(res))
        {
            break;
        }

        // otherwise loop around and refill the buffer
    }
    return ret;
}

bool buffered_reader_find(
    buffered_reader *br,
    writer w,
    str match,
    io_result *res)
{
    char buf_[256] = {0};
    str buf = str_new(buf_, sizeof(buf_));

    match_reader mr = match_reader_new(br, match);
    while (true)
    {
        io_result read_res = match_reader_read(&mr, buf);
        if (read_res.size < 1)
        {
            return true;
        }

        io_result write_res = writer_write(
            w,
            str_slice(buf, 0, read_res.size));

        if (io_result_is_err(write_res))
        {
            res->err = write_res.err;
            return false;
        }

        if (io_result_is_err(read_res))
        {
            res->err = read_res.err;
            return false;
        }

        if (read_res.size != write_res.size)
        {
            *res = IO_RESULT_ERR(ERR_SHORT_WRITE);
            return false;
        }
    }
}

void buffered_reader_to_reader(buffered_reader *br, reader *r)
{
    r->data = (void *)br;
    r->read = (read_func)buffered_reader_read;
}