from functools import lru_cache

from cmdbox.cli.ui.console import ConsoleUI
from cmdbox.cli.ui.theme_builder import build_theme
from cmdbox.core.fields import (
    COMMAND_DISPLAY_FIELDS,
    COMMAND_SEARCH_FIELDS,
    VARIABLE_DISPLAY_FIELDS,
    VARIABLE_SEARCH_FIELDS,
    TAG_DISPLAY_FIELDS,
    TAG_SEARCH_FIELDS,
)
from cmdbox.repositories.history_repository import HistoryRepository
from cmdbox.repositories.profile_repository import ProfileRepository
from cmdbox.services.export_service import ExportService
from cmdbox.services.field_selection import FieldSelectionResolver
from cmdbox.services.history_service import HistoryService
from cmdbox.services.import_service import ImportService
from cmdbox.services.variable_services import VariableServices
from cmdbox.settings.models import Settings
from cmdbox.settings.settings_repository import SettingsRepository
from cmdbox.settings.settings_service import SettingsService
from cmdbox.core.paths import get_app_data_dir
from cmdbox.database import get_db
from cmdbox.repositories.command_repository import CommandRepository
from cmdbox.repositories.tag_repository import TagRepository
from cmdbox.repositories.variable_repository import VariableRepository
from cmdbox.resolve.lookup import MemoizedLookup, RepoLookup
from cmdbox.resolve.resolver import Resolver
from cmdbox.runtime.executor import Executor
from cmdbox.services.command_services import CommandServices
from cmdbox.services.run_service import RunService
from cmdbox.services.tag_services import TagServices
from cmdbox.services.profile_services import ProfileServices


@lru_cache(maxsize=1)
def get_settings_service() -> SettingsService:
    profile_service = get_profile_service()
    config_path = profile_service.resolve_settings_path(get_app_data_dir())
    repo = SettingsRepository(config_path)
    settings = SettingsService(repo)
    return settings


def get_settings() -> Settings:
    return get_settings_service().get()


@lru_cache(maxsize=1)
def get_console() -> ConsoleUI:
    _settings = get_settings()
    theme = build_theme(_settings)
    return ConsoleUI(
        theme,
        use_color=_settings.ui.use_color,
        pager_mode=_settings.ui.pager_mode,
        pager_min_rows=_settings.ui.pager_min_rows,
        pager_page_step=_settings.ui.pager_page_step,
        pager_line_step=_settings.ui.pager_line_step,
    )


@lru_cache(maxsize=1)
def get_command_repo() -> CommandRepository:
    profile_repo = get_profile_repo()
    return CommandRepository(profile_repo)


@lru_cache(maxsize=1)
def get_variable_repo() -> VariableRepository:
    profile_repo = get_profile_repo()
    return VariableRepository(profile_repo)


@lru_cache(maxsize=1)
def get_tag_repo() -> TagRepository:
    return TagRepository()


@lru_cache(maxsize=1)
def get_history_repo() -> HistoryRepository:
    get_db()
    profile_repo = get_profile_repo()
    return HistoryRepository(profile_repo)


@lru_cache(maxsize=1)
def get_profile_repo() -> ProfileRepository:
    return ProfileRepository()


@lru_cache(maxsize=1)
def get_resolver(strict: bool = False) -> Resolver:
    get_db()
    command_repo = get_command_repo()
    variable_repo = get_variable_repo()
    repo_lookup = RepoLookup(command_repo, variable_repo)
    lookup = MemoizedLookup(repo_lookup)
    return Resolver(lookup, strict=strict)


@lru_cache(maxsize=1)
def get_run_service() -> RunService:
    get_db()
    cmd_repo = get_command_repo()
    resolver = get_resolver()
    executor = Executor()
    return RunService(
        cmd_repo,
        resolver,
        executor,
        history_repo=get_history_repo(),
        get_settings=get_settings,
    )


@lru_cache(maxsize=1)
def get_command_services() -> CommandServices:
    get_db()
    cmd_repo = get_command_repo()
    tag_repo = get_tag_repo()
    return CommandServices(command_repository=cmd_repo, tag_repository=tag_repo)


@lru_cache(maxsize=1)
def get_variable_services() -> VariableServices:
    get_db()
    var_repo = get_variable_repo()
    tag_repo = get_tag_repo()
    return VariableServices(variable_repository=var_repo, tag_repository=tag_repo)


@lru_cache(maxsize=1)
def get_tag_services() -> TagServices:
    get_db()
    tag_repo = get_tag_repo()
    return TagServices(tag_repository=tag_repo)


@lru_cache(maxsize=1)
def get_history_service() -> HistoryService:
    return HistoryService(
        repo=get_history_repo(),
        get_settings=get_settings,
    )


@lru_cache(maxsize=1)
def get_profile_service() -> ProfileServices:
    return ProfileServices(profile_repository=get_profile_repo())


@lru_cache(maxsize=1)
def get_export_service() -> ExportService:
    return ExportService(
        cmd_service=get_command_services(),
        var_service=get_variable_services(),
    )


@lru_cache(maxsize=1)
def get_import_service() -> ImportService:
    return ImportService(
        cmd_service=get_command_services(),
        var_service=get_variable_services(),
        tag_service=get_tag_services(),
    )


@lru_cache(maxsize=1)
def get_command_display_field_resolver() -> FieldSelectionResolver:
    return FieldSelectionResolver(allowed_fields=COMMAND_DISPLAY_FIELDS)


@lru_cache(maxsize=1)
def get_command_search_field_resolver() -> FieldSelectionResolver:
    return FieldSelectionResolver(allowed_fields=COMMAND_SEARCH_FIELDS)


@lru_cache(maxsize=1)
def get_variable_display_field_resolver() -> FieldSelectionResolver:
    return FieldSelectionResolver(allowed_fields=VARIABLE_DISPLAY_FIELDS)


@lru_cache(maxsize=1)
def get_variable_search_field_resolver() -> FieldSelectionResolver:
    return FieldSelectionResolver(allowed_fields=VARIABLE_SEARCH_FIELDS)


@lru_cache(maxsize=1)
def get_tag_display_field_resolver() -> FieldSelectionResolver:
    return FieldSelectionResolver(allowed_fields=TAG_DISPLAY_FIELDS)


@lru_cache(maxsize=1)
def get_tag_search_field_resolver() -> FieldSelectionResolver:
    return FieldSelectionResolver(allowed_fields=TAG_SEARCH_FIELDS)
