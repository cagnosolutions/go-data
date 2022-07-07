package util

/* Contains various function wrappers for commonly used operations */

func Ternary[E comparable](cond bool, exp1, exp2 E) E {
	if cond {
		return exp1
	}
	return exp2
}

func TernaryAny(cond bool, exp1, exp2 any) any {
	if cond {
		return exp1
	}
	return exp2
}

func TernaryFunc(cond bool, exp1, exp2 func()) func() {
	if cond {
		return exp1
	}
	return exp2
}

func TernaryInt(cond bool, exp1, exp2 int) int {
	if cond {
		return exp1
	}
	return exp2
}

func TernaryString(cond bool, exp1, exp2 int) int {
	if cond {
		return exp1
	}
	return exp2
}
