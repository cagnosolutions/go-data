Closed Hashing (Open Addressing) Hash Map
---
This hash map implementation uses a closed hashing (open addressing) technique with
linear probing for resolving any hash collisions. The exact algorithm it utilizes
is called 'robin hood hashing.' More information about this can technique can be found
in the links at the bottom.

### Basic Idea Behind the Robin Hood Hashing Technique
1) Calculate the hash key value
2) Calculate the initial index of the entry to be inserted (using the hash key)
 - If the index is empty, insert the entry
 - Otherwise, start scanning linearly from the index found. Keep count and when <br>
   you finally find a (close by) index to insert the entry into, store the count <br>
   of the distance from the initial index/bucket. This is called the DIB.
 - *Note, the entry is always stored along with the DIB, even if it uses the first
   index found.* 
3) If we encounter an entry (at any point during step 2) which has a DIB less than <br>
   the one of the entry to be inserted, swap them and continue. 

To conclude, linear probing along with the storage of the DIB and algorithm that <br>
ultimately ends up as *Robin Hood Hashing* works by improving the locality of <br>
reference between all the entries resulting in a more stable average between them.

*For more information about closed hashing (and robin hood hashing in particular)*
- [Robin Hood Hashing](https://andre.arko.net/2017/08/24/robin-hood-hashing/)
- [Robin Hood Hashing, Pedro Celis, 1986](https://cs.uwaterloo.ca/research/tr/1986/CS-86-14.pdf)
- [Distributional Analysis of Robin Hood Linear Probing Hashing with Buckets](https://www.dmtcs.org/pdfpapers/dmAD0127.pdf)
- [Numerical Experiments in Hashing, Part 1](https://www.pvk.ca/Blog/numerical_experiments_in_hashing.html)
- [Numerical Experiments in Hashing, Part 2](https://www.pvk.ca/Blog/more_numerical_experiments_in_hashing.html)
- [Robin Hood Hashing Should be Your Default Hash Table Implementation](https://www.sebastiansylvan.com/post/robin-hood-hashing-should-be-your-default-hash-table-implementation/)
- [More On Robin Hood Hashing](https://www.sebastiansylvan.com/post/more-on-robin-hood-hashing-2/)
- [Robin Hood Hashing, Part 1](http://codecapsule.com/2013/11/11/robin-hood-hashing/)
- [Robin Hood Hashing (Backwards Shift Deletion), Part 2](http://codecapsule.com/2013/11/17/robin-hood-hashing-backward-shift-deletion/)
- [The Other Robin Hood Hashing](https://www.pvk.ca/Blog/2013/11/26/the-other-robin-hood-hashing/)
