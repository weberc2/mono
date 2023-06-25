#ifndef METADATA_H
#define METADATA_H

#include "core/error/error.h"
#include "core/result/result.h"
#include "core/io/reader.h"
#include "std/string/string.h"
#include "std/vector/vector.h"

typedef struct
{
    string title;
    string date;

    // tags is a vector of `str` each of which points into a store of shared
    // tags (so we don't have multiple instances for each tag's data).
    vector tags;
} metadata;

metadata metadata_new(string title, string date, vector tags);
// void metadata_parse(reader r, metadata *md, result *res);

#endif // METADATA_H