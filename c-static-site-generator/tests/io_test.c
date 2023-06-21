#include <stdbool.h>
#include <string.h>
#include <stdlib.h>

#include "test.h"
#include "core/io/copy.h"
#include "core/result/result.h"
#include "core/io/str_reader.h"
#include "core/io/buffered_reader.h"
#include "std/string/string.h"
#include "std/string/string_formatter.h"
#include "std/string/string_writer.h"

bool test_str_reader()
{
    test_init("test_str_reader");

    char data[] = "helloworld";
    str source;
    str_init(&source, data, sizeof(data) - 1);

    char buf[6] = ".....";
    str buffer;
    str_init(&buffer, buf, sizeof(buf) - 1);

    str_reader br;
    str_reader_init(&br, source);

    reader r;
    str_reader_to_reader(&br, &r);

    result res;
    result_ok(&res);

    size_t nr = reader_read(r, buffer, &res);

    if (nr != buffer.len)
    {
        return test_fail("nr: wanted `%zu`; found `%zu`", buffer.len, nr);
    }
    ASSERT_OK(res);

    str wanted;
    str_init(&wanted, "hello", 5);
    if (!str_eq(wanted, buffer))
    {
        return test_fail(
            "data: wanted `%s`; found `%s`",
            wanted.data,
            buffer.data);
    }
    ASSERT_OK(res);

    // Read a second time to get the rest of the data
    nr = reader_read(r, buffer, &res);

    if (nr != buffer.len)
    {
        return test_fail("nr: wanted `%zu`; found `%zu`", buffer.len, nr);
    }

    str_init(&wanted, "world", 5);
    if (!str_eq(wanted, buffer))
    {
        return test_fail(
            "data: wanted `%s`; found `%s`",
            wanted.data,
            buffer.data);
    }
    ASSERT_OK(res);

    // read a third time to get the eof
    nr = reader_read(r, buffer, &res);
    if (nr != 0)
    {
        return test_fail("nr: wanted `0`; found `%zu`", nr);
    }
    ASSERT_OK(res);

    return test_success();
}

bool test_copy()
{
    test_init("test_copy");

    char srcdata[] = "helloworld";
    str src;
    str_init(&src, srcdata, sizeof(srcdata) - 1);

    string dst;
    string_init(&dst);
    TEST_DEFER(string_drop, &dst);

    result res;
    result_ok(&res);

    str_reader str_reader;
    str_reader_init(&str_reader, src);

    reader r;
    str_reader_to_reader(&str_reader, &r);

    writer w;
    string_writer(&w, &dst);

    copy(w, r, &res);

    str dstslice;
    string_borrow(&dst, &dstslice);
    if (!str_eq(src, dstslice))
    {
        return test_fail("wanted `%s`; found `%s`", src.data, dst.data);
    }

    return test_success();
}

bool assert_str_eq(char *wanted_c, str found)
{
    str wanted;
    str_init(&wanted, wanted_c, strlen(wanted_c));
    if (!str_eq(wanted, found))
    {
        char *found_c = malloc(found.len + 1);
        str_copy_to_c(found_c, found, found.len);
        bool result = test_fail(
            "wanted `%s` (len: `%zu`); "
            "found `%s` (len: `%zu`)",
            wanted_c,
            wanted.len,
            found_c,
            found.len);
        free(found_c);
        return result;
    }
    return true;
}

bool assert_count(const char *ctx, size_t wanted, size_t found)
{
    if (wanted != found)
    {
        return test_fail("%s: wanted `%zu`; found `%zu`", ctx, wanted, found);
    }
    return true;
}

bool assert_read(char *wanted_c, size_t nr, str found, result res)
{
    size_t len = strlen(wanted_c);
    str_slice(found, &found, 0, len);
    return assert_count("bytes read", len, nr) &&
           assert_ok(res) &&
           assert_str_eq(wanted_c, found);
}

bool assert_buffered_read(
    buffered_reader *br,
    str buf,
    char *wanted_c)
{
    result res;
    size_t nr = buffered_reader_read(br, buf, &res);
    return assert_read(wanted_c, nr, buf, res);
}

bool test_buffered_reader_read()
{
    test_init("test_buffered_reader_read");

    char srcalloc[] = "helloworld!";
    str src_str;
    str_init(&src_str, srcalloc, sizeof(srcalloc) - 1);

    str_reader src_str_reader;
    str_reader_init(&src_str_reader, src_str);

    reader src_reader;
    str_reader_to_reader(&src_str_reader, &src_reader);

    char internal_bufalloc[5] = {0};
    str internal_buffer;
    str_init(&internal_buffer, internal_bufalloc, sizeof(internal_bufalloc));

    buffered_reader br;
    buffered_reader_init(&br, src_reader, internal_buffer);

    char bufalloc[2] = {0};
    str buf;
    str_init(&buf, bufalloc, sizeof(bufalloc));

    result res;
    result_ok(&res);

#define ASSERT_BUFFERED_READ(wanted)               \
    if (!assert_buffered_read(&br, buf, (wanted))) \
    {                                              \
        return false;                              \
    }
    ASSERT_BUFFERED_READ("he");
    ASSERT_BUFFERED_READ("ll");
    ASSERT_BUFFERED_READ("ow");
    ASSERT_BUFFERED_READ("or");
    ASSERT_BUFFERED_READ("ld");
    ASSERT_BUFFERED_READ("!");
#undef ASSERT_BUFFERED_READ

    if (br.cursor != src_str.len % buf.len)
    {
        return test_fail(
            "cursor: wanted `%zu`; found `%zu`",
            src_str.len,
            br.cursor);
    }

    return test_success();
}

