package dtos

//WebSocketMessage represents a message ready to be marshalled into JSON
type WebSocketMessage struct {
}

//WebSocketMessageMarshaller is used to abstract away the details of marshalling
//ws messages into byte slices
type WebSocketMessageMarshaller interface {
	Marshall(*WebSocketMessage) ([]byte, error)
	Unmarshall([]byte) (*WebSocketMessage, error)
}
