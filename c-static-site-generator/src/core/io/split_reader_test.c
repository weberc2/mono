#include "core/io/split_reader.h"
#include "core/io/str_reader.h"
#include "core/testing/test.h"
#include "std/string/string.h"
#include "std/string/string_writer.h"
#include <stdio.h>

#define TEST_READER(input)                           \
    (reader)                                         \
    {                                                \
        .data = (void *)&STR_READER(STR_LIT(input)), \
        .read = (read_func)str_reader_io_read,       \
    }

#define TEST_BUFFER(size, init) STR_ARR((char[size]){init})

typedef struct split_reader_section_write_to_test
{
    char *name;
    reader source;
    str delim;
    str buffer;
    split_reader_init_status wanted_init;
    str wanted_section;
    result wanted_res;
} split_reader_section_write_to_test;

split_reader_section_write_to_test split_reader_section_write_to_tests[] = {
    {
        .name = "empty-delim",
        .source = TEST_READER(""),
        .delim = STR_EMPTY,
        .buffer = TEST_BUFFER(3, 0),
        .wanted_init = split_reader_init_status_zero_length_delim,

        // these shouldn't be checked; just populating them so we have
        // determinism in case anything does use them.
        .wanted_section = STR_EMPTY,
        .wanted_res = RESULT_OK,
    },
    {
        .name = "buffer-shorter-than-delim",
        .source = TEST_READER(""),
        .delim = STR_LIT("foobar"),
        .buffer = TEST_BUFFER(3, 0),
        .wanted_init = split_reader_init_status_buffer_shorter_than_delim,

        // these shouldn't be checked; just populating them so we have
        // determinism in case anything does use them.
        .wanted_section = STR_EMPTY,
        .wanted_res = RESULT_OK,
    },
    {
        .name = "empty-source",
        .source = TEST_READER(""),
        .delim = STR_LIT("\n"),
        .buffer = TEST_BUFFER(3, 0),
        .wanted_init = split_reader_init_status_ok,
        .wanted_section = STR_EMPTY,
        .wanted_res = RESULT_OK,
    },
    {
        .name = "no-delims",
        .source = TEST_READER("foobar"),
        .delim = STR_LIT("\n"),
        .buffer = TEST_BUFFER(3, 0),
        .wanted_init = split_reader_init_status_ok,
        .wanted_section = STR_LIT("foobar"),
        .wanted_res = RESULT_OK,
    },
    {
        .name = "one-delim-entirely-in-single-buffer",
        .source = TEST_READER("foo\nbar"),
        .delim = STR_LIT("\n"),
        .buffer = TEST_BUFFER(3, 0),
        .wanted_init = split_reader_init_status_ok,
        .wanted_section = STR_LIT("foo"),
        .wanted_res = RESULT_OK,
    },
    {
        .name = "one-delim-split-across-buffers",
        .source = TEST_READER("ab---cd"),
        .delim = STR_LIT("---"),
        .buffer = TEST_BUFFER(6, 0),
        .wanted_init = split_reader_init_status_ok,
        .wanted_section = STR_LIT("ab"),
        .wanted_res = RESULT_OK,
    },
    {
        // confirm that delimiter matching works appropriately when
        // `delim.len == 2 * (buffer.len - delim.len)` and thus we have to
        // call `next_chunk()` multiple times to read a delimiter.
        .name = "delim-bigger-than-multiple-writable-buffers",
        .source = TEST_READER("abc----def"),
        .delim = STR_LIT("----"),
        .buffer = TEST_BUFFER(6, 0),
        .wanted_init = split_reader_init_status_ok,
        .wanted_section = STR_LIT("abc"),
        .wanted_res = RESULT_OK,
    },
    {
        // when the first chunk ends in some incomplete delimiter prefix and
        // the second chunk DOES NOT complete the match, then expect that the
        // false-delimiter is included in the resulting data.
        .name = "false-match-at-end-of-first-chunk",
        .source = TEST_READER("ab--cd"),
        .delim = STR_LIT("---"),
        .buffer = TEST_BUFFER(6, 0),
        .wanted_init = split_reader_init_status_ok,
        .wanted_section = STR_LIT("ab--cd"),
        .wanted_res = RESULT_OK,
    },
};

bool split_reader_section_write_to_test_run(
    split_reader_section_write_to_test *tc,
    string *section)
{
    char name[256] = {0};
    sprintf(name, "test_split_reader_section_write_to:%s", tc->name);
    test_init(name);
    split_reader sr;
    split_reader_init_status found_init = split_reader_init(
        &sr,
        tc->source,
        tc->delim,
        tc->buffer);
    if (found_init != tc->wanted_init)
    {
        return test_fail(
            "init: wanted `%s`; found `%s`",
            split_reader_init_status_to_str(tc->wanted_init).data,
            split_reader_init_status_to_str(found_init).data);
    }
    if (tc->wanted_init != split_reader_init_status_ok)
    {
        return test_success();
    }

    result found_res = result_new();
    split_reader_section_write_to(&sr, string_writer(section), &found_res);
    if (tc->wanted_res.ok != found_res.ok)
    {
        return test_fail(
            "result.ok: wanted `%s`; found `%s`",
            tc->wanted_res.ok ? "true" : "false",
            found_res.ok ? "true" : "false");
    }

    if (!str_eq(tc->wanted_section, string_borrow(section)))
    {
        // make sure it's null terminated for printf()
        string_push_char(section, '\0');
        return test_fail(
            "section: wanted `%s`; found `%s`",
            tc->wanted_section.data,
            section->data);
    }

    return test_success();
}

int main()
{
    int code = 0;
    string section = string_new();
    for (
        size_t i = 0;
        i < sizeof(split_reader_section_write_to_tests) /
                sizeof(split_reader_section_write_to_test);
        i++)
    {
        string_reset(&section);
        if (!split_reader_section_write_to_test_run(
                &split_reader_section_write_to_tests[i],
                &section))
        {
            code = 1;
            break;
        }
    }
    string_drop(&section);
    return code;
}