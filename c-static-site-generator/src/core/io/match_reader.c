#include "core/io/match_reader.h"

match_reader match_reader_new(buffered_reader *source, str match)
{
    return (match_reader){
        .source = source,
        .match = match,
        .match_cursor = 0,
        .found_match = 0,
    };
}

size_t match_reader_read(match_reader *mr, str buf, result *res)
{
    // if the last invocation was a match, return eof so the callers know there
    // was a match. further, reset the `found_match` flag so the next call will
    // begin searching for a new instance of `match`.
    if (mr->found_match)
    {
        *res = result_ok();
        mr->found_match = false;
        return 0;
    }

    size_t nr = buffered_reader_read(mr->source, buf, res);
    mr->source->cursor -= nr;

    if (nr < 1)
    {
        return 0;
    }

    str read_slice = str_slice(buf, 0, nr);
    for (size_t start = 0; start < read_slice.len; start++)
    {
        for (size_t end = 0; end < mr->match.len - mr->match_cursor; end++)
        {
            if (start + end > read_slice.len)
            {
                mr->match_cursor += end;
                return read_slice.len;
            }

            // check to see if there is a match--if so, continue; otherwise
            // reset the match cursor and jump back to the beginning of the
            // outer loop to start the match over at the next source character.
            char buf_char = read_slice.data[start + end];
            char match_char = mr->match.data[mr->match_cursor + end];
            if (buf_char != match_char)
            {
                mr->match_cursor = 0;
                goto OUTER;
            }
        }

        // rewind the buffered_reader's cursor so the next read resumes from
        // the end of the match.
        mr->source->cursor = start + (mr->match.len - mr->match_cursor);

        // set the `found_match` flag so the next call will properly indicate
        // eof.
        mr->found_match = true;

        // return the starting position of the match. note that the call to
        // `buffered_reader_read()` already updated `res` whether success or
        // error--we're just going to return it along with any matches.
        nr = start;
        break;

    OUTER:
        continue;
    }

    return nr;
}

reader match_reader_to_reader(match_reader *mr)
{
    return reader_new((void *)mr, (read_func)match_reader_read);
}