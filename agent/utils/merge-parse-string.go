package utils

import (
	"fmt"
)  
  
func MergeParseString(string1 string,  
	string2 string,  
	length int,
) (string) {  
	halfLength1 := len(string1) / 2  
	halfLength2 := len(string2) / 2  
  
	firstHalf := string1[:halfLength1]  
	secondHalf := string2[halfLength2:]  
  
	concatenated := fmt.Sprintf("%s%s", firstHalf, secondHalf)  
	if length > len(concatenated) {
				length = len(concatenated)
		}
	return concatenated[:length]
}
