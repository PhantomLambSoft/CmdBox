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
	ErrAliasConflict  = errors.New("a command with that alias already exists")
	ErrUnknownAlias   = errors.New("no command found with that alias")
	ErrUnknownCommand = errors.New("no command found with that id")

	// Variable errors
	ErrNameConflict    = errors.New("a variable with that name already exists")
	ErrUnKnownName     = errors.New("no variable found with that name")
	ErrUnknownVariable = errors.New("no variable found with that id")

	ErrTagNameConflict = errors.New("a tag with that name already exists")
	ErrUnknownTagName  = errors.New("no tag found with that name")
	ErrUnknownTag      = errors.New("no tag found with that id")

	ErrUnknownCommandHistory = errors.New("no command history found with that id")

	ErrTagAttachFailed        = errors.New("could not attach tags")
	ErrTagDetachFailed        = errors.New("could not detach tags")
	ErrNoUpdateTarget         = errors.New("nothing provided for update")
	ErrConflictingClearAndSet = errors.New("cannot both set and clear the same field")
)
