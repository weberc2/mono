#include <string.h>
#include <stdint.h>

#include "core/io/split_reader.h"
#include "core/io/writer.h"

split_reader_init_status split_reader_init(
    split_reader *sr,
    reader source,
    str delim,
    str buffer)
{
    if (delim.len < 1)
    {
        // zero-out the split_reader to prevent heisenbugs if someone tries to
        // use the split_reader when `.ok = false`.
        memset((void *)sr, 0, sizeof(*sr));
        return split_reader_init_status_zero_length_delim;
    }

    // the buffer needs to be a contiguous slice of bytes partitioned in two:
    // the first partition must be `delim.len - 1` bytes long to hold any
    // incomplete delimiter prefixes (hence `delim.len - 1`) and the second
    // partition will be passed to `reader_read(source, <partition>, ...)` for
    // reading in data. in cases where a read fills the second partition with
    // data ending in a delimiter prefix (i.e., incomplete), we will note that
    // there may be a match in progress by advancing the `delim_cursor` and we
    // will return everything up to the start of that prefix. if the subsequent
    // read completes the match, then we'll set the state to `end_of_section`
    // and return an empty slice, but if the match is incomplete then the
    // prefix is actually not part of a delimiter at all and thus it must be
    // returned with the data from the second partition--in this case, the
    // first partition allows us to write that prefix into the front of the
    // buffer so we can return the prefix *and* all of the data in the second
    // partition.
    if (buffer.len < delim.len)
    {
        // zero-out the split_reader to prevent heisenbugs if someone tries to
        // use the split_reader when `.ok = false`.
        memset((void *)sr, 0, sizeof(*sr));
        return split_reader_init_status_buffer_shorter_than_delim;
    }

    str_copy(buffer, delim);
    *sr = (split_reader){
        .source = source,
        .delim = delim,
        .buffer = buffer,
        .cursor = 0,
        .delim_cursor = 0,
        .last_read_size = 0,
        .state = split_reader_state_ready,
    };
    return split_reader_init_status_ok;
}

static inline str split_reader_write_partition(split_reader *sr)
{
    return (str){
        .data = sr->buffer.data + sr->delim.len - 1,
        .len = sr->buffer.len - sr->delim.len + 1,
    };
}

bool split_reader_refresh(split_reader *sr)
{
    result res = result_new();
    sr->last_read_size = reader_read(
        sr->source,
        split_reader_write_partition(sr),
        &res);
    sr->cursor = 0;
    if (!res.ok)
    {
        sr->state = split_reader_state_error;
        sr->err = res.err;
        return false;
    }
    if (sr->last_read_size < 1)
    {
        sr->state = split_reader_state_end_of_source;
        return false;
    }

    return true;
}

bool split_reader_next_chunk(split_reader *sr, str *chunk)
{
    switch (sr->state)
    {
    case split_reader_state_error:
    case split_reader_state_end_of_section:
    case split_reader_state_end_of_source:
        *chunk = STR_EMPTY;
        return false;
    case split_reader_state_ready:
        break;
    }

    if (!split_reader_refresh(sr))
    {
        *chunk = STR_EMPTY;
        return false;
    }

    size_t write_partition_offset = sr->delim.len - 1;
    size_t good_data_end = write_partition_offset + sr->last_read_size;
    str data = str_slice(sr->buffer, write_partition_offset, good_data_end);

    for (size_t start = 0; start < data.len; start++)
    {
        size_t delim_remaining = sr->delim.len - sr->delim_cursor;
        for (size_t cursor = 0; cursor < delim_remaining; cursor++)
        {
            // check to see if we've matched all the bytes in the populated
            // slice of the buffer.
            if (start + cursor >= data.len)
            {
                // if we got here, we either have an incomplete delimiter match
                // (in cases where `cursor >= 0` and thus `start < data.len`)
                // or definitely no match at all (in cases where `cursor == 0`
                // and thus `start == data.len`). In either case, we should
                // return up to `start` (which is either the start of the
                // delimiter or the end of the buffer) and we should increment
                // the `delim_cursor` by `cursor` (which may be `0`).
                *chunk = str_slice(data, 0, start);
                sr->delim_cursor += cursor;
                return true;
            }

            // check to see if the current position fails the match
            if (data.data[start + cursor] !=
                sr->delim.data[sr->delim_cursor + cursor])
            {
                // we've failed to match

                // if the previous write partition ended with a delimiter
                // prefix and we've just failed the match, then we need to
                // write the delimiter prefix into the delimiter prefix scratch
                // partition preceding the write partition so we can return the
                // prefix as part of the data. in order to communicate that the
                // returned slice needs to begin somewhere in the preceding
                // slice, we will shift the `write_partition_offset` value to
                // the left by `delim_cursor` bytes.
                str delim_prefix = str_slice(sr->delim, 0, sr->delim_cursor);
                str delim_prefix_scratch_partition = str_slice(
                    sr->buffer,
                    (sr->delim.len - 1 - sr->delim_cursor),
                    sr->delim.len - 1);
                str_copy(
                    delim_prefix_scratch_partition,
                    delim_prefix);
                write_partition_offset -= sr->delim_cursor;

                // set `delim_cursor` to `0` and loop around to start the
                // matching process again from the beginning of the delimiter.
                sr->delim_cursor = 0;
                goto NEXT_START_CHAR;
            }
        }

        // if we've got here, then we've matched a delimiter. return the slice
        // preceding the delimiter (which may be `.len == 0` if the delimiter
        // began at the front of the current buffer or--incompletely--at the
        // end of the previous buffer) and set the state to end-of-section.
        *chunk = str_slice(data, 0, start);
        sr->state = split_reader_state_end_of_section;
        return true;

    NEXT_START_CHAR:
        continue;
    }

    // if we got here, then we definitely found no portion of a delimiter.
    *chunk = str_slice(sr->buffer, write_partition_offset, good_data_end);
    return true;
}

bool split_reader_next_section(split_reader *sr)
{
    return false;
}

size_t write_helper(split_reader *sr, writer w, str chunk, result *res)
{
    size_t nr = writer_write(w, chunk, res);
    if (!res->ok)
    {
        return nr;
    }
    if (nr != chunk.len)
    {
        *res = result_err(ERR_SHORT_WRITE);
    }
    return nr;
}

size_t split_reader_section_write_to(split_reader *sr, writer w, result *res)
{
    str chunk = STR_EMPTY;
    size_t total_written = 0;
    while (split_reader_next_chunk(sr, &chunk))
    {
        total_written += write_helper(sr, w, chunk, res);
        if (!res->ok)
        {
            return total_written;
        }
    }

    // in case of error, write anything remaining in `chunk` before returning
    total_written += write_helper(sr, w, chunk, res);
    return total_written;
}