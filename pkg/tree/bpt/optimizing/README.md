B+Tree Benchmarks
===

| key type | ns/op  | op/sec | B/op | allocs/op |
|:--------:|:------:|--------|:----:|:---------:|
|  string  | 66,205 | 15,104 |  2   |     0     |
|  []byte  | 66,596 | 15,015 |  2   |     0     |
| [8]byte  | 55,205 | 18,114 |  1   |     0     |
|  uint32  | 21,633 | 46,225 |  0   |     0     |
|  uint64  | 21,870 | 45,724 |  0   |     0     |

B+Tree Memory Details
===

### Adding 100K Elements

|     key type      | time to add | total alloc | system alloc | num gc |
|:-----------------:|:-----------:|:-----------:|:------------:|:------:|
|      uint64       |    26 ms    |    5 MB     |    17 MB     |   1    |
|      string       |    49 ms    |    8 MB     |    27 MB     |   2    |
| map[string]uint32 |    13 ms    |    7 MB     |    13 MB     |   2    |

---

### Adding 1 Million Elements

|         key type         | time to add | total alloc | system alloc | num gc |
|:------------------------:|:-----------:|:-----------:|:------------:|:------:|
|          uint64          |   174 ms    |    51 MB    |    64 MB     |   4    |
|          string          |   465 ms    |    84 MB    |    98 MB     |   5    |
|  map[string]uint32       |   243 ms    |   114 MB    |    92 MB     |   8    |

---

### Adding 10 Million Elements

|         key type         | time to add | total alloc | system alloc | num gc |
|:------------------------:|:-----------:|:-----------:|:------------:|:------:|
|          uint64          |   1.8 sec   |   507 MB    |    541 MB    |   8    |
|          string          |   4.8 sec   |   840 MB    |    884 MB    |   9    |
|  map[string]uint32       |   2.2 sec   |   929 MB    |    654 MB    |   13   |


*** note: the map[string]uint32 key type is just the entire data structure type in the case of a map, and is used for comparison in order to compare the relative space/speed between a b+tree and a swiss style hashmap