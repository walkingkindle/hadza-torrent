package parser

import "errors"

var ErrInvalidMagnetLink = errors.New("invalid magnet link")

var ErrMissingExactTopic = errors.New("magnet link is missing required exact topic (xt)")
