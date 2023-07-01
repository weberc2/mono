#ifndef FIELD_H
#define FIELD_H

#include "core/str/str.h"
#include "core/io/writer.h"

typedef enum field_status
{
    field_status_inconclusive,
    field_status_disqualified,
    field_status_matched,
} field_status;

typedef struct field
{
    str key;
    writer value;
    field_status status;
} field;

#define FIELD(s, v)                          \
    (field)                                  \
    {                                        \
        .key = STR_LIT(s),                   \
        .value = v,                          \
        .status = field_status_inconclusive, \
    }

typedef struct fields
{
    field *data;
    size_t len;
    size_t cursor;
} fields;

#define FIELDS(...)                                            \
    (fields)                                                   \
    {                                                          \
        .data = (field[]){__VA_ARGS__},                        \
        .len = sizeof((field[]){__VA_ARGS__}) / sizeof(field), \
        .cursor = 0,                                           \
    }

#endif // FIELD_H