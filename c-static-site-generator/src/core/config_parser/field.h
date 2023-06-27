#ifndef FIELD_H
#define FIELD_H

#include <stddef.h>
#include "core/str/str.h"
#include "core/io/writer.h"

typedef enum field_match_status
{
    field_match_in_progress,
    field_match_success,
    field_match_failed,
} field_match_status;

str field_match_status_str(field_match_status str);

typedef struct field
{
    str name;
    writer dst;
    field_match_status match_status;
} field;

field field_new(str name, writer dst);

#define FIELD(n, d, ms)       \
    (field)                   \
    {                         \
        .name = (n),          \
        .dst = (d),           \
        .match_status = (ms), \
    }

typedef struct fields
{
    field *data;
    size_t len;
} fields;

#define FIELDS(...)                                            \
    (fields)                                                   \
    {                                                          \
        .data = (field[]){__VA_ARGS__},                        \
        .len = sizeof((field[]){__VA_ARGS__}) / sizeof(field), \
    }

bool fields_has_matches_in_progress(fields fields);

typedef size_t field_handle;

#endif // FIELD_H