package lib

import (
	"fmt"
	"testing"
)

// 测试GetRetCodePlain
func TestGetRetCodePlain(t *testing.T) {
	var retCode RetCode = 9999
	fmt.Printf("RET_CODE_SUCCESS : %v\n", GetRetCodePlain(RET_CODE_SUCCESS))
	fmt.Printf("RET_CODE_WARNING_CALL_TIMEOUT: %v\n", GetRetCodePlain(RET_CODE_WARNING_CALL_TIMEOUT))
	fmt.Printf("RET_CODE_ERROR_CALL : %v\n", GetRetCodePlain(RET_CODE_ERROR_CALL))
	fmt.Printf("RET_CODE_ERROR_RESPONSE : %v\n", GetRetCodePlain(RET_CODE_ERROR_RESPONSE))
	fmt.Printf("RET_CODE_ERROR_CALEE : %v\n", GetRetCodePlain(RET_CODE_ERROR_CALEE))
	fmt.Printf("RET_CODE_FATAL_CALL : %v\n", GetRetCodePlain(RET_CODE_FATAL_CALL))
	fmt.Printf("UNKONW ERROR CODE: %v\n", GetRetCodePlain(retCode))
}
