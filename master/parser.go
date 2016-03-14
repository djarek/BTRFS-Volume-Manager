package main

import "github.com/djarek/btrfs-volume-manager/common/dtos"

/*messageParser parses the received WebSocketMessage and dispatches appropriate
handler functions. Implements the RecvMessageParser interface.
*/
type messageParser struct{}

/*ParseRecvMsg parses the received WebSocketMessage and dispatches appropriate
handler functions. */
func (mp messageParser) ParseRecvMsg(dtos.WebSocketMessage) error {
	return nil
}
