package bits

// BiIsEven returns true if the provided unsigned
// integer is even, and false if it is odd.
func BiIsEven(i uint) bool {
	return i&1 == 0
}

// BiAdd returns the sum of two unsigned integers
// using i and j as addends.
func BiAdd(i, j uint) uint {
	var k, r uint
	k = (i & j) << 1
	r = i ^ j
	// if k == 0, there is a carry.
	if k == 0 {
		return r
	}
	// otherwise, recurse
	return BiAdd(k, r)
}

// BiSub returns the difference of two unsigned integers
// using i as the minuend and j as the subtrahend.
func BiSub(i, j uint) uint {
	var bor uint
	for j > 0 {
		bor = (^i) & j
		i = i ^ j
		j = bor << 1
	}
	return i
}

// BiMul returns the product of two unsigned integers
// using i as the multiplicand and j as the multiplier.
func BiMul(i, j uint) uint {
	var r uint
	for j > 0 {
		if j&1 > 0 {
			r += i
		}
		i = i << 1
		j = j >> 1
	}
	return r
}

// BiDiv returns the quotient of two unsigned integers
// using i as the dividend and j as the divisor.
func _BiDiv(i, j uint) uint {
	// So a binary division works from the high order digits to the low
	// order digits and generates a quotient with each step. Very similar
	// to how long division works on paper but, this will be split into two
	// steps.
	// Step 1: We shift the upper bits of the dividend into the remainder.
	var bit, nbits uint
	var rem, d uint
	for rem < j {
		bit = (i & 0x80000000) >> 31
		rem = (rem << 1) | bit
		d = i
		i = i << 1
		nbits--
	}
	// Loop above takes it one op too far so, we reverse the last iteration.
	i = d
	rem = rem >> 1
	nbits++
	// Step 2: Subtract the divisor from the value in the remainder. The high
	// order bit of the result becomes a bit of the quotient.
	var tmp, quo, q uint
	for k := uint(0); k < nbits; k++ {
		bit = (i & 0x80000000) >> 31
		rem = (rem << 1) | bit
		tmp = rem - j
		q = ^((tmp & 0x80000000) >> 31)
		i = i << 1
		quo = (quo << 1) | q
		if q != 0 {
			rem = tmp
		}
	}
	return quo
}

func BiDiv(n, d uint) uint {
	// n is dividend, d is divisor, q holds the result.
	var q uint = 0
	// as long as the divisor fits into the remainder there is something to do...
	for n >= d {
		var i, dT uint = 0, d
		// determine to which power of two the divisor still fits the dividend;
		// our intention is to subtract the divisor multiplied by powers of two
		// yielding us a one in the binary representation of the result.
		for ; n >= (dT << 1); i++ {
			dT <<= 1
		}
		// set the corresponding bit in the result
		q |= 1 << i
		// subtract the multiple of the divisor to be left with the remainder
		n -= dT
		// repeat until the divisor does not fit into the remainder anymore
	}
	return q
}
