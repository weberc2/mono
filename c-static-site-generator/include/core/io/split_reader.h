#ifndef SPLIT_READER_H
#define SPLIT_READER_H

#include "reader.h"
#include "writer.h"

typedef enum split_reader_state
{
    split_reader_state_ready,
    split_reader_state_end_of_section,
    split_reader_state_end_of_source,
    split_reader_state_error,
} split_reader_state;

typedef enum split_reader_init_status
{
    split_reader_init_status_ok,
    split_reader_init_status_zero_length_delim,
    split_reader_init_status_buffer_shorter_than_delim,
} split_reader_init_status;

static inline str split_reader_init_status_to_str(split_reader_init_status s)
{
    switch (s)
    {
    case split_reader_init_status_ok:
        return STR_LIT("SPLIT_READER_INIT_STATUS_OK");
    case split_reader_init_status_zero_length_delim:
        return STR_LIT("SPLIT_READER_INIT_STATUS_ZERO_LENGTH_DELIM");
    case split_reader_init_status_buffer_shorter_than_delim:
        return STR_LIT("SPLIT_READER_INIT_STATUS_BUFFER_SHORTER_THAN_DELIM");
    }
}

typedef struct split_reader
{
    reader source;
    str delim;
    str buffer;
    size_t cursor;
    size_t delim_cursor;
    size_t last_read_size;
    split_reader_state state;
    error err;
} split_reader;

split_reader_init_status split_reader_init(
    split_reader *sr,
    reader source,
    str delim,
    str buffer);
bool split_reader_next_chunk(split_reader *sr, str *chunk);
bool split_reader_next_section(split_reader *sr);
size_t split_reader_section_write_to(split_reader *sr, writer w, result *res);

#endif // SPLIT_READER_H