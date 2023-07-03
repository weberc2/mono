#include "core/testing/test.h"
#include "core/io/str_reader.h"

#include "core/config_parser/config_parser.h"

#define BUFFER(sz, contents) STR_ARR((char[sz]){contents})

#define LIT_READER(input)                            \
    (reader)                                         \
    {                                                \
        .data = (void *)&STR_READER(STR_LIT(input)), \
        .read = (read_func)str_reader_io_read,       \
    }

#define PARSER(inpt, bfsz, bfcontents, cs, lre, st) \
    (config_parser)                                 \
    {                                               \
        .source = LIT_READER(inpt),                 \
        .buffer = BUFFER(bfsz, bfcontents),         \
        .cursor = (cs),                             \
        .last_read_size = (lre),                    \
        .state = (st),                              \
    }

#define PARSER_NEW(inpt, bfsz) \
    PARSER(inpt, bfsz, 0, 0, 0, config_parser_state_start)

#define PARSER_NEW_W_STATE(inpt, bfsz, st) \
    PARSER(inpt, bfsz, 0, 0, 0, st)

#define CPS(rest) config_parser_state_##rest

typedef struct test_case
{
    char *name;
    config_parser parser;
    config_parser_result (*next_func)(config_parser *);
    config_parser_result wanted;
} test_case;

