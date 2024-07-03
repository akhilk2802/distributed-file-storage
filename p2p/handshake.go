package p2p

import "errors"

var ErrInvalidHandShake = errors.New("invalid handshake")

type HandeShakeFunc func(Peer) error

func NOPHandShakeFunc(Peer) error { return nil }
