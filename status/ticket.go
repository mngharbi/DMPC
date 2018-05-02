package status

import (
	"github.com/mngharbi/DMPC/core"
)

/*
	Ticket definition
*/
type Ticket string

/*
	Function to generate a new ticket
*/
type TicketGenerator func() Ticket

/*
	Generates a new ticket through the core package utility
*/
func RequestNewTicket() Ticket {
	return Ticket(core.GenerateUniqueId())
}
