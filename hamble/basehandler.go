package hamble

import (
	"github.com/dawnzzz/hamble-tcp-server/iface"
)

type BaseHandler struct {
}

func (handler *BaseHandler) PreHandle(_ iface.IRequest) {
}

func (handler *BaseHandler) Handle(_ iface.IRequest) {
}

func (handler *BaseHandler) PostHandle(_ iface.IRequest) {
}
