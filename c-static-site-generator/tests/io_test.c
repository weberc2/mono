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

    size_t nc = copy(w, r, &res);
    if (nc != sizeof(srcdata) - 1)
    {
        return test_fail(
            "bytes copied: wanted `%zu`; found `%zu`",
            sizeof(srcdata) - 1,
            nc);
    }

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

bool test_buffered_reader_read__partial_rewind()
{
    test_init("test_buffered_reader_read__partial_rewind");
    char src_[] = "foo";
    char innerbuf_[128] = {0};
    char outerbuf_[256] = {0};

    str src, innerbuf, outerbuf;
    str_init(&src, src_, sizeof(src_) - 1);
    str_init(&innerbuf, innerbuf_, sizeof(innerbuf_) - 1);
    str_init(&outerbuf, outerbuf_, sizeof(outerbuf_) - 1);

    str_reader src_str_reader;
    reader r;
    buffered_reader br;
    str_reader_init(&src_str_reader, src);
    str_reader_to_reader(&src_str_reader, &r);
    buffered_reader_init(&br, r, innerbuf);

    result res;
    result_init(&res);
    size_t nr = buffered_reader_read(&br, outerbuf, &res);
    ASSERT_OK(res);

    if (nr != sizeof(src_) - 1)
    {
        return test_fail(
            "bytes read: wanted `%zu`; found `%zu`",
            sizeof(src_) - 1,
            nr);
    }

    str found;
    str_slice(outerbuf, &found, 0, nr);
    if (!str_eq(src, found))
    {
        char found_[256] = {0};
        str_copy_to_c(found_, found, sizeof(found_));
        return test_fail("wanted `%s`; found `%s`", src_, found_);
    }

    // try rewinding, but not all the way
    size_t new_cursor = 1;
    br.cursor = new_cursor;
    nr = buffered_reader_read(&br, outerbuf, &res);
    ASSERT_OK(res);

    if (nr != sizeof(src_) - 1 - new_cursor)
    {
        return test_fail(
            "bytes read: wanted `%zu`; found `%zu`",
            sizeof(src_) - 1 - new_cursor,
            nr);
    }

    str wanted;
    str_slice(outerbuf, &found, 0, nr);
    str_slice(src, &wanted, new_cursor, src.len);
    if (!str_eq(wanted, found))
    {
        char found_[256] = {0};
        str_copy_to_c(found_, found, sizeof(found_));
        return test_fail("wanted `%s`; found `%s`", src_ + new_cursor, found_);
    }

    return test_success();
}

bool io_tests()
{
    return test_str_reader() &&
           test_copy() &&
           test_buffered_reader_read() &&
           test_buffered_reader_read__partial_rewind();
}