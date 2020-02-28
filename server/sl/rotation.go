// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package sl

import (
	"fmt"

	"github.com/mattermost/mattermost-plugin-solar-lottery/server/utils/kvstore"
	"github.com/mattermost/mattermost-plugin-solar-lottery/server/utils/types"
)

type Rotation struct {
	PluginVersion string
	RotationID    types.ID
	AutofillType  string
	IsArchived    bool
	IssueSources  *types.Index // of *IssueSource
	// ShiftSource *ShiftSource
	Pending           []*Task
	InProgress        []*Task
	MattermostUserIDs *types.IDIndex `json:",omitempty"`

	users UserMap
}

func NewRotation() *Rotation {
	r := &Rotation{}
	r.init()
	return r
}

func (r *Rotation) init() {
	if r.MattermostUserIDs == nil {
		r.MattermostUserIDs = types.NewIDIndex()
	}
	if r.users == nil {
		r.users = UserMap{}
	}
}

func (rotation *Rotation) WithMattermostUserIDs(pool UserMap) *Rotation {
	newRotation := *rotation
	newRotation.MattermostUserIDs = types.NewIDIndex()
	for id := range pool {
		newRotation.MattermostUserIDs.Set(id)
	}
	if pool == nil {
		pool = UserMap{}
	}
	newRotation.users = pool
	return &newRotation
}

func (r *Rotation) String() string {
	return r.Name()
}

func (r *Rotation) Name() string {
	return kvstore.NameFromID(r.RotationID)
}

func (r *Rotation) Markdown() string {
	return r.Name()
}

func (r *Rotation) MarkdownBullets() string {
	out := fmt.Sprintf("- **%s**\n", r.Name())
	out += fmt.Sprintf("  - ID: `%s`.\n", r.RotationID)
	out += fmt.Sprintf("  - Users (%v): %s.\n", r.MattermostUserIDs.Len(), r.users.MarkdownWithSkills())

	// if rotation.Autopilot.On {
	// 	out += fmt.Sprintf("  - Autopilot: **on**\n")
	// 	out += fmt.Sprintf("    - Auto-start: **%v**\n", rotation.Autopilot.StartFinish)
	// 	out += fmt.Sprintf("    - Auto-fill: **%v**, %v days prior to start\n", rotation.Autopilot.Fill, rotation.Autopilot.FillPrior)
	// 	out += fmt.Sprintf("    - Notify users in advance: **%v**, %v days prior to transition\n", rotation.Autopilot.Notify, rotation.Autopilot.NotifyPrior)
	// } else {
	// 	out += fmt.Sprintf("  - Autopilot: **off**\n")
	// }

	return out
}

func (r *Rotation) MapUsers(ids *types.IDIndex) UserMap {
	users := UserMap{}
	for _, id := range r.MattermostUserIDs.IDs() {
		users[id] = r.users[id]
	}
	return users
}

func (r *Rotation) IssueSource(sourceName types.ID) (*IssueSource, int) {
	for i, id := range r.IssueSources.IDs() {
		if id == sourceName {
			return r.IssueSources.Get(id).(*IssueSource), i
		}
	}
	return nil, -1
}
