#include "metadata.h"
#include "core/io/buffered_reader.h"
#include "core/io/match_reader.h"
#include "std/string/string_writer.h"

metadata metadata_new(string title, string date, vector tags)
{
    return (metadata){.title = title, .date = date, .tags = tags};
}

void __attribute__((constructor)) init()
{
    error_const(&ERR_MISSING_LEAD_FENCE, "missing lead frontmatter fence");
    error_const(&ERR_MISSING_TAIL_FENCE, "missing tail frontmatter fence");
}

const error ERR_MISSING_LEAD_FENCE;
const error ERR_MISSING_TAIL_FENCE;