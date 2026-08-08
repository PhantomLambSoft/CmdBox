from typing import Sequence
from datetime import datetime

from peewee import IntegrityError

from .base_repository import BaseRepository
from .errors import (
    UnknownProfileError,
    ProfileConflictError,
    ProfileNotEmptyError,
    ActiveProfileDeleteError,
    UpdateError,
    ValidationError,
    DefaultProfileProtectionError,
)
from cmdbox.models import Profile, ProfileState, Command, Variable
from .validators import ProfileValidator


class ProfileRepository(BaseRepository):

    model = Profile

    def __init__(self, validator: ProfileValidator | None = None):
        self.validator = validator or ProfileValidator()

    def create(self, name: str, description: str | None = None) -> Profile:
        """
        Creates a new profile with the specified name and optional description.

        This method validates the input parameters and attempts to create a new
        profile using the provided details. If a profile already exists with the
        same name, a conflict error is raised.

        Args:
            name: The unique name of the profile to be created.
            description: An optional description for the profile.

        Returns:
            Profile: The newly created profile object.

        Raises:
            ProfileConflictError: If a profile with the given name already exists.
            IntegrityError: If any database integrity error occurs unrelated to name
                uniqueness.
        """
        self.validator.validate_create(name=name, description=description)
        try:
            return Profile.create(
                name=name,
                description=description,
                date_created=datetime.now(),
                last_used=None,
            )
        except IntegrityError as exc:
            if self._is_unique_name_violation(exc):
                raise ProfileConflictError(name=name) from exc
            raise

    def get_by_name(self, name: str) -> Profile:
        """
        Retrieves a profile by its name. If no profile is found with the specified name,
        an UnknownProfileError is raised.

        Args:
            name: The name of the profile to retrieve.

        Returns:
            Profile: The profile object corresponding to the specified name.

        Raises:
            UnknownProfileError: If no profile is found with the specified name.
        """
        profile = Profile.get_or_none(Profile.name == name)
        if profile is None:
            raise UnknownProfileError(name=name)
        return profile

    def get_by_id(self, profile_id: int) -> Profile:
        """
        Retrieves a Profile object from the database based on the provided profile ID.

        Args:
            profile_id (int): The unique identifier of the profile to retrieve.

        Returns:
            Profile: The Profile object associated with the provided profile ID.

        Raises:
            UnknownProfileError: If no profile is found with the given profile ID.
        """
        profile = Profile.get_or_none(Profile.id == profile_id)
        if profile is None:
            # Just use the profile_id as the name instead of creating another exception class
            raise UnknownProfileError(name=str(profile_id))
        return profile

    def update(self, profile: Profile, **fields) -> Profile:
        """
        Updates the fields of a given Profile instance and saves it.

        This method allows updating specific fields of a Profile object. It validates field
        names and values before applying updates. Additionally, it ensures that names are
        stripped of excess whitespace. If any issues occur during validation or saving,
        appropriate errors are raised.

        Args:
            profile (Profile): The Profile instance to update.
            **fields: Arbitrary keyword arguments representing field names to update and
                their corresponding values.

        Returns:
            Profile: The updated Profile instance.

        Raises:
            UpdateError: If no profile or fields are provided for update.
            ValidationError: If a provided field name is invalid.
            ProfileConflictError: If a unique constraint is violated for a profile name.
            IntegrityError: If any other database integrity error occurs while saving.
            DefaultProfileProtectionError: If the default profile is attempted to be renamed.
        """
        if not profile:
            raise UpdateError("No profile provided for update.")
        if not fields:
            raise UpdateError("No fields provided for update.")

        if profile.name == "default" and "name" in fields:
            new_name = fields["name"].strip() if fields["name"] else fields["name"]
            if new_name != "default":
                raise DefaultProfileProtectionError(action="renamed")

        if "name" in fields and fields.get("name") is not None:
            fields["name"] = fields.get("name").strip()

        try:
            for key, value in fields.items():
                if not hasattr(profile, key):
                    raise ValidationError(f"Invalid field: {key}")
                setattr(profile, key, value)
            profile.save()
            return profile
        except IntegrityError as exc:
            name = fields.get("name", "")
            if name and self._is_unique_name_violation(exc):
                raise ProfileConflictError(name=name) from exc
            raise

    def delete(self, profile: Profile, force: bool = False) -> bool:
        """
        Deletes a given profile from the system.

        This method removes the specified profile and its associated data, such as commands
        and variables, based on the provided conditions. If the profile is actively in use
        or contains associated commands or variables, the deletion operation may be blocked
        or require the force flag to proceed.

        Args:
            profile (Profile): The profile object to be deleted.
            force (bool, optional): Determines if the profile should be deleted even if it
                contains associated commands or variables. Defaults to False.

        Raises:
            ActiveProfileDeleteError: If the profile is currently active and cannot be deleted.
            ProfileNotEmptyError: If the profile contains commands or variables and the `force`
                flag is not set.

        Returns:
            bool: Returns True if the profile was successfully deleted, False otherwise.
        """
        if not profile:
            return False

        if profile.name == "default":
            raise DefaultProfileProtectionError(action="deleted")

        state = self.get_state()
        if profile.id in (
            state.active_command_profile_id,
            state.active_variable_profile_id,
            state.active_settings_profile_id,
        ):
            raise ActiveProfileDeleteError(name=profile.name)

        command_count = Command.select().where(Command.profile == profile).count()
        variable_count = Variable.select().where(Variable.profile == profile).count()

        if (command_count or variable_count) and not force:
            raise ProfileNotEmptyError(
                name=profile.name,
                command_count=command_count,
                variable_count=variable_count,
            )

        profile.delete_instance(recursive=force)
        return True

    def list_all(
        self, order_by: str | Sequence[str] = "name", limit: int = 25
    ) -> list[Profile]:
        """
        Retrieves a list of all Profile objects from the database with optional ordering and
        limiting of results.

        Args:
            order_by (str | Sequence[str]): The column(s) to order the results by. Defaults to "name".
            limit (int): Maximum number of Profile objects to retrieve. Defaults to 25.

        Returns:
            list[Profile]: A list of Profile objects sorted and limited as specified.
        """
        ordering = self._resolve_ordering(order_by)
        return list(Profile.select().order_by(*ordering).limit(limit))

    def search(
        self,
        query: str,
        fields: str | Sequence[str] | None = ("name", "description"),
        limit: int = 25,
    ) -> list[Profile]:
        """
        Searches for profiles based on a given query and specified fields, returning a list of
        profiles that match the search criteria. The search operation is limited by the provided
        limit parameter and uses "name" as the secondary ordering criterion.

        Args:
            query (str): The search query string used to find matching profiles.
            fields (str | Sequence[str] | None): The fields to search within. Defaults to
                ("name", "description") if not provided.
            limit (int): The maximum number of profiles to return. Defaults to 25.

        Returns:
            list[Profile]: A list of Profile objects that match the search query.
        """
        return self._search(
            query,
            secondary_ordering="name",
            fields=fields,
            limit=limit,
        )

    def record_use(self, profile_id: int) -> None:
        """
        Records the usage of a profile by updating its last used timestamp.

        Args:
            profile_id (int): The unique identifier of the profile to update.
        """
        Profile.update(last_used=datetime.now()).where(
            Profile.id == profile_id
        ).execute()

    def get_state(self) -> ProfileState:
        """
        Retrieves the single ProfileState row.

        Returns:
            ProfileState: The current active profile state.
        """
        return ProfileState.select().get()

    def set_active_command_profile(self, profile: Profile) -> ProfileState:
        """
        Sets the active command profile and saves the updated state.

        This method assigns the provided profile to the active command profile
        within the current state. The modified state is then saved and returned.

        Args:
            profile (Profile): The profile to set as the active command profile.

        Returns:
            ProfileState: The updated state containing the new active command
            profile.
        """
        state = self.get_state()
        state.active_command_profile = profile
        state.save()
        return state

    def set_active_variable_profile(self, profile: Profile) -> ProfileState:
        """
        Sets the active variable profile and saves the updated state.

        This method assigns the provided profile to the active variable profile
        within the current state. The modified state is then saved and returned.

        Args:
            profile (Profile): The profile to set as the active variable profile.

        Returns:
            ProfileState: The updated state containing the new active variable
            profile.
        """
        state = self.get_state()
        state.active_variable_profile = profile
        state.save()
        return state

    def set_active_settings_profile(self, profile: Profile) -> ProfileState:
        """
        Sets the active settings profile and saves the updated state.

        This method assigns the provided profile to the active settings profile
        within the current state. The modified state is then saved and returned.

        Args:
            profile (Profile): The profile to set as the active settings profile.

        Returns:
            ProfileState: The updated state containing the new active settings
            profile.
        """
        state = self.get_state()
        state.active_settings_profile = profile
        state.save()
        return state