typedef struct
{
    string pre_match_data;
    writer w;
    str_reader src_str_reader;
    buffered_reader br;
    str match;
    result res;
} buffered_reader_find_test_case;

void buffered_reader_find_test_case_init(
    buffered_reader_find_test_case *tc,
    char *source_data,
    size_t source_data_size,
    char *buf,
    size_t buf_size,
    char *match_data,
    size_t match_data_size)
{
    string_init(&tc->pre_match_data);
    string_writer(&tc->w, &tc->pre_match_data);

    str src_str;
    str_init(&src_str, source_data, source_data_size);
    str_reader_init(&tc->src_str_reader, src_str);

    reader r;
    str_reader_to_reader(&tc->src_str_reader, &r);

    str buf_str;
    str_init(&buf_str, buf, buf_size);

    buffered_reader_init(&tc->br, r, buf_str);

    str_init(&tc->match, match_data, match_data_size);
}

void buffered_reader_find_test_case_drop(buffered_reader_find_test_case *tc)
{
    string_drop(&tc->pre_match_data);
}

bool buffered_reader_find_test_case_run(
    buffered_reader_find_test_case *tc,
    bool wanted_match,
    char *wanted_pre_match_data,
    size_t wanted_pre_match_data_size,
    char *wanted_post_match_data,
    size_t wanted_post_match_data_size,
    bool wanted_err)
{
    bool match = buffered_reader_find(&tc->br, tc->w, &tc->res, tc->match);
    if (match && !wanted_match)
    {
        return test_fail("unexpected match");
    }
    if (!match && wanted_match)
    {
        return test_fail("unexpected match failure");
    }

    str wanted_pre_match, pre_match;
    str_init(
        &wanted_pre_match,
        wanted_pre_match_data,
        wanted_pre_match_data_size);
    string_borrow(&tc->pre_match_data, &pre_match);
    if (!str_eq(wanted_pre_match, pre_match))
    {
        char wanted[256] = {0}, found[256] = {0};
        str_copy_to_c(wanted, wanted_pre_match, sizeof(wanted));
        str_copy_to_c(found, pre_match, sizeof(found));
        return test_fail(
            "pre-match data: wanted `%s` (len %zu); found `%s` (len %zu)",
            wanted,
            wanted_pre_match.len,
            found,
            pre_match.len);
    }

    // check post-match data
    string s;
    string_init(&s);
    TEST_DEFER(string_drop, &s);
    writer w;
    string_writer(&w, &s);
    reader r;
    buffered_reader_to_reader(&tc->br, &r);
    result copy_res;
    result_init(&copy_res);
    copy(w, r, &copy_res);
    if (!copy_res.ok)
    {
        string display;
        string_init(&display);
        TEST_DEFER(string_drop, &display);
        formatter f;
        string_formatter(&f, &display);
        error_display(copy_res.err, f);
        char message[256] = {0};
        string_copy_to_c(message, &display, sizeof(message));
        return test_fail(
            "unexpected error copying post-match result: %s",
            message);
    }

    str wanted_post_match, post_match;
    string_borrow(&s, &post_match);
    str_init(
        &wanted_post_match,
        wanted_post_match_data,
        wanted_post_match_data_size);
    if (!str_eq(wanted_post_match, post_match))
    {
        char wanted[256] = {0}, found[256] = {0};
        str_copy_to_c(wanted, wanted_post_match, sizeof(wanted));
        str_copy_to_c(found, post_match, sizeof(found));
        return test_fail(
            "post-match data: wanted `%s` (len: %zu); found `%s` (len %zu)",
            wanted,
            wanted_post_match.len,
            found,
            post_match.len);
    }

    if (wanted_err && tc->res.ok)
    {
        return test_fail("expected error, found no error");
    }

    if (!wanted_err && !tc->res.ok)
    {
        char message[256] = {0};
        string s;
        string_init(&s);
        formatter f;
        string_formatter(&f, &s);
        if (!error_display(tc->res.err, f))
        {
            return test_fail("unexpected error formatting error message");
        }
        string_copy_to_c(message, &s, sizeof(message));
        return test_fail("unexpected err: %s", message);
    }

    return test_success();
}

bool test_buffered_reader_find()
{
    test_init("test_buffered_reader_find");
    buffered_reader_find_test_case tc;
    char src[] = "foo bar baz";
    char buf[256] = {0};
    char match[] = "bar";
    buffered_reader_find_test_case_init(
        &tc,
        src, sizeof(src) - 1,
        buf, sizeof(buf),
        match, sizeof(match) - 1);
    TEST_DEFER(buffered_reader_find_test_case_drop, &tc);

    return buffered_reader_find_test_case_run(
        &tc,
        true,
        "foo ", sizeof("foo ") - 1,
        " baz", sizeof(" baz") - 1,
        false);
}

bool io_tests()
{
    return test_str_reader() &&
           test_copy() &&
           test_buffered_reader_read() &&
           test_buffered_reader_find();
}