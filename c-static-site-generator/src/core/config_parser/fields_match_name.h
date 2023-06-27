#ifndef FIELDS_MATCH_NAME_H
#define FIELDS_MATCH_NAME_H

#include <stddef.h>
#include <stdbool.h>
#include "core/str/str.h"
#include "field.h"

typedef struct fields_match_result
{
    bool match;
    size_t buffer_position;
    size_t field_handle;
} fields_match_result;

#define FIELDS_MATCH_OK(fh, bp)  \
    (fields_match_result)        \
    {                            \
        .match = true,           \
        .buffer_position = (bp), \
        .field_handle = (fh),    \
    }

#define FIELDS_MATCH_FAILURE  \
    (fields_match_result)     \
    {                         \
        .match = false,       \
        .buffer_position = 0, \
        .field_handle = 0,    \
    }

fields_match_result fields_match_name(
    fields fields,
    size_t field_name_cursor,
    str buf);

#endif // FIELDS_MATCH_NAME_H