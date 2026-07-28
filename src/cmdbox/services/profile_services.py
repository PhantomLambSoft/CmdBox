from dataclasses import dataclass
from pathlib import Path
from typing import Sequence

from cmdbox.models import Profile, ProfileState
from cmdbox.repositories.profile_repository import ProfileRepository


@dataclass
class ProfileStatus:
    """
    Represents the status of different profiles and their linkage.

    This class is designed to hold and compare the status of command, variable,
    and settings profiles. It determines whether all the profiles are in sync
    or linked, providing a way to check consistency across these profiles.

    Attributes:
        command_profile (str): The command profile status.
        variable_profile (str): The variable profile status.
        settings_profile (str): The settings profile status.
    """

    command_profile: str
    variable_profile: str
    settings_profile: str

    @property
    def linked(self) -> bool:
        return self.command_profile == self.variable_profile == self.settings_profile


class ProfileServices:
    """
    Provides services for managing and interacting with user profiles.

    This class acts as a higher-level abstraction for handling various operations on
    profiles, such as creating, updating, deleting, and switching between them. It
    utilizes a repository object to interact with the underlying data storage or model.
    It is designed to manage profiles with attributes like command, settings, and
    variable profiles and provides mechanisms to modify their states or retrieve their
    status.

    Attributes:
        profile_repository (ProfileRepository): Repository object used for interacting
            with the profile storage.
    """

    def __init__(self, profile_repository: ProfileRepository):
        self._repo = profile_repository

    def create_profile(self, name: str, description: str | None = None) -> Profile:
        """
        Creates a new profile with the given name and an optional description.

        This method interacts with the repository to create and persist a
        profile object.

        Args:
            name (str): The name of the profile to be created.
            description (str | None): An optional description for the profile.

        Returns:
            Profile: The newly created profile object.
        """
        return self._repo.create(name, description)

    def update_profile(self, name: str, **fields) -> Profile:
        """
        Updates the profile of a user with the given name using the specified fields.

        Retrieves the user profile by name and updates it with the provided
        keyword arguments. Returns the updated profile.

        Args:
            name (str): The name of the user whose profile is to be updated.
            **fields: Arbitrary keyword arguments representing field updates
                for the user profile.

        Returns:
            Profile: The updated profile object.
        """
        profile = self._repo.get_by_name(name)
        return self._repo.update(profile, **fields)

    def delete_profile(self, name: str, force: bool = False) -> bool:
        """
        Deletes a user profile from the repository.

        This method removes a profile identified by the given name. If the `force`
        flag is set to True, the profile will be deleted forcibly, regardless
        of any restrictions.

        Args:
            name (str): The name of the profile to delete.
            force (bool): Whether to forcefully delete the profile. Defaults to False.

        Returns:
            bool: True if the profile was successfully deleted, False otherwise.
        """
        profile = self._repo.get_by_name(name)
        return self._repo.delete(profile, force=force)

    def get_profile(self, name: str) -> Profile:
        """
        Fetches and returns a user profile by name.

        Args:
            name (str): The name of the profile to retrieve.

        Returns:
            Profile: The profile object that corresponds to the given name.
        """
        return self._repo.get_by_name(name)

    def list_profiles(
        self, order_by: str | Sequence[str] = "name", limit: int = 25
    ) -> list[Profile]:
        """
        Lists all profiles from the repository with specified ordering and limit.

        This function retrieves a list of profiles from the repository. The ordering
        of the list can be customized by specifying one or more fields to sort by,
        and the number of profiles returned can be limited to a specific count.

        Args:
            order_by (str): A string or a sequence of strings indicating the fields by
                which the profiles should be ordered. Defaults to "name".
            limit (int): An integer specifying the maximum number of profiles to retrieve.
                Defaults to 25.

        Returns:
            A list of Profile objects retrieved from the repository.
        """
        return self._repo.list_all(order_by=order_by, limit=limit)

    def switch_profile(self, name: str) -> ProfileState:
        """
        Switches the current active profile to the specified profile by name.

        This method retrieves the profile by its name and updates the active
        command profile, variable profile, and settings profile in the repository.

        Args:
            name (str): The name of the profile to be activated.

        Returns:
            ProfileState: The state of the newly activated profile.
        """
        profile = self._repo.get_by_name(name)
        self._repo.set_active_command_profile(profile)
        self._repo.set_active_variable_profile(profile)
        return self._repo.set_active_settings_profile(profile)

    def switch_command_profile(self, name: str) -> ProfileState:
        profile = self._repo.get_by_name(name)
        return self._repo.set_active_command_profile(profile)

    def switch_variable_profile(self, name: str) -> ProfileState:
        profile = self._repo.get_by_name(name)
        return self._repo.set_active_variable_profile(profile)

    def switch_settings_profile(self, name: str) -> ProfileState:
        profile = self._repo.get_by_name(name)
        return self._repo.set_active_settings_profile(profile)

    def get_status(self) -> ProfileState:
        """
        Retrieves the current state of the profile.

        This method gathers active command, variable, and settings profiles from the
        repository's state and constructs a `ProfileState` instance with this information.

        Returns:
            ProfileState: An object encapsulating the current command, variable,
            and settings profiles.
        """
        state = self._repo.get_state()
        return ProfileState(
            command_profile=state.active_command_profile,
            variable_profile=state.active_variable_profile,
            settings_profile=state.active_settings_profile,
        )

    def resolve_settings_path(
        self, app_data_dir: Path, name: str | None = None
    ) -> Path:
        """
        Resolves the file path for the settings configuration file based on the
        provided application data directory and optional profile name.

        This method determines the appropriate settings file for a specific
        profile. If no profile name is provided, the active settings profile from
        the repository state will be used. For the default profile, the method will
        return the path to "config.toml". For all other profiles, it will construct
        a filename with the profile's name appended to "_config.toml".

        Args:
            app_data_dir (Path): The base directory where application data is stored.
            name (str | None): The profile name to resolve the settings path for.
                If None, the active profile name is used.

        Returns:
            Path: The resolved file path for the settings configuration.
        """
        if name is None:
            name = self._repo.get_state().active_settings_profile.name
        if name == "default":
            return app_data_dir / "config.toml"
        return app_data_dir / f"{name}_config.toml"
