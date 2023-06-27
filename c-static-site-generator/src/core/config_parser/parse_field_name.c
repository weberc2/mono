#include "parse_field_name.h"
#include "fields_match_name.h"

parse_field_name_result parse_field_name(
    reader r,
    fields fields,
    str buf,
    size_t cursor,
    size_t last_read_end)
{
    size_t field_name_cursor = 0;
    while (fields_has_matches_in_progress(fields))
    {
        str gooddata = str_slice(buf, cursor, last_read_end);

        fields_match_result match_res = fields_match_name(
            fields,
            field_name_cursor,
            gooddata);
        if (match_res.match)
        {
            return PARSE_FIELD_NAME_OK(
                match_res.field_handle,
                match_res.buffer_position);
        }

        // otherwise we're still matching, so add the size of the gooddata to
        // the field_name_cursor and continue with another buffer's worth of
        // data.
        field_name_cursor += gooddata.len;
        result res = result_new();
        last_read_end = reader_read(r, buf, &res);
        cursor = 0;
        if (!res.ok)
        {
            return PARSE_FIELD_NAME_IO_ERROR(res.err);
        }

        if (last_read_end < 1)
        {
            // if we got here, then we still have valid fields, but we've hit
            // the end of the file, which means we haven't matched anything.
            // Return `parse_match_failure`.
            return PARSE_FIELD_NAME_MATCH_FAILURE;
        }
    }

    return PARSE_FIELD_NAME_MATCH_FAILURE;
}