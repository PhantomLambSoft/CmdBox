import unittest
from datetime import datetime
from peewee import IntegrityError

from cmdbox.database import db, get_db, ensure_schema
from cmdbox.repositories.errors import (
    ValidationError,
    ProfileConflictError,
    UnknownProfileError,
    ProfileNotEmptyError,
    ActiveProfileDeleteError,
    UpdateError,
)
from cmdbox.models import Profile, ProfileState, Command, Variable, ALL_MODELS
from cmdbox.repositories.profile_repository import ProfileRepository


class TestProfileRepository(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        get_db(testing=True)
        ensure_schema()

    @classmethod
    def tearDownClass(cls):
        db.drop_tables(ALL_MODELS)
        db.close()

    def setUp(self):
        # Clear tables before each test
        Command.delete().execute()
        Variable.delete().execute()
        ProfileState.delete().execute()
        Profile.delete().execute()
        
        # ProfileState needs exactly one row for get_state() to work if it's supposed to exist.
        # But wait, how is ProfileState initialized in the real app?
        # Usually by migrations or first run. 
        # In ProfileRepository.get_state(), it calls ProfileState.select().get()
        # This will fail if no row exists.
        
        self.repo = ProfileRepository()
        
        # Create a default profile to initialize ProfileState if needed, 
        # though usually migrations should handle it.
        # Let's see if we need to manually create the state row.
        self.default_profile = Profile.create(name="default")
        self.state = ProfileState.create(
            active_command_profile=self.default_profile,
            active_variable_profile=self.default_profile,
            active_settings_profile=self.default_profile
        )

    def _create_profile_group(self):
        self.p1 = Profile.create(name="alpha", description="First profile")
        self.p2 = Profile.create(name="beta", description="Second profile")
        self.p3 = Profile.create(name="gamma", description="Third profile")

    # =================================================================================
    # SECTION: CREATE TESTS
    # =================================================================================

    def test_create(self):
        profile = self.repo.create(name="test", description="Test description")
        self.assertTrue(isinstance(profile, Profile))
        self.assertEqual("test", profile.name)
        self.assertEqual("Test description", profile.description)
        self.assertIsNotNone(profile.date_created)
        # Check database count (default + test)
        self.assertEqual(2, Profile.select().count())

    def test_create_no_description_supplied(self):
        profile = self.repo.create(name="test")
        self.assertTrue(isinstance(profile, Profile))
        self.assertEqual("test", profile.name)
        self.assertIsNone(profile.description)

    def test_create_duplicate_name_raises_exception(self):
        self.repo.create(name="test")
        with self.assertRaises(ProfileConflictError):
            self.repo.create(name="test")

    def test_create_invalid_name_raises_exception(self):
        with self.assertRaises(ValidationError):
            self.repo.create(name="")
        with self.assertRaises(ValidationError):
            self.repo.create(name="  ")
        with self.assertRaises(ValidationError):
            self.repo.create(name="with spaces")

    # =================================================================================
    # SECTION: GET TESTS
    # =================================================================================

    def test_get_by_name(self):
        created = self.repo.create(name="test")
        retrieved = self.repo.get_by_name("test")
        self.assertEqual(created.id, retrieved.id)
        self.assertEqual("test", retrieved.name)

    def test_get_by_name_not_found_raises_exception(self):
        with self.assertRaises(UnknownProfileError):
            self.repo.get_by_name("nonexistent")

    def test_get_by_id(self):
        created = self.repo.create(name="test")
        retrieved = self.repo.get_by_id(created.id)
        self.assertEqual(created.id, retrieved.id)
        self.assertEqual("test", retrieved.name)

    def test_get_by_id_not_found_raises_exception(self):
        with self.assertRaises(UnknownProfileError):
            self.repo.get_by_id(999)

    # =================================================================================
    # SECTION: UPDATE TESTS
    # =================================================================================

    def test_update_name(self):
        profile = self.repo.create(name="old_name")
        updated = self.repo.update(profile, name="new_name")
        self.assertEqual("new_name", updated.name)
        
        # Verify in DB
        retrieved = Profile.get_by_id(profile.id)
        self.assertEqual("new_name", retrieved.name)

    def test_update_description(self):
        profile = self.repo.create(name="test", description="old desc")
        updated = self.repo.update(profile, description="new desc")
        self.assertEqual("new desc", updated.description)

    def test_update_name_to_duplicate_raises_exception(self):
        self.repo.create(name="other")
        profile = self.repo.create(name="test")
        with self.assertRaises(ProfileConflictError):
            self.repo.update(profile, name="other")

    def test_update_invalid_field_raises_exception(self):
        profile = self.repo.create(name="test")
        with self.assertRaises(ValidationError):
            self.repo.update(profile, invalid_field="value")

    def test_update_with_no_profile_raises_exception(self):
        with self.assertRaises(UpdateError):
            self.repo.update(None, name="new")

    def test_update_with_no_fields_raises_exception(self):
        profile = self.repo.create(name="test")
        with self.assertRaises(UpdateError):
            self.repo.update(profile)

    # =================================================================================
    # SECTION: DELETE TESTS
    # =================================================================================

    def test_delete_empty_profile(self):
        profile = self.repo.create(name="to_delete")
        result = self.repo.delete(profile)
        self.assertTrue(result)
        self.assertIsNone(Profile.get_or_none(Profile.name == "to_delete"))

    def test_delete_non_empty_profile_raises_exception(self):
        profile = self.repo.create(name="not_empty")
        Command.create(alias="cmd", template="echo", profile=profile)
        
        with self.assertRaises(ProfileNotEmptyError):
            self.repo.delete(profile)
            
        # Verify profile still exists
        self.assertIsNotNone(Profile.get_or_none(Profile.name == "not_empty"))

    def test_delete_non_empty_profile_with_force(self):
        profile = self.repo.create(name="not_empty")
        Command.create(alias="cmd", template="echo", profile=profile)
        Variable.create(name="var", value="val", profile=profile)
        
        result = self.repo.delete(profile, force=True)
        self.assertTrue(result)
        self.assertIsNone(Profile.get_or_none(Profile.name == "not_empty"))
        # Verify cascaded deletion
        self.assertEqual(0, Command.select().where(Command.profile == profile).count())
        self.assertEqual(0, Variable.select().where(Variable.profile == profile).count())

    def test_delete_active_command_profile_raises_exception(self):
        profile = self.repo.create(name="active")
        self.repo.set_active_command_profile(profile)
        
        with self.assertRaises(ActiveProfileDeleteError):
            self.repo.delete(profile)

    def test_delete_active_variable_profile_raises_exception(self):
        profile = self.repo.create(name="active")
        self.repo.set_active_variable_profile(profile)
        
        with self.assertRaises(ActiveProfileDeleteError):
            self.repo.delete(profile)

    def test_delete_active_settings_profile_raises_exception(self):
        profile = self.repo.create(name="active")
        self.repo.set_active_settings_profile(profile)
        
        with self.assertRaises(ActiveProfileDeleteError):
            self.repo.delete(profile)

    def test_delete_none_returns_false(self):
        self.assertFalse(self.repo.delete(None))

    # =================================================================================
    # SECTION: LIST AND UTIL TESTS
    # =================================================================================

    def test_list_all(self):
        self._create_profile_group()
        # default + alpha + beta + gamma = 4
        profiles = self.repo.list_all()
        self.assertEqual(4, len(profiles))

    def test_list_all_ordered_by_name(self):
        self._create_profile_group()
        profiles = self.repo.list_all(order_by="name")
        names = [p.name for p in profiles]
        self.assertEqual(["alpha", "beta", "default", "gamma"], names)

    def test_list_all_limit(self):
        self._create_profile_group()
        profiles = self.repo.list_all(limit=2)
        self.assertEqual(2, len(profiles))

    def test_record_use(self):
        profile = self.repo.create(name="test")
        self.assertIsNone(profile.last_used)
        
        self.repo.record_use(profile.id)
        
        updated = self.repo.get_by_id(profile.id)
        self.assertIsNotNone(updated.last_used)

    # =================================================================================
    # SECTION: STATE TESTS
    # =================================================================================

    def test_get_state(self):
        state = self.repo.get_state()
        self.assertTrue(isinstance(state, ProfileState))
        self.assertEqual(self.default_profile.id, state.active_command_profile_id)

    def test_set_active_command_profile(self):
        new_profile = self.repo.create(name="new")
        self.repo.set_active_command_profile(new_profile)
        
        state = self.repo.get_state()
        self.assertEqual(new_profile.id, state.active_command_profile_id)

    def test_set_active_variable_profile(self):
        new_profile = self.repo.create(name="new")
        self.repo.set_active_variable_profile(new_profile)
        
        state = self.repo.get_state()
        self.assertEqual(new_profile.id, state.active_variable_profile_id)

    def test_set_active_settings_profile(self):
        new_profile = self.repo.create(name="new")
        self.repo.set_active_settings_profile(new_profile)
        
        state = self.repo.get_state()
        self.assertEqual(new_profile.id, state.active_settings_profile_id)
