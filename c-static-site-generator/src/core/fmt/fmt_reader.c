#include "core/fmt/fmt_arg.h"
#include "core/fmt/fmt_reader.h"

fmt_reader fmt_reader_new(str format, fmt_args args)
{
    return (fmt_reader){
        .format = format,
        .args = args,
        .cursor = 0,
        .reading_arg = false,
    };
}

static size_t fmt_reader_read_arg(fmt_reader *fr, str buf, size_t buf_cursor)
{
    fmt_arg arg = fmt_args_first(&fr->args);
    if (fmt_arg_is_null(arg))
    {
        arg = FMT_STR(STR_LIT("{}(MISSING)"));
    }

    // read until we have a full cursor or until we've finished reading
    // the arg.
    while (buf_cursor < buf.len)
    {
        size_t nf = fmt_arg_format(arg, str_slice(buf, buf_cursor, buf.len));
        if (nf < 1)
        {
            goto FINISH_ARG;
        }
        buf_cursor += nf;
    }
    return buf_cursor;

FINISH_ARG:
    fmt_args_pop(&fr->args);
    fr->reading_arg = false;
    return buf_cursor;
}

static size_t fmt_reader_read(fmt_reader *fr, str buf, result *res)
{
    *res = result_ok(); // always ok
    size_t buf_cursor = 0;
    if (fr->reading_arg)
    {
        buf_cursor = fmt_reader_read_arg(fr, buf, buf_cursor);
    }
    while (fr->cursor < fr->format.len && buf_cursor < buf.len)
    {
        if (
            fr->cursor + 1 < fr->format.len &&
            fr->format.data[fr->cursor] == '{' &&
            fr->format.data[fr->cursor + 1] == '}')
        {
            fr->reading_arg = true;
            buf_cursor = fmt_reader_read_arg(fr, buf, buf_cursor);
            fr->cursor += 2;
            continue;
        }

        buf.data[buf_cursor] = fr->format.data[fr->cursor];
        buf_cursor++;
        fr->cursor++;
    }

    return buf_cursor;
}

reader fmt_reader_to_reader(fmt_reader *fr)
{
    return reader_new((void *)fr, (read_func)fmt_reader_read);
}