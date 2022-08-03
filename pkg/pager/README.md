### This is a sample of what the pager (**API**) may end up like

#### also, latest notes here: https://go.dev/play/p/bL4AJCDANF3

![https://docs.google.com/drawings/d/e/2PACX-1vQ6CkIChGCWjAR9OAkjasoQzdmd6IRpHCkMkZP12a5VQ163B5BrfvPmaQ_EO8U1yUaTgklucTg29c0y/pub?w=1912&h=790](https://docs.google.com/drawings/d/e/2PACX-1vQ6CkIChGCWjAR9OAkjasoQzdmd6IRpHCkMkZP12a5VQ163B5BrfvPmaQ_EO8U1yUaTgklucTg29c0y/pub?w=1912&h=790)

If we use 8 pages ber block, 8 blocks per extent, and 8 extents per segment we get a beautiful
layout of 2mb segment sizes with 4kb page sizes, and 32mb segment sizes with 64kb

| Page Size | Initial Segment Size | Max Segment Size | 
|:---------:|:--------------------:|:----------------:|
|   4KiB    |        256kb         |       2mb        |
|   8KiB    |        512kb         |       4mb        |
|   16KiB   |         1mb          |       8mb        |
|   32KiB   |         2mb          |       16mb       |
|   64KiB   |         4mb          |       32mb       |


![foo bar](../util/pager-sample.png)

***source and credit goes to*** [github.com/tidwall/bfile](https://github.com/tidwall/bfile)