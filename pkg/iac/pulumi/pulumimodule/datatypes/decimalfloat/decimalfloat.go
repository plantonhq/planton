// Package decimalfloat widens a spec's float (a proto float, 32 bits) to the
// float64 a Pulumi input takes without changing the number the user wrote.
//
// float64(f) keeps the float32's binary value, which for most decimals is not
// the decimal: 0.3 becomes 0.30000001192092896. The provider sends that to
// the API, the API stores 0.3, and every later preview proposes changing it
// back -- a re-plan that is never clean. The Terraform modules never see the
// problem, because the tfvars carry the decimal the JSON encoder writes. Use
// Float64 wherever a module passes a fractional spec float to Pulumi.
package decimalfloat

import "strconv"

// Float64 is the float64 whose shortest decimal form is f's: the number as
// the manifest wrote it.
func Float64(f float32) float64 {
	value, err := strconv.ParseFloat(strconv.FormatFloat(float64(f), 'g', -1, 32), 64)
	if err != nil {
		// FormatFloat always yields a number ParseFloat reads; NaN and the
		// infinities round-trip too.
		return float64(f)
	}
	return value
}
