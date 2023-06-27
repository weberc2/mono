#include "fields_match_name.h"

fields_match_result fields_match_name(
    fields fields,
    size_t field_name_cursor,
    str buf)
{
    const char DELIM = ':';
    if (buf.len < 1)
    {
        return FIELDS_MATCH_FAILURE;
    }

    for (field_handle i = 0; i < fields.len; i++)
    {
        // if the field is already disqualified (because it has already been
        // marked a failure or a success), then continue to the next field
        if (fields.data[i].match_status != field_match_in_progress)
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
                fields.data[i].match_status = field_match_success;
                return FIELDS_MATCH_OK(i, 0);
            }
            else
            {
                fields.data[i].match_status = field_match_failed;
                continue;
            }
        }

        // if the field name is longer than the buffer, check to see if the
        // buffer is a prefix of the field name--in which case we continue the
        // match; otherwise we mark it as a failed match.
        if (field_name.len > buf.len)
        {
            if (!str_has_prefix(field_name, buf))
            {
                fields.data[i].match_status = field_match_failed;
            }
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
                fields.data->match_status = field_match_success;
                return FIELDS_MATCH_OK(i, field_name.len);
            }
            fields.data[i].match_status = field_match_failed;
            continue;
        }

        // lastly, if the buffer and the unmatched portion of the field name
        // are exactly the same length (which is necessarily true if we get to
        // this point), then check them for equality. If they are not equal,
        // then disqualify the field; however, if they are equal, then we *may*
        // still have a match (depending on whether or not the next buffer
        // begins with the key/value delimeter).
        if (!str_eq(buf, field_name))
        {
            fields.data[i].match_status = field_match_failed;
        }
    }

    return FIELDS_MATCH_FAILURE;
}