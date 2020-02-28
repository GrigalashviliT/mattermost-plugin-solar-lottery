// Copyright (c) 2019-present Mattermost, Inc. All Rights Reserved.
// See License for license information.

package sl

import (
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-plugin-solar-lottery/server/utils/bot"
	"github.com/mattermost/mattermost-plugin-solar-lottery/server/utils/kvstore"
	"github.com/mattermost/mattermost-plugin-solar-lottery/server/utils/types"
)

type filterf func(*sl) error

func (sl *sl) Setup(filters ...filterf) error {
	for _, filter := range filters {
		err := filter(sl)
		if err != nil {
			return err
		}
	}
	return nil
}

func withRotation(rotationID string) func(sl *sl) error {
	return func(sl *sl) error {
		return nil
	}
}

func withRotationExpanded(rotation *Rotation) func(sl *sl) error {
	return func(sl *sl) error {
		return sl.ExpandRotation(rotation)
	}
}

func withRotationIsNotArchived(rotation *Rotation) func(sl *sl) error {
	return func(sl *sl) error {
		if rotation.IsArchived {
			return errors.Errorf("rotation %s is archived", rotation.Markdown())
		}
		return nil
	}
}

func withUser(user *User) func(sl *sl) error {
	return func(sl *sl) error {
		if user.loaded {
			return nil
		}
		loadedUser, _, err := sl.loadOrMakeUser(user.MattermostUserID)
		if err != nil {
			return err
		}
		*user = *loadedUser
		return nil
	}
}

func withActingUser(sl *sl) error {
	if sl.actingUser != nil {
		return nil
	}
	user, _, err := sl.loadOrMakeUser(sl.actingMattermostUserID)
	if err != nil {
		return err
	}
	sl.actingUser = user
	return nil
}

func withActingUserExpanded(sl *sl) error {
	if sl.actingUser != nil && sl.actingUser.mattermostUser != nil {
		return nil
	}
	err := withActingUser(sl)
	if err != nil {
		return err
	}
	return sl.ExpandUser(sl.actingUser)
}

func withUserExpanded(user *User) func(sl *sl) error {
	return func(sl *sl) error {
		if user != nil && user.mattermostUser != nil {
			return nil
		}
		err := withUser(user)(sl)
		if err != nil {
			return err
		}
		return sl.ExpandUser(sl.actingUser)
	}
}

func withKnownSkills(sl *sl) error {
	if sl.knownSkills != nil {
		return nil
	}

	skills, err := sl.Store.IDIndex(KeyKnownSkills).Load()
	if err == kvstore.ErrNotFound {
		sl.knownSkills = types.NewIDIndex()
		return nil
	}
	if err != nil {
		return err
	}
	sl.knownSkills = skills
	return nil
}

func withValidSkillName(skillName types.ID) func(sl *sl) error {
	return func(sl *sl) error {
		err := sl.Setup(withKnownSkills)
		if err != nil {
			return err
		}
		if !sl.knownSkills.Contains(skillName) {
			return errors.Errorf("skill %s is not found", skillName)
		}
		return nil
	}
}

func withActiveRotations(sl *sl) error {
	if sl.activeRotations != nil && sl.activeRotations.Len() > 0 {
		return nil
	}

	rotations, err := sl.Store.IDIndex(KeyActiveRotations).Load()
	if err == kvstore.ErrNotFound {
		sl.activeRotations = types.NewIDIndex()
		return nil
	}
	if err != nil {
		return err
	}
	sl.activeRotations = rotations
	return nil
}

func pushLogger(apiName string, logContext bot.LogContext) func(*sl) error {
	return func(sl *sl) error {
		err := withActingUserExpanded(sl)
		if err != nil {
			return err
		}

		logger := sl.Logger
		logger = logger.With(logContext)
		logger = logger.With(bot.LogContext{
			ctxActingUsername: sl.actingUser.MattermostUsername(),
			ctxAPI:            apiName,
		})

		if sl.loggers == nil {
			sl.loggers = []bot.Logger{logger}
		} else {
			sl.loggers = append(sl.loggers, logger)
		}
		sl.Logger = logger
		return nil
	}
}

func (sl *sl) popLogger() {
	l := len(sl.loggers)
	if l == 0 {
		return
	}
	sl.Logger = sl.loggers[l-1]
	sl.loggers = sl.loggers[:l-1]
}
