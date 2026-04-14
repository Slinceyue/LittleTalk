package enum

type MessageType int8

const (
	Text MessageType = iota + 1
	File MessageType = iota + 1
)
