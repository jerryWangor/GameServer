package tcp

func Handler() Handler {
	var hander Handler
	hander = &ServeHandler{}
	hander.Handle(context.Context())
	return hander
}
