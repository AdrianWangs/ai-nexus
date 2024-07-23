package response

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

type Body struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func Response(r *http.Request, w http.ResponseWriter, res any, err error) {

	if err != nil {
		body := Body{
			Code: 500,
			Msg:  err.Error(),
			Data: res,
		}
		httpx.WriteJson(w, http.StatusOK, body)
		return
	}

	body := Body{
		Code: 200,
		Msg:  "success",
		Data: res,
	}
	httpx.WriteJson(w, http.StatusOK, body)

}
