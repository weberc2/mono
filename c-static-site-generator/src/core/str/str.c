#include "core/str/str.h"
#include "core/math/math.h"
#include "core/panic/panic.h"
#include <string.h>

str SPACE_CHARS;

static void __attribute__((constructor)) init()
{
    SPACE_CHARS = STR(" \t");
}

str str_new(char *data, size_t len)
{
    str s;
    str_init(&s, data, len);
    return s;
}

void str_init(str *s, char *data, size_t len)
{
    s->data = data;
    s->len = len;
}

str str_slice(str s, size_t start, size_t end)
{
    if (start > s.len)
    {
        panic(
            "slicing str with len `%zu`: start index `%zu` is out of "
            "bounds",
            s.len,
            start);
    }
    else if (end > s.len)
    {
        panic(
            "slicing str with len `%zu`: end index `%zu` is out of "
            "bounds",
            s.len,
            end);
    }
    else if (start > end)
    {
        panic(
            "slicing str with len `%zu`: start index `%zu` exceeds end index "
            "`%zu`",
            s.len,
            start,
            end);
    }

    return str_new(s.data + start, end - start);
}

size_t str_copy(str dst, str src)
{
    size_t sz = min(dst.len, src.len);
    memmove(dst.data, src.data, sz);
    return sz;
}

size_t str_copy_at(str dst, str src, size_t start)
{
    return str_copy(dst, str_slice(src, start, src.len));
}

bool str_eq(str lhs, str rhs)
{
    return lhs.len == rhs.len && memcmp(lhs.data, rhs.data, lhs.len) == 0;
}

bool str_has_prefix(str s, str prefix)
{
    if (prefix.len > s.len)
    {
        return false;
    }

    for (size_t i = 0; i < prefix.len; i++)
    {
        if (s.data[i] != prefix.data[i])
        {
            return false;
        }
    }

    return true;
}

size_t str_copy_to_c(char *dst, str src, size_t len)
{
    size_t copied = min(len, src.len);
    memmove(dst, src.data, copied);
    dst[copied] = '\0';
    return copied;
}

str str_trim_left(str s, str cutset)
{
    for (size_t start = 0; start < s.len; start++)
    {
        for (size_t j = 0; j < cutset.len; j++)
        {
            if (s.data[start] == cutset.data[j])
            {
                goto OUTER;
            }
        }

        return str_slice(s, start, s.len);

    OUTER:
        continue;
    }

    return s;
}

str str_trim_right(str s, str cutset)
{
    for (size_t cursor = s.len - 1; cursor < s.len; cursor--)
    {
        for (size_t j = 0; j < cutset.len; j++)
        {
            if (s.data[cursor] == cutset.data[j])
            {
                goto OUTER;
            }
        }

        return str_slice(s, 0, cursor + 1);

    OUTER:
        continue;
    }

    return s;
}

str str_trim(str s, str cutset)
{
    return str_trim_right(str_trim_left(s, cutset), cutset);
}

str str_trim_space_left(str s)
{
    return str_trim_left(s, SPACE_CHARS);
}

str str_trim_space_right(str s)
{
    return str_trim_right(s, SPACE_CHARS);
}

str str_trim_space(str s)
{
    return str_trim(s, SPACE_CHARS);
}

str_find_result str_find(str src, str match)
{
    if (src.len < match.len)
    {
        return (str_find_result){.found = false, .index = 0};
    }

    // the outer loop iterates through characters in the src string. If we get
    // to (src.len - match.len) and we still haven't matched the starting
    // character of the match string, then we've definitely failed.
    for (size_t start = 0; start <= src.len - match.len; start++)
    {
        // for each character in the src string, check to see if it is the
        // start of a match.
        for (size_t end = 0; end < match.len; end++)
        {
            if (src.data[start + end] != match.data[end])
            {
                goto RESET;
            }
        }

        // if we get here, we've successfully found the match.
        return (str_find_result){.found = true, .index = start};

    RESET:
        continue;
    }

    return (str_find_result){.found = false, .index = 0};
}

str_find_result str_find_char(str src, char match)
{
    for (size_t i = 0; i < src.len; i++)
    {
        if (src.data[i] == match)
        {
            return (str_find_result){.found = true, .index = i};
        }
    }

    return (str_find_result){.found = false, .index = 0};
}