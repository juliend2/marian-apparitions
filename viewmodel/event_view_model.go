package viewmodel

import (
	//"fmt"
	"strings"

	"marianapparitions/model"
)

type EventViewModel struct {
	model.Event
}
func (e *EventViewModel) GetName() string     { return e.Name }
func (e *EventViewModel) GetCategory() string { return e.Category }
func (e *EventViewModel) GetYears() string    { return e.Years }


func NewEventVM(event *model.Event) *EventViewModel {
	return &EventViewModel{*event}
}

// Returns whether the church ("Catholic Church" and so on) approved the event
func (vm *EventViewModel) IsChurchApproved(churchNameSubstr string) bool {
	for _, block := range vm.Event.Blocks {
		if strings.Contains(block.ChurchAuthority, churchNameSubstr) && block.AuthorityPosition == "approved" {
			return true
		}
	}
	return false
}

func (vm *EventViewModel) GetApproverChurch(churchNameSubstr string) string {
	for _, block := range vm.Event.Blocks {
		if strings.Contains(block.ChurchAuthority, churchNameSubstr) && block.AuthorityPosition == "approved" {
			return block.ChurchAuthority
		}
	}
	return ""
}
