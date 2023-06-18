#ifndef BYTESLICE_READER_H
#define BYTESLICE_READER_H

#include "byteslice/byteslice.h"
#include "reader.h"

typedef struct
{
    byteslice buffer;
    size_t cursor;
} byteslice_reader;

void byteslice_reader_init(byteslice_reader *br, byteslice buffer);
size_t byteslice_reader_read(byteslice_reader *br, byteslice buffer);
void byteslice_reader_to_reader(byteslice_reader *br, reader *out);

#endif // BYTESLICE_READER_H