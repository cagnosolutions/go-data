package main

var data = []byte(`Note: The source for bufio.Reader is at
source: https://cs.opensource.google/go/go/+/refs/tags/go1.18:src/bufio/bufio.go;l=62
So lets say we want use a bufio.Reader. We can
see that the function signature for it looks like
this: bufio.NewReader(rd io.Reader) *bufio.Reader
So it takes an io.Reader (interface) and returns
a concrete type of *bufio.Reader.

Okay, so let's say I have a file I want to read.
How do I do i use bufio.Reader with it? Well, an
*os.File implements the io.Reader interface, so
you can just pass the file into the *bufio.Reader
directly, ie:

file, err := os.Open("myfile")
if err != nil {
	panic(err)
}
r := bufio.NewReader(file)

Now you can call the methods of r, and do your thing
and simply forget about interacting with file at all.

Okay, so what if I just have a string or a byte
slice in memory and I want to use a *bufio.Reader
to read it with? How would I do that? Glad you
asked. Let's check that out. Hmm, looks like we
have a bit of an issue. A string or []byte doesn't
seem to "implement" the io.Reader interface. So
we can't do bufio.NewReader([]byte("my byte data"))
so what are we to do? Well, the "bytes" and "strings"
packages have a lot of utilities for working with
raw bytes and strings. Also they are mirror packages,
as in they pretty much have all of the same exact
functions and methods in both packages. One for
strings, and one for bytes.

Ahh, look at this! I found a bytes.Reader type
in the bytes package, and a strings.Reader type
in the strings package. Okay, let's look at the
bytes.NewReader(b []byte) *bytes.Reader function.
Ahh, so it appears that it takes a []byte, and
returns a concrete *bytes.Reader type. This type
is just a simple struct that wraps the []byte
and implements the io.Reader interface. That means
we can pass it into a *bufio.Reader. Let's see
what that looks like.

data := []byte("some crazy data up in here\n")
br := bytes.NewReader(data)
r := bufio.NewReader(br)
	
Now you can call the methods of r, and do your
thing and simply forget about interacting with
the *bytes.Reader, or the []byte.

Okay, so let's try and read this text that I 
typed up and see what we can do.
.
`)
