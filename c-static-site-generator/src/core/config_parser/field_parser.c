#include "core/panic/panic.h"
#include "core/io/writer.h"
#include "field_parser.h"

field field_new(str name, writer dst)
{
    return (field){
        .name = name,
        .dst = dst,
        .match_failed = false,
    };
}

bool fields_has_valid(fields fields)
{
    for (size_t i = 0; i < fields.len; i++)
    {
        if (!fields.data[i].match_failed)
        {
            return true;
        }
    }

    return false;
}

field_match_result fields_match_name(
    fields fields,
    size_t field_name_cursor,
    str buf)
{
    const char DELIM = ':';
    // throw this check in here because there's at least one instance where we
    // look at the first character in the buffer later in the function.
    if (buf.len < 1)
    {
        panic("fields_match_name(): zero-length buffer!");
    }

    for (size_t i = 0; i < fields.len; i++)
    {
        // if the field is already disqualified, then continue to the next
        // field
        if (fields.data[i].match_failed)
        {
            continue;
        }

        // get the unmatched portion of the field name
        str field_name = str_slice(
            fields.data[i].name,
            field_name_cursor,
            fields.data[i].name.len);

        // if the unmached portion of the field name is zero length (i.e.,
        // we've matched the field name exactly at the end of the previous
        // buffer and now we're checking to see if the field name is followed
        // immediately by the key/value delimiter) and the first char in the
        // buffer is a key/value delimiter, then we've found a complete match.
        // If the first char in the buffer is NOT a key/value delimiter, then
        // we've failed to match for the field.
        if (field_name.len < 1)
        {
            if (buf.data[0] == DELIM)
            {
                return FIELD_MATCH_RESULT_SUCCESS(i, 0);
            }
            else
            {
                fields.data[i].match_failed = true;
                continue;
            }
        }

        // if the field name is longer than the buffer, check to see if the
        // buffer is a prefix of the field name--in which case we continue the
        // match; otherwise we mark it as a failed match.
        if (field_name.len > buf.len)
        {
            fields.data[i].match_failed &= str_has_prefix(field_name, buf);
            continue;
        }

        // otherwise, if the buffer is longer than the field name, check to see
        // if the buffer is prefixed with the field name AND the next character
        // is a key/value delimiter--if so, we have a match; if not, then the
        // field is disqualified.
        if (buf.len > field_name.len)
        {
            if (
                str_has_prefix(buf, field_name) &&
                buf.data[field_name.len] == DELIM)
            {
                return FIELD_MATCH_RESULT_SUCCESS(i, field_name.len);
            }
            fields.data[i].match_failed = true;
            continue;
        }

        // lastly, if the buffer and the unmatched portion of the field name
        // are exactly the same length (which is necessarily true if we get to
        // this point), then check them for equality. If they are not equal,
        // then disqualify the field; however, if they are equal, then we *may*
        // still have a match (depending on whether or not the next buffer
        // begins with the key/value delimeter).
        fields.data[i].match_failed = !str_eq(buf, field_name);
    }

    return FIELD_MATCH_RESULT_FAILURE;
}

field_match_result parse_field_name(
    reader r,
    fields fields,
    str buf)
{
    size_t field_name_cursor = 0;
    while (fields_has_valid(fields))
    {
        field_match_result match_res = FIELD_MATCH_RESULT_FAILURE;
        io_result res = reader_read(r, buf);
        if (io_result_is_err(res))
        {
            match_res.io_err = res.err;
            match_res.buffer_position = res.size;
        }
        if (res.size < 1)
        {
            return match_res;
        }

        str gooddata = str_slice(buf, 0, res.size);

        match_res = fields_match_name(
            fields,
            field_name_cursor,
            gooddata);
        // if we found a match *or* if the reader encountered an io error
        // return the match result (in the latter case, the match result
        // contains the error information already, so we'll still be
        // communicating the error).
        if (match_res.match || io_result_is_err(res))
        {
            return match_res;
        }

        // otherwise we're still matching, so add the size of the gooddata to
        // the field_name_cursor and continue with another buffer's worth of
        // data.
        field_name_cursor += gooddata.len;
    }

    return FIELD_MATCH_RESULT_FAILURE;
}

parse_field_value_result parse_field_value(reader r, writer w, str buf)
{
    size_t size = 0;
    str_find_result find_res = (str_find_result){.found = false, .index = 0};
    while (!find_res.found)
    {
        io_result res = reader_read(r, buf);
        if (res.size < 1)
        {
            return (parse_field_value_result){
                .ok = true,
                .total_size = size,
                .buffer_position = 0,
                .err = error_null(),
            };
        }

        str gooddata = str_slice(buf, 0, res.size);
        find_res = str_find_char(gooddata, '\n');
        str value = find_res.found
                        ? str_slice(gooddata, 0, find_res.index)
                        : gooddata;
        res = writer_write(w, value);
        size += res.size;

        if (res.size != value.len)
        {
            return (parse_field_value_result){
                .ok = false,
                .total_size = size,
                .buffer_position = res.size,
                .err = ERR_SHORT_WRITE,
            };
        }

        if (io_result_is_err(res))
        {
            return (parse_field_value_result){
                .ok = false,
                .total_size = size,
                .buffer_position = res.size,
                .err = res.err,
            };
        }
    }

    return (parse_field_value_result){
        .ok = true,
        .total_size = size,
        .buffer_position = find_res.index,
        .err = error_null(),
    };
}
