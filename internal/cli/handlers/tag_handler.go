package handlers

import (
	"fmt"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui/presenters"
	"github.com/PhantomLambSoft/CmdBox/internal/container"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
)

type AddTagArgs struct {
	Name           string
	Description    string
	HasDescription bool
	Interactive    bool
}

type UpdateTagArgs struct {
	CurrentName    string
	NewName        string
	Description    string
	HasDescription bool
	SetPairs       string
	EditMode       bool
	EditFields     string
}

type ListTagArgs struct {
	ListParams
}

type SearchTagArgs struct {
	SearchParams
}

type DeleteTagArgs struct {
	Name string
}

func RunAddTag(args AddTagArgs) error {
	tagService, err := container.GetTagService()
	if err != nil {
		return fmt.Errorf("getting tag service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	var description *string
	if args.HasDescription {
		description = &args.Description
	}
	tag, err := tagService.CreateTag(args.Name, description)
	if err != nil {
		return fmt.Errorf("creating tag: %w", err)
	}

	console.Print(presenters.RenderTagCreated(*tag, console.Palette()))
	return nil
}

func RunGetTag(name string) error {
	tagService, err := container.GetTagService()
	if err != nil {
		return fmt.Errorf("getting tag service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	tag, err := tagService.GetTag(name)
	if err != nil {
		return fmt.Errorf("getting tag: %w", err)
	}

	console.Print(presenters.RenderTag(*tag, console.Palette()))
	return nil
}

func RunUpdateTag(args UpdateTagArgs) error {
	tagService, err := container.GetTagService()
	if err != nil {
		return fmt.Errorf("getting tag service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	input := services.UpdateTagConfig{}
	if args.NewName != "" {
		input.NewName = &args.NewName
	}
	if args.Description != "" {
		input.Description = &args.Description
	}

	tag, err := tagService.UpdateTag(args.CurrentName, input)
	if err != nil {
		return fmt.Errorf("updating tag: %w", err)
	}

	console.Print(presenters.RenderTagUpdated(*tag, console.Palette()))
	return nil
}

func RunListTag(args ListTagArgs) error {
	tagService, err := container.GetTagService()
	if err != nil {
		return fmt.Errorf("getting tag service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	var limit *int
	if args.HasLimit {
		limit = &args.Limit
	}

	tags, err := tagService.ListTags(args.Order, limit)
	if err != nil {
		return fmt.Errorf("listing tags: %w", err)
	}

	rendered := presenters.RenderTagList(tags, "Tags", args.Fields, console.Palette())
	rowCount := len(tags)
	return console.PrintPaged(rendered, &rowCount, args.Page)
}

func RunSearchTag(args SearchTagArgs) error {
	tagService, err := container.GetTagService()
	if err != nil {
		return fmt.Errorf("getting tag service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}
	var limit *int
	if args.HasLimit {
		limit = &args.Limit
	}

	tags, err := tagService.Search(args.Term, args.SearchIn, limit)
	if err != nil {
		return fmt.Errorf("searching tags: %w", err)
	}

	rendered := presenters.RenderTagList(tags, "Tag search results", args.Fields, console.Palette())
	rowCount := len(tags)
	return console.PrintPaged(rendered, &rowCount, args.Page)
}

func RunDeleteTag(args DeleteTagArgs) error {
	tagService, err := container.GetTagService()
	if err != nil {
		return fmt.Errorf("getting tag service: %w", err)
	}
	console, err := container.GetConsole()
	if err != nil {
		return fmt.Errorf("getting console: %w", err)
	}

	tag, err := tagService.GetTag(args.Name)
	if err != nil {
		return fmt.Errorf("getting tag: %w", err)
	}
	if err := tagService.DeleteTag(args.Name); err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}
	console.Print(presenters.RenderTagDeleted(*tag, console.Palette()))
	return nil
}
