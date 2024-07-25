// Package error_code @Author Adrian.Wang 2024/7/25 下午3:14:00
package error_code

// ErrorCode 自定义错误码
type ErrorCode uint32

const (
	SUCCESS        ErrorCode = 200   //成功
	ERROR          ErrorCode = 500   //服务器错误
	AUTH_ERROR     ErrorCode = 401   // 鉴权失败
	INVALID_PARAMS ErrorCode = 10001 //参数错误
)

// CodeMsgMap 错误码映射错误信息
var CodeMsgMap = map[ErrorCode]string{
	SUCCESS:        "ok",
	ERROR:          "fail",
	AUTH_ERROR:     "鉴权失败",
	INVALID_PARAMS: "请求参数错误",
}
