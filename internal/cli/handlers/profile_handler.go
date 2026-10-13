package handlers

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui/presenters"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/models"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
)

type AddProfileArgs struct {
	Name           string
	Description    string
	HasDescription bool
	Interactive    bool
}

type UpdateProfileArgs struct {
	CurrentName    string
	NewName        string
	Description    string
	HasDescription bool
	SetPairs       string
	EditMode       bool
	EditFields     string
}

type ListProfileArgs struct {
	ListParams
}

type DeleteProfileArgs struct {
	Name  string
	Force bool
}

type SwitchProfileArgs struct {
	Name     string
	Cmd      bool
	Var      bool
	Settings bool
}

func RunAddProfile(args AddProfileArgs) error {
	profileService, err := container.GetProfileService()
	if err != nil {
		return fmt.Errorf("getting profile service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	var description *string
	if args.HasDescription {
		description = &args.Description
	}

	profile, err := profileService.CreateProfile(args.Name, description)
	if err != nil {
		return fmt.Errorf("creating profile: %w", err)
	}
	console.Print(presenters.RenderProfileCreated(*profile, console.Palette()))
	return nil
}

func RunGetProfile(name string) error {
	profileService, err := container.GetProfileService()
	if err != nil {
		return fmt.Errorf("getting profile service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	profile, err := profileService.GetProfile(name)
	if err != nil {
		return fmt.Errorf("getting profile: %w", err)
	}
	console.Print(presenters.RenderProfile(*profile, console.Palette()))
	return nil
}

func RunUpdateProfile(args UpdateProfileArgs) error {
	profileService, err := container.GetProfileService()
	if err != nil {
		return fmt.Errorf("getting profile service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	input := services.UpdateProfileConfig{}

	if args.NewName != "" {
		input.NewName = &args.NewName
	}
	if args.HasDescription {
		input.Description = &args.Description
	}

	profile, err := profileService.UpdateProfile(args.CurrentName, input)
	if err != nil {
		return fmt.Errorf("updating profile: %w", err)
	}
	console.Print(presenters.RenderProfileUpdated(*profile, console.Palette()))
	return nil
}

func RunListProfile(args ListProfileArgs) error {
	profileService, err := container.GetProfileService()
	if err != nil {
		return fmt.Errorf("getting profile service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}
	var limit *int
	if args.HasLimit {
		limit = &args.Limit
	}

	profiles, err := profileService.ListProfiles(args.Order, limit)
	if err != nil {
		return fmt.Errorf("listing profiles: %w", err)
	}

	rendered := presenters.RenderProfileList(profiles, "Profiles", args.Fields, console.Palette())
	rowCount := len(profiles)
	return console.PrintPaged(rendered, &rowCount, args.Page)
}

func RunDeleteProfile(args DeleteProfileArgs) error {
	profileService, err := container.GetProfileService()
	if err != nil {
		return fmt.Errorf("getting profile service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	profile, err := profileService.GetProfile(args.Name)
	if err != nil {
		return fmt.Errorf("getting profile: %w", err)
	}

	result, err := profileService.DeleteProfile(args.Name, args.Force)
	if err != nil {
		return fmt.Errorf("deleting profile: %w", err)
	}
	console.Print(presenters.RenderProfileDeleted(*profile, result, console.Palette()))
	return nil
}

func RunProfileStatus() error {
	profileService, err := container.GetProfileService()
	if err != nil {
		return fmt.Errorf("getting profile service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	status, err := profileService.GetStatus()
	if err != nil {
		return fmt.Errorf("getting profile status: %w", err)
	}
	console.Print(presenters.RenderProfileStatus(status, console.Palette()))
	return nil
}

func RunSwitchProfile(args SwitchProfileArgs) error {
	profileService, err := container.GetProfileService()
	if err != nil {
		return fmt.Errorf("getting profile service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	var state *models.ProfileState
	var switchedKinds []string

	switchOne := func(fn func(string) (*models.ProfileState, error), label string) error {
		s, err := fn(args.Name)
		if err != nil {
			return fmt.Errorf("switching %s profiles: %w", label, err)
		}
		state = s
		switchedKinds = append(switchedKinds, label)
		return nil
	}

	if !args.Cmd && !args.Var && !args.Settings {
		if err := switchOne(profileService.SwitchProfile, "command"); err != nil {
			return err
		}
		switchedKinds = []string{"command", "variable", "settings"}
	} else {
		if args.Cmd {
			if err := switchOne(profileService.SwitchCommandProfile, "command"); err != nil {
				return err
			}
		}
		if args.Var {
			if err := switchOne(profileService.SwitchVariableProfile, "variable"); err != nil {
				return err
			}
		}
		if args.Settings {
			if err := switchOne(profileService.SwitchSettingsProfile, "settings"); err != nil {
				return err
			}
		}
	}

	console.Print(presenters.RenderProfileSwitched(state, switchedKinds, args.Name, console.Palette()))
	return nil
}
