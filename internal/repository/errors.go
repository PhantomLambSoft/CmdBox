package repository

import "errors"

var (
	// Profile errors
	ErrProfileNotFound      = errors.New("profile not found")
	ErrProfileNameExists    = errors.New("a profile with that name already exists")
	ErrDefaultProfileName   = errors.New("the default profile cannot be renamed")
	ErrDefaultProfileDelete = errors.New("the default profile cannot be deleted")
	ErrProfileInUse         = errors.New("profile is currently active and cannot be deleted. Switch to another profile first")
	ErrProfileHasContent    = errors.New("profile still has commands or variables assigned to it")

	// Command errors
	ErrAliasConflict          = errors.New("a command with that alias already exists")
	ErrUnknownAlias           = errors.New("no command found with that alias")
	ErrUnknownCommand         = errors.New("no command found with that id")
	ErrTagAttachFailed        = errors.New("could not attach tags to command")
	ErrTagDetachFailed        = errors.New("could not detach tags from command")
	ErrNoUpdateTarget         = errors.New("no command provided for update")
	ErrConflictingClearAndSet = errors.New("cannot both set and clear the same field")
)