test_case test_cases[] = {
    {
        .name = "test_config_parser_key_next:start-to-eof-empty",
        .parser = PARSER_NEW("", 1),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_EOF(STR_EMPTY),
    },
    {
        .name = "test_config_parser_key_next:start-to-parsed",
        .parser = PARSER_NEW("a:", 2),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSED_KEY(STR_LIT("a")),
    },
    {
        .name = "test_config_parser_key_next:start-to-parsing",
        .parser = PARSER_NEW("abc:", 3),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSING_KEY(STR_LIT("abc")),
    },
    {
        .name = "test_config_parser_key_next:parsing-to-parsing",
        .parser = PARSER("defghi:", 3, "abc", 3, 3, CPS(parsing_key)),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSING_KEY(STR_LIT("def")),
    },
    {
        .name = "test_config_parser_key_next:parsing-to-parsed-empty",
        .parser = PARSER(":", 3, "abc", 3, 3, CPS(parsing_key)),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSED_KEY(STR_EMPTY),
    },
    {
        .name = "test_config_parser_key_next:parsing-to-parsed-not-empty",
        .parser = PARSER("de:", 3, "abc", 3, 3, CPS(parsing_key)),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSED_KEY(STR_LIT("de")),
    },
    {
        .name = "test_config_parser_key_next:eof-to-eof",
        .parser = PARSER_NEW_W_STATE("", 3, CPS(eof)),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_EOF(STR_EMPTY),
    },
    {
        .name = "test_config_parser_key_next:start-to-parse-error",
        .parser = PARSER_NEW("foo\n:bar", 10),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSE_ERROR,
    },
    {
        .name = "test_config_parser_key_next:parse-error-to-parse-error",
        .parser = PARSER_NEW_W_STATE("", 10, CPS(parse_error)),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSE_ERROR,
    },
    {
        .name = "test_config_parser_key_next:io-error-to-io-error",
        .parser = PARSER_NEW_W_STATE("", 10, CPS(io_error)),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_IO_ERROR(STR_EMPTY, ERROR_NULL),
    },
    {
        .name = "test_config_parser_key_next:parsed-key-to-parsed-key",
        .parser = PARSER_NEW_W_STATE("bar", 10, CPS(parsed_key)),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSED_KEY(STR_EMPTY),
    },
    {
        .name = "test_config_parser_key_next:parsing-value-to-parsing-value",
        .parser = PARSER_NEW_W_STATE("bar", 10, CPS(parsing_value)),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSING_VALUE(STR_EMPTY),
    },
    {
        .name = "test_config_parser_key_next:parsed-value-to-parsing-key",
        .parser = PARSER_NEW_W_STATE("bar:", 3, CPS(parsed_value)),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSING_KEY(STR_LIT("bar")),
    },
    {
        .name = "test_config_parser_key_next:parsed-value-to-parsed-key",
        .parser = PARSER_NEW_W_STATE("bar:", 4, CPS(parsed_value)),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSED_KEY(STR_LIT("bar")),
    },
    {
        .name = "test_config_parser_key_next:skip-leading-space-single-buf",
        .parser = PARSER_NEW_W_STATE("  \tfoo:", 10, CPS(start)),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSED_KEY(STR_LIT("foo")),
    },
    {
        .name = "test_config_parser_key_next:skip-leading-space-multi-bufs",
        .parser = PARSER_NEW_W_STATE("  \t  \t  \tf:", 3, CPS(start)),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSED_KEY(STR_LIT("f")),
    },
    {
        .name = "test_config_parser_key_next:skip-blank-lines",
        .parser = PARSER_NEW_W_STATE("\n\n\tf:", 3, CPS(start)),
        .next_func = config_parser_key_next,
        .wanted = CONFIG_PARSER_PARSED_KEY(STR_LIT("f")),
    },
    {
        .name = "test_config_parser_value_next:eof-to-eof",
        .parser = PARSER_NEW_W_STATE("", 4, CPS(eof)),
        .next_func = config_parser_value_next,
        .wanted = CONFIG_PARSER_EOF(STR_EMPTY),
    },
    {
        .name = "test_config_parser_value_next:io-error-to-io-error",
        .parser = PARSER_NEW_W_STATE("", 4, CPS(io_error)),
        .next_func = config_parser_value_next,
        .wanted = CONFIG_PARSER_IO_ERROR(STR_EMPTY, ERROR_NULL),
    },
    {
        .name = "test_config_parser_value_next:parse-error-to-parse-error",
        .parser = PARSER_NEW_W_STATE("", 4, CPS(parse_error)),
        .next_func = config_parser_value_next,
        .wanted = CONFIG_PARSER_PARSE_ERROR,
    },
    {
        .name = "test_config_parser_value_next:parsed-value-to-parsed-value",
        .parser = PARSER_NEW_W_STATE("asdf:", 4, CPS(parsed_value)),
        .next_func = config_parser_value_next,
        .wanted = CONFIG_PARSER_PARSED_VALUE(STR_EMPTY),
    },
    {
        .name = "test_config_parser_value_next:start-to-start",
        .parser = PARSER_NEW_W_STATE("asdf:", 4, CPS(start)),
        .next_func = config_parser_value_next,
        .wanted = CONFIG_PARSER_START,
    },
    {
        .name = "test_config_parser_value_next:parsing-key-to-parsing-key",
        .parser = PARSER_NEW_W_STATE("asdf:", 4, CPS(parsing_key)),
        .next_func = config_parser_value_next,
        .wanted = CONFIG_PARSER_PARSING_KEY(STR_EMPTY),
    },
    {
        .name = "test_config_parser_value_next:parsed-key-to-parsing-value",
        .parser = PARSER_NEW_W_STATE("world", 4, CPS(parsed_key)),
        .next_func = config_parser_value_next,
        .wanted = CONFIG_PARSER_PARSING_VALUE(STR_LIT("worl")),
    },
    {
        .name = "test_config_parser_value_next:parsed-key-to-parsed-value",
        .parser = PARSER_NEW_W_STATE("bar\nbaz", 4, CPS(parsed_key)),
        .next_func = config_parser_value_next,
        .wanted = CONFIG_PARSER_PARSED_VALUE(STR_LIT("bar")),
    },
    {
        .name = "test_config_parser_value_next:parsing-value-to-parsing-value",
        .parser = PARSER_NEW_W_STATE("helloworld", 4, CPS(parsing_value)),
        .next_func = config_parser_value_next,
        .wanted = CONFIG_PARSER_PARSING_VALUE(STR_LIT("hell")),
    },
    {
        .name = "test_config_parser_value_next:parsing-value-to-parsed-value",
        .parser = PARSER_NEW_W_STATE("bar\nbaz", 4, CPS(parsing_value)),
        .next_func = config_parser_value_next,
        .wanted = CONFIG_PARSER_PARSED_VALUE(STR_LIT("bar")),
    },
    {
        .name = "test_config_parser_value_next:skip-leading-space-single-buf",
        .parser = PARSER_NEW_W_STATE("  \tfoo\n", 10, CPS(parsed_key)),
        .next_func = config_parser_value_next,
        .wanted = CONFIG_PARSER_PARSED_VALUE(STR_LIT("foo")),
    },
    {
        .name = "test_config_parser_value_next:skip-leading-space-multi-bufs",
        .parser = PARSER_NEW_W_STATE("  \t  \t  \tf\n", 3, CPS(parsed_key)),
        .next_func = config_parser_value_next,
        .wanted = CONFIG_PARSER_PARSED_VALUE(STR_LIT("f")),
    },
    {
        // NOTE: a newline following a sequence of leading spaces is still an
        // end-of-value delimiter, so the sequence ` \t\n` will be interpreted
        // as a zero-length value.
        .name = "test_config_parser_value_next:skip-leading-space-before-eol",
        .parser = PARSER_NEW_W_STATE(" \t\n\n\tf\n", 3, CPS(parsed_key)),
        .next_func = config_parser_value_next,
        .wanted = CONFIG_PARSER_PARSED_VALUE(STR_EMPTY),
    },
};

static bool assert_result_eq(config_parser_result w, config_parser_result f)
{
    if (w.state != f.state)
    {
        return test_fail(
            "state: wanted `%s`; found `%s`",
            config_parser_state_to_str(w.state).data,
            config_parser_state_to_str(f.state).data);
    }

    if (!str_eq(w.bytes, f.bytes))
    {
        char w_[256] = {0}, f_[256] = {0};
        str_copy_to_c(w_, w.bytes, sizeof(w_));
        str_copy_to_c(f_, f.bytes, sizeof(f_));
        return test_fail("bytes: wanted `%s`; found `%s`", w_, f_);
    }

    return true;
}

#define ASSERT_RESULT_EQ(w, f)       \
    if (!assert_result_eq((w), (f))) \
    {                                \
        return false;                \
    }

bool test_case_run(test_case *tc)
{
    test_init(tc->name);
    config_parser_result found = tc->next_func(&tc->parser);
    ASSERT_RESULT_EQ(tc->wanted, found);
    return test_success();
}

bool test_config_parser()
{
    for (size_t i = 0; i < sizeof(test_cases) / sizeof(test_case); i++)
    {
        if (!test_case_run(&test_cases[i]))
        {
            return false;
        }
    }

    return true;
}