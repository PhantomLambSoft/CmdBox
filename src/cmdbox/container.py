from functools import lru_cache

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


@lru_cache(maxsize=1)
def get_command_repo() -> CommandRepository:
    return CommandRepository()


@lru_cache(maxsize=1)
def get_variable_repo() -> VariableRepository:
    return VariableRepository()


@lru_cache(maxsize=1)
def get_tag_repo() -> TagRepository:
    return TagRepository()


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
    return RunService(cmd_repo, resolver, executor)


@lru_cache(maxsize=1)
def get_command_services() -> CommandServices:
    get_db()
    cmd_repo = get_command_repo()
    tag_repo = get_tag_repo()
    return CommandServices(command_repository=cmd_repo, tag_repository=tag_repo)


@lru_cache(maxsize=1)
def get_tag_services() -> TagServices:
    get_db()
    tag_repo = get_tag_repo()
    return TagServices(tag_repository=tag_repo)
