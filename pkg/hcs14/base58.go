package hcs14

const base58Alphabet = "123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz"

func base58Encode(input []byte) string {
	if len(input) == 0 {
		return ""
	}

	zeros := 0
	for zeros < len(input) && input[zeros] == 0 {
		zeros++
	}

	if zeros == len(input) {
		return repeatString("1", zeros)
	}

	digits := []int{0}
	for index := zeros; index < len(input); index++ {
		carry := int(input[index])
		for digitIndex := 0; digitIndex < len(digits); digitIndex++ {
			value := (digits[digitIndex] << 8) + carry
			digits[digitIndex] = value % 58
			carry = value / 58
		}
		for carry > 0 {
			digits = append(digits, carry%58)
			carry /= 58
		}
	}

	output := repeatString("1", zeros)
	for index := len(digits) - 1; index >= 0; index-- {
		output += string(base58Alphabet[digits[index]])
	}
	return output
}

func repeatString(value string, count int) string {
	result := ""
	for index := 0; index < count; index++ {
		result += value
	}
	return result
}

func base58Decode(input string) ([]byte, error) {
	if len(input) == 0 {
		return []byte{}, nil
	}

	zeros := 0
	for zeros < len(input) && input[zeros] == '1' {
		zeros++
	}

	output := make([]int, 0)
	for index := zeros; index < len(input); index++ {
		character := input[index]
		value := indexByte(base58Alphabet, character)
		if value < 0 {
			return nil, ErrInvalidBase58Character
		}

		carry := value
		for outputIndex := 0; outputIndex < len(output); outputIndex++ {
			x := output[outputIndex]*58 + carry
			output[outputIndex] = x & 0xff
			carry = x >> 8
		}
		for carry > 0 {
			output = append(output, carry&0xff)
			carry >>= 8
		}
	}

	for index := 0; index < zeros; index++ {
		output = append(output, 0)
	}

	for left, right := 0, len(output)-1; left < right; left, right = left+1, right-1 {
		output[left], output[right] = output[right], output[left]
	}

	decoded := make([]byte, len(output))
	for index := range output {
		decoded[index] = byte(output[index])
	}
	return decoded, nil
}

func decodeMultibaseB58btc(value string) ([]byte, error) {
	if len(value) < 2 || value[0] != 'z' {
		return nil, ErrInvalidMultibase
	}
	return base58Decode(value[1:])
}

func indexByte(input string, target byte) int {
	for index := 0; index < len(input); index++ {
		if input[index] == target {
			return index
		}
	}
	return -1
}
