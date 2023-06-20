#ifndef COPY_H
#define COPY_H

#include "reader.h"
#include "writer.h"
#include "io_result.h"

size_t copy(writer dst, reader src, io_result *res);

#endif // COPY_H