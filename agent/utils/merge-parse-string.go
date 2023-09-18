package utils

import (
	"fmt"
)  
  
func MergeParseString(string1 string,  
	string2 string,  
	length int,  
) (string) {  
	concatenated := fmt.Sprintf("%s%s", string1, string2)  
	if length > len(concatenated) {  
		length = len(concatenated)  
	}  
	return concatenated[:length]  
}  
