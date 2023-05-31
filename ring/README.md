# RING

This is a toy ringbuffer implementation that I wrote to learn about ringbuffer
data structures. At the moment it only includes a slice-based implementation
which presumably has better cache locality properties), but in the future I
will add a doubly-linked-list-based implementation which will be more suitable
for something like an LRU cache (which swaps elements frequently).