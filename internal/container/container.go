package container

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui"
	"github.com/PhantomLambSoft/CmdBox/internal/cli/ui/prompts"
	"github.com/PhantomLambSoft/CmdBox/internal/database"
	"github.com/PhantomLambSoft/CmdBox/internal/paths"
	"github.com/PhantomLambSoft/CmdBox/internal/repository"
	"github.com/PhantomLambSoft/CmdBox/internal/resolve"
	"github.com/PhantomLambSoft/CmdBox/internal/runtime"
	"github.com/PhantomLambSoft/CmdBox/internal/services"
	"github.com/PhantomLambSoft/CmdBox/internal/settings"
	"gorm.io/gorm"
)

var GetAppDataDir = sync.OnceValues(paths.GetAppDataDir)

var GetDB = sync.OnceValues(func() (*gorm.DB, error) {
	dir, err := GetAppDataDir()
	if err != nil {
		return nil, fmt.Errorf("resolving app data dir: %w", err)
	}
	dbPath := filepath.Join(dir, "go-cmdbox.db") // TODO: change this to the real database name
	return database.Bootstrap(dbPath)
})

var GetConsole = sync.OnceValues(func() (*ui.Console, error) {
	dir, err := GetAppDataDir()
	if err != nil {
		return nil, err
	}
	settingsService, err := GetSettingsService()
	if err != nil {
		return nil, err
	}
	cfg := settingsService.Get()
	console, unknown, err := ui.Setup(cfg.UI, dir)
	if err != nil {
		return nil, err
	}
	for _, k := range unknown {
		fmt.Fprintf(os.Stderr, "warning: unknown color override key %q\n", k)
	}
	return console, nil
})

var GetPrompter = sync.OnceValues(func() (*prompts.Prompter, error) {
	console, err := GetConsole()
	if err != nil {
		return nil, err
	}
	return prompts.NewPrompter(console.Palette(), console.UseColor()), nil
})

var GetProfileRepo = sync.OnceValues(func() (repository.ProfileRepository, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	return repository.NewProfileRepository(db, nil), nil
})

var GetCommandRepo = sync.OnceValues(func() (repository.CommandRepository, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	profileRepo, err := GetProfileRepo()
	if err != nil {
		return nil, err
	}
	return repository.NewCommandRepository(db, profileRepo, nil), nil
})

var GetVariableRepo = sync.OnceValues(func() (repository.VariableRepository, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	profileRepo, err := GetProfileRepo()
	if err != nil {
		return nil, err
	}
	return repository.NewVariableRepository(db, profileRepo, nil), nil
})

var GetTagRepo = sync.OnceValues(func() (repository.TagRepository, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	return repository.NewTagRepository(db, nil), nil
})

var GetHistoryRepo = sync.OnceValues(func() (repository.HistoryRepository, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	profileRepo, err := GetProfileRepo()
	if err != nil {
		return nil, err
	}
	return repository.NewHistoryRepository(db, profileRepo), nil
})

func GetResolver(strict bool) (*resolve.Resolver, error) {
	commandRepo, err := GetCommandRepo()
	if err != nil {
		return nil, err
	}
	variableRepo, err := GetVariableRepo()
	if err != nil {
		return nil, err
	}
	repoLookup := resolve.NewLookup(commandRepo, variableRepo)
	lookup := resolve.NewMemoizedLookup(repoLookup)
	return resolve.NewResolver(lookup, strict, 25), nil
}

var GetRunService = sync.OnceValues(func() (services.RunService, error) {
	commandRepo, err := GetCommandRepo()
	if err != nil {
		return nil, err
	}
	resolver, err := GetResolver(false)
	if err != nil {
		return nil, err
	}
	profileRepo, err := GetProfileRepo()
	if err != nil {
		return nil, err
	}
	historyRepo, err := GetHistoryRepo()
	if err != nil {
		return nil, err
	}
	settingsService, err := GetSettingsService()
	if err != nil {
		return nil, err
	}
	executor := runtime.NewExecutor()
	config := services.RunServiceConfig{
		CmdRepo:          commandRepo,
		Resolver:         resolver,
		Executor:         *executor,
		ProfileRepo:      profileRepo,
		HistoryRepo:      historyRepo,
		SettingsProvider: settingsService,
	}
	return services.NewRunService(config), nil
})

var GetSettingsService = sync.OnceValues(func() (*settings.Service, error) {
	dir, err := GetAppDataDir()
	if err != nil {
		return nil, err
	}
	profileService, err := GetProfileService()
	if err != nil {
		return nil, err
	}
	settingsPath, err := settings.ResolveSettingsPath(dir, nil, profileService)
	if err != nil {
		return nil, err
	}
	repo := settings.NewSettingsRepository(settingsPath)
	return settings.NewService(repo, nil, profileService)
})

var GetProfileService = sync.OnceValues(func() (services.ProfileService, error) {
	profileRepo, err := GetProfileRepo()
	if err != nil {
		return nil, err
	}
	return services.NewProfileService(profileRepo), nil
})

var GetCommandService = sync.OnceValues(func() (services.CommandService, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	commandRepo, err := GetCommandRepo()
	if err != nil {
		return nil, err
	}
	tagRepo, err := GetTagRepo()
	if err != nil {
		return nil, err
	}
	profileRepo, err := GetProfileRepo()
	if err != nil {
		return nil, err
	}
	return services.NewCommandService(db, commandRepo, tagRepo, profileRepo), nil
})

var GetVariableService = sync.OnceValues(func() (services.VariableService, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}
	variableRepo, err := GetVariableRepo()
	if err != nil {
		return nil, err
	}
	tagRepo, err := GetTagRepo()
	if err != nil {
		return nil, err
	}
	profileRepo, err := GetProfileRepo()
	if err != nil {
		return nil, err
	}
	return services.NewVariableService(db, variableRepo, tagRepo, profileRepo), nil
})

var GetTagService = sync.OnceValues(func() (services.TagService, error) {
	tagRepo, err := GetTagRepo()
	if err != nil {
		return nil, err
	}
	return services.NewTagService(tagRepo), nil
})

var GetHistoryService = sync.OnceValues(func() (services.HistoryService, error) {
	historyRepo, err := GetHistoryRepo()
	if err != nil {
		return nil, err
	}
	profileRepo, err := GetProfileRepo()
	if err != nil {
		return nil, err
	}
	return services.NewHistoryService(historyRepo, profileRepo), nil
})

var GetExportService = sync.OnceValues(func() (*services.ExportService, error) {
	commandService, err := GetCommandService()
	if err != nil {
		return nil, err
	}
	variableService, err := GetVariableService()
	if err != nil {
		return nil, err
	}
	profileRepo, err := GetProfileRepo()
	if err != nil {
		return nil, err
	}
	return services.NewExportService(commandService, variableService, profileRepo), nil
})

var GetImportService = sync.OnceValues(func() (*services.ImportService, error) {
	commandService, err := GetCommandService()
	if err != nil {
		return nil, err
	}
	variableService, err := GetVariableService()
	if err != nil {
		return nil, err
	}
	tagService, err := GetTagService()
	if err != nil {
		return nil, err
	}
	profileRepo, err := GetProfileRepo()
	if err != nil {
		return nil, err
	}
	return services.NewImportService(commandService, variableService, tagService, profileRepo), nil
})
