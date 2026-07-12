package entity

import (
	"time"

	"github.com/fairride/shared/errors"
)

// CallSession is an audit-only record of a phone-call action (Part 1 — no
// real telephony/virtual-number integration exists; this exists purely so
// "who called whom, when, for which trip" is answerable later).
type CallSession struct {
	ID        string
	TripID    string
	CallerID  string
	CalleeID  string
	CreatedAt time.Time
}

func NewCallSession(id, tripID, callerID, calleeID string, now time.Time) (*CallSession, error) {
	if id == "" || tripID == "" || callerID == "" || calleeID == "" {
		return nil, errors.InvalidArgument("call session requires id, trip_id, caller_id and callee_id")
	}
	return &CallSession{ID: id, TripID: tripID, CallerID: callerID, CalleeID: calleeID, CreatedAt: now}, nil
}
