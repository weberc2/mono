#include "core/io/scanner.h"
#include <string.h>

scanner_new_result scanner_new(reader source, str buffer, str delim)
{
    if (buffer.len < delim.len)
    {
        scanner_new_result ret = {.ok = false};
        memset(&ret.scanner, 0, sizeof(ret.scanner));
        return ret;
    }

    return (scanner_new_result){
        .ok = true,
        .scanner = {
            .source = source,
            .buffer = buffer,
            .delim = delim,
            .buffer_cursor = 0,
            .delim_cursor = 0,
            .last_read_size = 0,
            .end_of_section = false,
            .err = ERROR_NULL,
        },
    };
}

bool error_is_eof(error err)
{
    return (err.data == ERR_EOF.data && err.display == ERR_EOF.display);
}

void scanner_refresh(scanner *s)
{
    // the write partition is where in the buffer we will read data into. it is
    // essentially the whole buffer minus some delimiter prefix partition at
    // the beginning where we will write any partial delimiter prefix we
    // encountered at the end of the previous frame in case this new frame
    // doesn't complete the delimiter match (in which case we need to return
    // the delimiter prefix data).
    str write_partition = str_slice(s->buffer, s->delim_cursor, s->buffer.len);
    io_result res = reader_read(s->source, write_partition);
    s->last_read_size = res.size;
    if (io_result_is_err(res))
    {
        s->err = res.err;
    }
    else if (res.size < 1)
    {
        s->err = ERR_EOF;
    }
}

size_t ends_with_prefix(str s, str match)
{
    size_t largest_prefix = match.len - 1;

    for (size_t start = s.len - largest_prefix; start < s.len; start++)
    {
        for (size_t j = 0; j + start < s.len; j++)
        {
            if (s.data[start + j] != match.data[j])
            {
                goto RESET;
            }
        }

        // if we get here, then we've matched an entire prefix; return the
        // prefix cursor.
        return s.len - start;

    RESET:
        continue;
    }

    // if we get here, then we've failed to match a prefix; return 0 to
    // indicate that no prefix (or rather, the zero-length prefix) was matched.
    return 0;
}

scan_result scanner_next_frame(scanner *s)
{
    if (error_is_null(s->err) && s->buffer_cursor >= s->last_read_size)
    {
        scanner_refresh(s);
    }

    str write_partition = str_slice( // TODO: we already get the write_partition in `refresh()`; can we factor this out instead of doing it twice?
        s->buffer,
        s->delim_cursor,
        s->delim_cursor + s->last_read_size);
    str delim_remaining = str_slice(s->delim, s->delim_cursor, s->delim.len);
    if (str_has_prefix(write_partition, delim_remaining))
    {
        // if we got here, then we've completed a delimiter match. set the
        // cursor to the end of the delimiter and set the error to ERR_EOF if it's
        // not already set (also set the `end_of_section` flag so we can
        // distinguish between end-of-section and end-of-source-stream
        // conditions). return an empty string since the delimiter began in the
        // previous frame (no part of this frame's data is part of the
        // section).
        s->buffer_cursor = s->delim_cursor;
        if (error_is_null(s->err))
        {
            s->err = ERR_EOF;
            s->end_of_section = true;
        }
        return (scan_result){.data = STR_EMPTY, .err = s->err};
    }

    // if we got here, then we didn't complete a match and we have to return
    // all of the data we read into the buffer as well as any prefix data that
    // we started to match at the end of the previous frame.
    str delim_prefix_partition = str_slice(s->buffer, 0, s->delim_cursor);
    str_copy(delim_prefix_partition, s->delim);

    // check to see if the write partition contains a delimiter... if so,
    // return up to that delimiter (updating the bookkeeping information as
    // well)
    str_find_result res = str_find(write_partition, s->delim);
    if (res.found)
    {
        s->buffer_cursor = (s->delim.len - s->delim_cursor) + res.index;
        if (error_is_null(s->err))
        {
            s->err = ERR_EOF;
            s->end_of_section = true;
        }
        return (scan_result){
            .data = str_slice(s->buffer, 0, s->delim_cursor + res.index),
            .err = s->err,
        };
    }

    // otherwise return the whole frame less any potential delimiter prefix at
    // the end of the buffer (advancing the cursor accordingly)
    s->buffer_cursor = s->delim_cursor + s->last_read_size;
    size_t last_read_end = s->delim_cursor + s->last_read_size;
    s->delim_cursor = ends_with_prefix(write_partition, s->delim);

    return (scan_result){
        // return the buffer beginning at 0 to the end of the write partition
        // less any delimiter prefix at the end of the write partition (in case
        // the next iteration determines that there was a prefix straddling the
        // boundary between buffers).
        .data = str_slice(
            s->buffer,
            0,
            // note that by this point `s->delim_cursor` refers to any
            // potential delimiter prefixes at the end of the most recent
            // write partition (rather than that of the previous iteration).
            // this bit essentially says "return up to the end of the write
            // partition LESS any delimiter prefixes at the end of the write
            // partition".
            last_read_end - s->delim_cursor),
        .err = s->err,
    };
}

bool scanner_begin_next_section(scanner *s)
{
    while (true)
    {
        scan_result res = scanner_next_frame(s);
        if (!error_is_null(res.err))
        {
            if (error_is_eof(res.err))
            {
                // if we're just at the end of a section (but not the end of
                // the source stream) then set the `end_of_section` flag to
                // `false`, null out the `err` field, and return `true`.
                if (s->end_of_section)
                {
                    s->end_of_section = false;
                    s->err = ERROR_NULL;
                    return true;
                }

                // we're already at the end of the source stream; return
                // `false`.
                return false;
            }

            // some other error; return `false`
            return false;
        }

        // no error, end-of-section, or end-of-source-stream conditions
        // encountered. keep looping...
    }
}

size_t scanner_write_to(scanner *s, writer dst, result *res)
{
    size_t total_written = 0;
    while (true)
    {
        scan_result scan_res = scanner_next_frame(s);
        io_result write_res = writer_write(dst, scan_res.data);
        res->err = write_res.err;
        res->ok = io_result_is_ok(write_res);
        total_written += write_res.size;

        // if there was a write error, return
        if (!res->ok)
        {
            break;
        }

        // if the write was too short, return
        if (write_res.size < scan_res.data.len)
        {
            *res = result_err(ERR_SHORT_WRITE);
            break;
        }

        // if there was a scan error (including eof), return
        if (!error_is_null(scan_res.err))
        {
            res->ok = false;
            res->err = scan_res.err;
            break;
        }

        // otherwise loop around
    }
    return total_written;
}

static void __attribute__((constructor)) init()
{
    error_const(&ERR_EOF, "end of file");
}