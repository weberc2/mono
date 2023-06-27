#include "field.h"
#include "core/panic/panic.h"

str field_match_status_str(field_match_status status)
{
    switch (status)
    {
    case field_match_in_progress:
        return STR_LIT("FIELD_MATCH_IN_PROGRESS");
    case field_match_success:
        return STR_LIT("FIELD_MATCH_SUCCESS");
    case field_match_failed:
        return STR_LIT("FIELD_MATCH_FAILED");
    default:
        panic("invalid `field_match_status`: %d", status);
    }
}

field field_new(str name, writer dst)
{
    return (field){
        .name = name,
        .dst = dst,
        .match_status = field_match_in_progress,
    };
}

bool fields_has_matches_in_progress(fields fields)
{
    for (size_t i = 0; i < fields.len; i++)
    {
        if (fields.data[i].match_status == field_match_in_progress)
        {
            return true;
        }
    }

    return false;
}