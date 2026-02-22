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

func (vm *EventViewModel) HasAnyApproval() bool {
	return vm.GetApproverChurch("Catholic") != "" || vm.GetApproverChurch("Orthodox") != "" || vm.GetApproverChurch("Anglican") != ""
}

func (vm *EventViewModel) GetApproverChurch(churchNameSubstr string) string {
	for _, block := range vm.Event.Blocks {
		if strings.Contains(block.ChurchAuthority, churchNameSubstr) && block.AuthorityPosition == "approved" {
			return block.ChurchAuthority
		}
	}
	return ""
}
