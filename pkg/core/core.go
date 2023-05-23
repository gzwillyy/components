package core

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gzwillyy/components/errors"
	"github.com/gzwillyy/components/log"
)

// ErrResponse 定义发生错误时的返回消息.
// 如果 Reference 不存在，将被忽略.
// swagger:model
type ErrResponse struct {
	// 代码定义业务错误代码.
	Code int `json:"code"`

	// 消息包含此消息的详细信息
	// 此消息适合暴露给外部
	Message string `json:"message"`

	// Reference 返回可能对解决此错误有用的参考文档
	Reference string `json:"reference,omitempty"`
}

// WriteResponse 将错误或响应数据写入 http 响应主体.
// 它使用 errors.ParseCoder 将任何错误解析为 errors.Coder
// errors.Coder 包含错误代码、用户安全错误消息 和 http 状态代码.
func WriteResponse(c *gin.Context, err error, data interface{}) {
	if err != nil {
		log.Errorf("%#+v", err)
		coder := errors.ParseCoder(err)
		c.JSON(coder.HTTPStatus(), ErrResponse{
			Code:      coder.Code(),
			Message:   coder.String(),
			Reference: coder.Reference(),
		})

		return
	}

	c.JSON(http.StatusOK, data)
}
