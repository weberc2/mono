#include "core/testing/test.h"
#include "config_parser.h"

typedef struct parse_field_test
{
    config_parser parser;
    parse_field_result wanted_result;
    str wanted_buf;
    size_t wanted_cursor;
    size_t wanted_last_read_end;
} parse_field_test;