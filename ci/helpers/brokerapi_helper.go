package helpers

import (
	"github.com/frodenas/brokerapi"
)

type ByServiceID []brokerapi.Service

func (a ByServiceID) Len() int           { return len(a) }
func (a ByServiceID) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByServiceID) Less(i, j int) bool { return a[i].ID < a[j].ID }
