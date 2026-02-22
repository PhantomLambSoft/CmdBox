import unittest
from unittest.mock import MagicMock, patch, ANY
import typer
from cmdbox.cli.handlers.tag_handlers import (
    AddTagArgs,
    run_add_tag,
    run_get_tag,
    run_update_tag,
    run_list_tags,
    run_search_tags,
    run_delete_tag,
)
from cmdbox.services.field_selection import FieldSelectionResolver


class TestTagHandlers(unittest.TestCase):

    DISPLAY_FIELDS = [
        "f1",
        "f2",
        "test_field_one",
        "test_field_two",
        "test_field_three",
        "test_field_four",
        "default_field_one",
        "default_field_two",
    ]
    SEARCH_FIELDS = [
        "sf1",
        "sf2",
        "test_field_one",
        "test_field_two",
        "test_field_three",
        "test_field_four",
        "default_field_three",
        "default_field_four",
    ]

    def setUp(self):
        self.mock_tag_services = MagicMock()
        self.mock_console = MagicMock()
        self.mock_settings = MagicMock()
        self.get_tag_services = lambda: self.mock_tag_services
        self.get_console = lambda: self.mock_console
        self.get_settings = lambda: self.mock_settings
        self.get_display_field_resolver = lambda: FieldSelectionResolver(
            self.DISPLAY_FIELDS
        )
        self.get_search_field_resolver = lambda: FieldSelectionResolver(
            self.SEARCH_FIELDS
        )

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_created")
    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_name")
    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_description")
    def test_run_add_tag_interactive(
        self, mock_prompt_desc, mock_prompt_name, mock_render
    ):
        args = AddTagArgs(name=None, description=None, interactive=True)
        mock_prompt_name.return_value = "tag1"
        mock_prompt_desc.return_value = "desc1"
        mock_tag = MagicMock()
        self.mock_tag_services.create_tag.return_value = mock_tag
        mock_render.return_value = "rendered_created"

        run_add_tag(
            args=args,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_tag_services.create_tag.assert_called_with(
            name="tag1", description="desc1"
        )
        mock_render.assert_called_once_with(mock_tag)
        self.mock_console.print.assert_called_with("rendered_created")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_created")
    def test_run_add_tag_non_interactive(self, mock_render):
        args = AddTagArgs(name="tag1", description="desc1", interactive=False)
        mock_tag = MagicMock()
        self.mock_tag_services.create_tag.return_value = mock_tag
        mock_render.return_value = "rendered_created"

        run_add_tag(
            args=args,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_tag_services.create_tag.assert_called_with(
            name="tag1", description="desc1"
        )
        mock_render.assert_called_once_with(mock_tag)
        self.mock_console.print.assert_called_with("rendered_created")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag")
    def test_run_get_tag(self, mock_render):
        mock_tag = MagicMock()
        self.mock_tag_services.get_tag.return_value = mock_tag
        mock_render.return_value = "rendered_tag"
        run_get_tag(
            name="tag1",
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )
        self.mock_tag_services.get_tag.assert_called_with("tag1")
        mock_render.assert_called_once_with(mock_tag)
        self.mock_console.print.assert_called_with("rendered_tag")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_updated")
    def test_run_update_tag_with_fields_supplied(self, mock_render):
        mock_tag = MagicMock()
        mock_tag.id = 1
        self.mock_tag_services.get_tag.return_value = mock_tag
        mock_updated_tag = MagicMock()
        self.mock_tag_services.get_tag_by_id.return_value = mock_updated_tag
        mock_render.return_value = "rendered_updated"

        run_update_tag(
            name="tag1",
            description="new_desc",
            new_name="new_name",
            set_pairs=[],
            edit_mode=False,
            edit_fields=None,
            get_tag_services=self.get_tag_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_tag_services.update_tag.assert_called_with(
            "tag1", description="new_desc", name="new_name"
        )
        mock_render.assert_called_once_with(mock_updated_tag)
        self.mock_console.print.assert_called_with("rendered_updated")

    def test_run_update_tag_no_fields_edit_mode_false(self):
        with self.assertRaises(typer.BadParameter):
            run_update_tag(
                name="tag1",
                description=None,
                new_name=None,
                set_pairs=[],
                edit_mode=False,
                edit_fields=None,
                get_tag_services=self.get_tag_services,
                get_settings=self.get_settings,
                get_console=self.get_console,
            )

    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_description")
    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_name")
    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_updated")
    def test_run_update_tag_with_fields_edit_mode_true(
        self,
        mock_render,
        mock_prompt_name,
        mock_prompt_desc,
    ):
        mock_render.return_value = "rendered_updated"
        mock_prompt_name.return_value = "new_prompt_name"
        mock_prompt_desc.return_value = "new_prompt_desc"

        mock_tag = MagicMock()
        mock_tag.id = 1
        mock_tag.name = "tag1"
        mock_tag.description = "old_desc"
        self.mock_tag_services.get_tag.return_value = mock_tag
        mock_updated_tag = MagicMock()
        self.mock_tag_services.get_tag_by_id.return_value = mock_updated_tag

        run_update_tag(
            name="tag1",
            description=None,
            new_name=None,
            set_pairs=[],
            edit_mode=True,
            edit_fields=None,
            get_tag_services=self.get_tag_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_tag_services.update_tag.assert_called_with(
            "tag1", description="new_prompt_desc", name="new_prompt_name"
        )
        mock_render.assert_called_once_with(mock_updated_tag)
        mock_prompt_name.assert_called_once_with(ANY, default="tag1")
        mock_prompt_desc.assert_called_once_with(default="old_desc")
        self.mock_console.print.assert_called_with("rendered_updated")

    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_description")
    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_name")
    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_updated")
    def test_run_update_tag_with_fields_edit_mode_true_edit_fields_provided(
        self,
        mock_render,
        mock_prompt_name,
        mock_prompt_desc,
    ):
        mock_render.return_value = "rendered_updated"
        mock_prompt_name.return_value = "new_prompt_name"
        mock_prompt_desc.return_value = "new_prompt_desc"

        mock_tag = MagicMock()
        mock_tag.id = 1
        mock_tag.name = "tag1"
        mock_tag.description = "old_desc"
        self.mock_tag_services.get_tag.return_value = mock_tag
        mock_updated_tag = MagicMock()
        self.mock_tag_services.get_tag_by_id.return_value = mock_updated_tag

        run_update_tag(
            name="tag1",
            description=None,
            new_name=None,
            set_pairs=[],
            edit_mode=True,
            edit_fields="description",
            get_tag_services=self.get_tag_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_tag_services.update_tag.assert_called_with(
            "tag1", description="new_prompt_desc"
        )
        mock_render.assert_called_once_with(mock_updated_tag)
        mock_prompt_name.assert_not_called()
        mock_prompt_desc.assert_called_once_with(default="old_desc")
        self.mock_console.print.assert_called_with("rendered_updated")

    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_description")
    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_name")
    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_updated")
    def test_run_update_tag_with_fields_edit_mode_true_multiple_edit_fields_provided(
        self,
        mock_render,
        mock_prompt_name,
        mock_prompt_desc,
    ):
        mock_render.return_value = "rendered_updated"
        mock_prompt_name.return_value = "new_prompt_name"
        mock_prompt_desc.return_value = "new_prompt_desc"

        mock_tag = MagicMock()
        mock_tag.id = 1
        mock_tag.name = "tag1"
        mock_tag.description = "old_desc"
        self.mock_tag_services.get_tag.return_value = mock_tag
        mock_updated_tag = MagicMock()
        self.mock_tag_services.get_tag_by_id.return_value = mock_updated_tag

        run_update_tag(
            name="tag1",
            description=None,
            new_name=None,
            set_pairs=[],
            edit_mode=True,
            edit_fields="name, description",
            get_tag_services=self.get_tag_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_tag_services.update_tag.assert_called_with(
            "tag1", description="new_prompt_desc", name="new_prompt_name"
        )
        mock_render.assert_called_once_with(mock_updated_tag)
        mock_prompt_name.assert_called_once_with(ANY, default="tag1")
        mock_prompt_desc.assert_called_once_with(default="old_desc")
        self.mock_console.print.assert_called_with("rendered_updated")

    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_description")
    @patch("cmdbox.cli.handlers.tag_handlers.prompt_for_name")
    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_updated")
    def test_run_update_tag_with_fields_edit_mode_true_edit_fields_provided_aliases_taken_from_settings(
        self,
        mock_render,
        mock_prompt_name,
        mock_prompt_desc,
    ):
        mock_render.return_value = "rendered_updated"
        mock_prompt_name.return_value = "new_prompt_name"
        mock_prompt_desc.return_value = "new_prompt_desc"

        mock_tag = MagicMock()
        mock_tag.id = 1
        mock_tag.name = "tag1"
        mock_tag.description = "old_desc"
        self.mock_tag_services.get_tag.return_value = mock_tag
        mock_updated_tag = MagicMock()
        self.mock_tag_services.get_tag_by_id.return_value = mock_updated_tag
        self.mock_settings.field_aliases.alias_mapping = {"description": ["desc"]}

        run_update_tag(
            name="tag1",
            description=None,
            new_name=None,
            set_pairs=[],
            edit_mode=True,
            edit_fields="desc",
            get_tag_services=self.get_tag_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_tag_services.update_tag.assert_called_with(
            "tag1", description="new_prompt_desc"
        )
        mock_render.assert_called_once_with(mock_updated_tag)
        mock_prompt_name.assert_not_called()
        mock_prompt_desc.assert_called_once_with(default="old_desc")
        self.mock_console.print.assert_called_with("rendered_updated")

    def test_run_update_tag_in_edit_mode_with_fields_raises_error(self):
        with self.assertRaises(typer.BadParameter):
            run_update_tag(
                name="tag1",
                description="ClearBlueSky",
                new_name=None,
                set_pairs=[],
                edit_mode=True,
                edit_fields=None,
                get_tag_services=self.get_tag_services,
                get_settings=self.get_settings,
                get_console=self.get_console,
            )

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_updated")
    def test_run_update_tag_does_not_update_if_fields_are_unchanged(self, mock_render):
        mock_render.return_value = "rendered_updated"
        mock_tag = MagicMock()
        mock_tag.id = 1
        mock_tag.name = "tag1"
        mock_tag.description = "old_desc"
        self.mock_tag_services.get_tag.return_value = mock_tag
        mock_updated_tag = MagicMock()
        self.mock_tag_services.get_tag_by_id.return_value = mock_updated_tag

        run_update_tag(
            name="tag1",
            description="old_desc",
            new_name="tag1",
            set_pairs=[],
            edit_mode=False,
            edit_fields=None,
            get_tag_services=self.get_tag_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_tag_services.update_tag.assert_not_called()
        mock_render.assert_not_called()
        self.mock_console.info.assert_called_with("No changes detected.")

    def test_run_update_raises_error_when_field_variable_is_also_in_set(self):
        with self.assertRaises(typer.BadParameter):
            run_update_tag(
                name="tag1",
                description="GiveItAway",
                new_name=None,
                set_pairs=["description=CheckYesOrNo"],
                edit_mode=False,
                edit_fields=None,
                get_tag_services=self.get_tag_services,
                get_settings=self.get_settings,
                get_console=self.get_console,
            )

    def test_run_update_raises_error_when_set_is_provided_with_edit_flag(self):
        with self.assertRaises(typer.BadParameter):
            run_update_tag(
                name="tag1",
                description=None,
                new_name=None,
                set_pairs=["description=Troubadour"],
                edit_mode=True,
                edit_fields=None,
                get_tag_services=self.get_tag_services,
                get_settings=self.get_settings,
                get_console=self.get_console,
            )

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_list")
    def test_run_list_tags(self, mock_render):
        tags = ["t1", "t2"]
        self.mock_tag_services.list_tags.return_value = tags
        mock_render.return_value = "rendered_list"
        run_list_tags(
            limit=10,
            order_by="name",
            fields=["f1"],
            get_tag_services=self.get_tag_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
        )
        self.mock_tag_services.list_tags.assert_called_with(limit=10, order_by="name")
        mock_render.assert_called_once_with(tags, title="Tags", fields=["f1"])
        self.mock_console.print.assert_called_with("rendered_list")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_list")
    def test_run_list_tags_uses_correct_defaults_from_settings(self, mock_render):
        tags = ["t1", "t2"]
        self.mock_tag_services.list_tags.return_value = tags
        mock_render.return_value = "rendered_list"
        self.mock_settings.default_fields.tag_output = [
            "default_field_one",
            "default_field_two",
        ]
        run_list_tags(
            limit=10,
            order_by="name",
            fields=None,
            get_tag_services=self.get_tag_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
        )
        self.mock_tag_services.list_tags.assert_called_with(limit=10, order_by="name")
        mock_render.assert_called_once_with(
            tags, title="Tags", fields=["default_field_one", "default_field_two"]
        )
        self.mock_console.print.assert_called_with("rendered_list")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_list")
    def test_run_list_tags_uses_all_keyword_correctly(self, mock_render):
        tags = ["t1", "t2"]
        self.mock_tag_services.list_tags.return_value = tags
        mock_render.return_value = "rendered_list"
        run_list_tags(
            limit=10,
            order_by="name",
            fields=["all"],
            get_tag_services=self.get_tag_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
        )
        self.mock_tag_services.list_tags.assert_called_with(limit=10, order_by="name")
        mock_render.assert_called_once_with(
            tags, title="Tags", fields=self.DISPLAY_FIELDS
        )
        self.mock_console.print.assert_called_with("rendered_list")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_list")
    def test_run_search_tags(self, mock_render):
        tags = ["t1"]
        self.mock_tag_services.search.return_value = tags
        mock_render.return_value = "rendered_search"
        run_search_tags(
            term="term",
            limit=5,
            search_fields=["sf1"],
            fields=["f1"],
            get_tag_services=self.get_tag_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
            get_search_field_resolver=self.get_search_field_resolver,
        )
        self.mock_tag_services.search.assert_called_with(
            "term", limit=5, fields=["sf1"]
        )
        mock_render.assert_called_once_with(tags, title="Search Results", fields=["f1"])
        self.mock_console.print.assert_called_with("rendered_search")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_list")
    def test_run_search_tags_uses_correct_defaults_from_settings(self, mock_render):
        tags = ["t1"]
        self.mock_tag_services.search.return_value = tags
        mock_render.return_value = "rendered_search"
        self.mock_settings.default_fields.tag_output = [
            "default_field_one",
            "default_field_two",
        ]
        self.mock_settings.default_fields.tag_search = [
            "default_field_three",
            "default_field_four",
        ]
        run_search_tags(
            term="term",
            limit=5,
            search_fields=None,
            fields=None,
            get_tag_services=self.get_tag_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
            get_search_field_resolver=self.get_search_field_resolver,
        )
        self.mock_tag_services.search.assert_called_with(
            "term", limit=5, fields=["default_field_three", "default_field_four"]
        )
        mock_render.assert_called_once_with(
            tags,
            title="Search Results",
            fields=["default_field_one", "default_field_two"],
        )
        self.mock_console.print.assert_called_with("rendered_search")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_list")
    def test_run_search_tags_uses_all_keyword_correctly(self, mock_render):
        tags = ["t1"]
        self.mock_tag_services.search.return_value = tags
        mock_render.return_value = "rendered_search"
        run_search_tags(
            term="term",
            limit=5,
            search_fields=["all"],
            fields=["all"],
            get_tag_services=self.get_tag_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
            get_search_field_resolver=self.get_search_field_resolver,
        )
        self.mock_tag_services.search.assert_called_with(
            "term", limit=5, fields=self.SEARCH_FIELDS
        )
        mock_render.assert_called_once_with(
            tags, title="Search Results", fields=self.DISPLAY_FIELDS
        )
        self.mock_console.print.assert_called_with("rendered_search")

    @patch("cmdbox.cli.handlers.tag_handlers.render_tag_deleted")
    def test_run_delete_tag_success(self, mock_render):
        mock_tag = MagicMock()
        self.mock_tag_services.get_tag.return_value = mock_tag
        self.mock_tag_services.delete_tag.return_value = True
        mock_render.return_value = "rendered_deleted"

        run_delete_tag(
            name="tag1",
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_tag_services.delete_tag.assert_called_with("tag1")
        mock_render.assert_called_once_with(mock_tag)
        self.mock_console.print.assert_called_with("rendered_deleted")

    def test_run_delete_tag_failure(self):
        self.mock_tag_services.delete_tag.return_value = False
        run_delete_tag(
            name="tag1",
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )
        self.mock_console.error.assert_called()
