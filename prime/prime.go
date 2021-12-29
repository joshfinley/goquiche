package prime

type Prime uint64

// get the next prime above n
func (p Prime) Next() Prime {
	// base case
	if p <= 1 {
		return 2
	}

	candidate := p
	found := false
	for !found {
		candidate++
		if IsPrime(uint(candidate)) {
			found = true
		}
	}

	return Prime(candidate)
}

// check if the given number is prime
func IsPrime(n uint) bool {
	// Corner cases
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}

	// This is checked so that we can skip
	// middle five numbers in below loop
	if n%2 == 0 || n%3 == 0 {
		return false
	}

	for i := uint(5); i*i <= n; i = i + 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}

	}
	return true
}
