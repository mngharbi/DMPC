/*
	Utilities for ids
*/

package core

import (
	"github.com/rs/xid"
)

func GenerateUniqueId() string {
	return xid.New().String()
}
