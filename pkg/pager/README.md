### This is a sample of what the pager (**API**) may end up like

#### also, latest notes here: https://go.dev/play/p/bL4AJCDANF3

![https://docs.google.com/drawings/d/e/2PACX-1vQ6CkIChGCWjAR9OAkjasoQzdmd6IRpHCkMkZP12a5VQ163B5BrfvPmaQ_EO8U1yUaTgklucTg29c0y/pub?w=1912&h=790](https://docs.google.com/drawings/d/e/2PACX-1vQ6CkIChGCWjAR9OAkjasoQzdmd6IRpHCkMkZP12a5VQ163B5BrfvPmaQ_EO8U1yUaTgklucTg29c0y/pub?w=1912&h=790)

If we use 8 pages ber block, 8 blocks per extent, and 8 extents per segment we get a beautiful
layout of 2mb segment sizes with 4kb page sizes, and 32mb segment sizes with 64kb

| Page Size |  Initial Segment Size  |  Max Segment Size  | 
|:---------:|:----------------------:|:------------------:|
|   4 KiB   |        256 KiB         |       2 MiB        |
|   8 KiB   |        512 KiB         |       4 MiB        |
|  16 KiB   |         1 MiB          |       8 MiB        |
|  32 KiB   |         2 MiB          |       16 MiB       |
|  64 KiB   |         4 MiB          |       32 MiB       |

## Consider doing what MS SQL does 

- [pages-and-extents-architecture-large-row-support](https://docs.microsoft.com/en-us/sql/relational-databases/pages-and-extents-architecture-guide?view=sql-server-ver16#large-row-support)
<br>
- [maximum-capacity-specifications-for-sql-server](https://docs.microsoft.com/en-us/sql/sql-server/maximum-capacity-specifications-for-sql-server?view=sql-server-ver16)
<br>

| Page Size | Pages Per Extent | Extents Per MiB |
|:---------:|:----------------:|:---------------:|
|   8 KiB   |   8  (64 KiB)    |       16        |

And other inspirations...
---

![foo bar](../util/pager-sample.png)

***source and credit goes to*** [github.com/tidwall/bfile](https://github.com/tidwall/bfile)