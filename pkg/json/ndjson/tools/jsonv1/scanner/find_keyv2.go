package scanner

import (
	"bytes"
)

// FindKey2 accepts a JSON object and returns the value associated with the key specified
func FindKey2(in []byte, k []byte) ([]byte, error) {
	// Initialize error and type variable to use throughout
	var err error
	var typ jsonType
	// The start variable will be available to hold our start position for each type
	// we scan through. Same for the end. And depth is there to measure any object
	// nesting that we may have to do.
	var start, end int
	// Skip past any initial space that may be present. We are looking for the beginning
	// of a JSON object, denoted by a left brace. `{`
	end, err = SkipSpace(in, start)
	if err != nil {
		return nil, err
	}
	start = end
	// We have skipped past the space, so we should be at the start of an object. Check
	// to make sure that is the case.
	if v := in[start]; v != '{' {
		// If we have not found the start of an object, throw an error for now. But it
		// must be noted that *technically* we may be starting with an array of objects
		// and not know it. In which case we would need to look for a left bracket. `[`
		return nil, NewError(start, v)
	}
	// If we reach this point, we have found our opening left brace '{'. We should now
	// increment the positional counter and then go into our loop.
	start++
	// We are now inside a JSON object.
	for {
		// Skip any leading whitespace, then look for a string.
		end, err = SkipSpace(in, start)
		if err != nil {
			return nil, err
		}
		start = end

		// Our key will be a string, so lets see if we have it by trying to return
		// the ending position of a string. If we do not encounter an error we
		// have found it.
		end, err = String(in, start)
		if err != nil {
			return nil, err
		}
		// We have successfully identified our first key.
		key := in[start+1 : end-1]

		// Check it against our supplied key to determine if we have a match and
		// store the result of our potential match for later.
		match := bytes.Equal(k, key)
		// It might be worth noting here that maybe we should check to see if we
		// have a match sooner than later, like now. And if we do not have a match,
		// then we can potentially make a choice when we get to our any value call.
		start = end

		// Next, skip past any potential whitespace.
		end, err = SkipSpace(in, start)
		if err != nil {
			return nil, err
		}
		start = end

		// Look for a colon. If we do not find one, return an error.
		end, err = Expect(in, start, ':')
		if err != nil {
			return nil, err
		}
		start = end
		// Otherwise, consume it and continue.

		// Skip past any potential whitespace.
		end, err = SkipSpace(in, start)
		if err != nil {
			return nil, err
		}
		start = end
		// We are now on the lookout for a value.

		// Our value could be of any type, but we think we have it in our sights
		// and the best way to find out is to try and find the end of it without
		// encountering any errors.
		typ, end, err = Value(in, start)
		if err != nil {
			return nil, err
		}
		// We must have found it, because we were not met with any error.
		// Ignore type for now
		_ = typ

		// Now, lets check to see if we have a match.
		if match {
			// If we do, we will return the value that we have isolated.
			return in[start:end], nil
		}
		start = end

		// Otherwise, we did not have a matching key. So we must continue on to
		// inspect more keys. So, we must skip past any potential whitespace.
		end, err = SkipSpace(in, start)
		if err != nil {
			return nil, err
		}
		start = end

		// After which, we will either be met with a comma, indicating that we
		// have more keys to inspect, or the end of the JSON object.
		switch in[start] {
		case ',':
			// More keys to inspect, so lets increment our positional counter
			// and start the loop over.
			start++
		case '}':
			// Oh no, we have found the end of a JSON object. There currently
			// I have not found a good way to go about object nesting, so it
			// looks like this is the end of the road, for now. There is nowhere
			// else to go, and we have not located our matching key. So sorry,
			// we must simply return an error.
			return nil, ErrKeyNotFound
		}
	}
}
