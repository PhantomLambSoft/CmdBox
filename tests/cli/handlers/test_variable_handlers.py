import unittest
from unittest.mock import MagicMock, patch, ANY
import typer
from cmdbox.cli.handlers.variable_handlers import (
    AddVariableArgs,
    run_add_variable,
    run_get_variable,
    run_update_variable,
    run_list_variables,
    run_search_variables,
    run_delete_variable,
)
from cmdbox.services.field_selection import FieldSelectionResolver


class TestVariableHandlers(unittest.TestCase):

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
        self.mock_var_services = MagicMock()
        self.mock_tag_services = MagicMock()
        self.mock_console = MagicMock()
        self.mock_settings = MagicMock()
        self.get_var_services = lambda: self.mock_var_services
        self.get_tag_services = lambda: self.mock_tag_services
        self.get_console = lambda: self.mock_console
        self.get_settings = lambda: self.mock_settings
        self.get_display_field_resolver = lambda: FieldSelectionResolver(
            self.DISPLAY_FIELDS
        )
        self.get_search_field_resolver = lambda: FieldSelectionResolver(
            self.SEARCH_FIELDS
        )

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_created")
    @patch("cmdbox.cli.handlers.variable_handlers.prompt_for_name")
    @patch("cmdbox.cli.handlers.variable_handlers.prompt_for_value")
    @patch("cmdbox.cli.handlers.variable_handlers.get_tags_interactive")
    def test_run_add_variable_interactive(
        self, mock_get_tags, mock_prompt_value, mock_prompt_name, mock_render
    ):
        args = AddVariableArgs(name=None, value=None, tags=None, interactive=True)
        mock_prompt_name.return_value = "var1"
        mock_prompt_value.return_value = "val1"
        mock_get_tags.return_value = ["tag1"]
        mock_var = MagicMock()
        self.mock_var_services.create_variable.return_value = mock_var
        mock_render.return_value = "rendered_created"

        run_add_variable(
            args=args,
            get_var_services=self.get_var_services,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_var_services.create_variable.assert_called_with(
            name="var1", value="val1", tags=["tag1"]
        )
        mock_render.assert_called_once_with(mock_var)
        self.mock_console.print.assert_called_with("rendered_created")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_created")
    def test_run_add_variable_non_interactive(self, mock_render):
        args = AddVariableArgs(
            name="var1", value="val1", tags=["tag1"], interactive=False
        )
        mock_var = MagicMock()
        self.mock_var_services.create_variable.return_value = mock_var
        mock_render.return_value = "rendered_created"

        run_add_variable(
            args=args,
            get_var_services=self.get_var_services,
            get_tag_services=self.get_tag_services,
            get_console=self.get_console,
        )

        self.mock_var_services.create_variable.assert_called_with(
            name="var1", value="val1", tags=["tag1"]
        )
        mock_render.assert_called_once_with(mock_var)
        self.mock_console.print.assert_called_with("rendered_created")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable")
    def test_run_get_variable(self, mock_render):
        mock_var = MagicMock()
        self.mock_var_services.get_variable.return_value = mock_var
        mock_render.return_value = "rendered_var"
        run_get_variable(
            name="var1",
            get_var_services=self.get_var_services,
            get_console=self.get_console,
        )
        self.mock_var_services.get_variable.assert_called_with("var1")
        mock_render.assert_called_once_with(mock_var)
        self.mock_console.print.assert_called_with("rendered_var")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_updated")
    def test_run_update_variable_with_fields_supplied(self, mock_render):
        mock_var = MagicMock()
        mock_var.id = 1
        self.mock_var_services.get_variable.return_value = mock_var
        mock_updated_var = MagicMock()
        self.mock_var_services.get_variable_by_id.return_value = mock_updated_var
        mock_render.return_value = "rendered_updated"

        run_update_variable(
            name="var1",
            value="new_val",
            new_name="new_name",
            set_pairs=None,
            edit_mode=False,
            edit_fields=None,
            get_var_services=self.get_var_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_var_services.update_variable.assert_called_with(
            "var1", value="new_val", name="new_name"
        )
        mock_render.assert_called_once_with(mock_updated_var)
        self.mock_console.print.assert_called_with("rendered_updated")

    def test_run_update_variable_no_fields_edit_mode_false(self):
        with self.assertRaises(typer.BadParameter):
            run_update_variable(
                name="var1",
                value=None,
                new_name=None,
                set_pairs=None,
                edit_mode=False,
                edit_fields=None,
                get_var_services=self.get_var_services,
                get_settings=self.get_settings,
                get_console=self.get_console,
            )

    @patch("cmdbox.cli.handlers.variable_handlers.prompt_for_value")
    @patch("cmdbox.cli.handlers.variable_handlers.prompt_for_name")
    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_updated")
    def test_run_update_variable_no_fields_edit_mode_true(
        self,
        mock_render,
        mock_prompt_name,
        mock_prompt_value,
    ):
        mock_render.return_value = "rendered_updated"
        mock_prompt_name.return_value = "new_prompt_name"
        mock_prompt_value.return_value = "new_prompt_val"

        mock_var = MagicMock()
        mock_var.id = 1
        mock_var.name = "old_name"
        mock_var.value = "old_val"

        self.mock_var_services.get_variable.return_value = mock_var
        mock_updated_var = MagicMock()
        self.mock_var_services.get_variable_by_id.return_value = mock_updated_var

        run_update_variable(
            name="var1",
            value=None,
            new_name=None,
            set_pairs=None,
            edit_mode=True,
            edit_fields=None,
            get_var_services=self.get_var_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_var_services.update_variable.assert_called_with(
            "var1", value="new_prompt_val", name="new_prompt_name"
        )
        mock_render.assert_called_once_with(mock_updated_var)
        mock_prompt_name.assert_called_with(ANY, default="old_name")
        mock_prompt_value.assert_called_with(default="old_val")
        self.mock_console.print.assert_called_with("rendered_updated")

    @patch("cmdbox.cli.handlers.variable_handlers.prompt_for_value")
    @patch("cmdbox.cli.handlers.variable_handlers.prompt_for_name")
    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_updated")
    def test_run_update_variable_no_fields_edit_mode_true_edit_field_provided(
        self,
        mock_render,
        mock_prompt_name,
        mock_prompt_value,
    ):
        mock_render.return_value = "rendered_updated"
        mock_prompt_name.return_value = "new_prompt_name"
        mock_prompt_value.return_value = "new_prompt_val"

        mock_var = MagicMock()
        mock_var.id = 1
        mock_var.name = "old_name"
        mock_var.value = "old_val"

        self.mock_var_services.get_variable.return_value = mock_var
        mock_updated_var = MagicMock()
        self.mock_var_services.get_variable_by_id.return_value = mock_updated_var

        run_update_variable(
            name="var1",
            value=None,
            new_name=None,
            set_pairs=None,
            edit_mode=True,
            edit_fields="value",
            get_var_services=self.get_var_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_var_services.update_variable.assert_called_with(
            "var1", value="new_prompt_val"
        )
        mock_render.assert_called_once_with(mock_updated_var)
        mock_prompt_name.assert_not_called()
        mock_prompt_value.assert_called_with(default="old_val")
        self.mock_console.print.assert_called_with("rendered_updated")

    @patch("cmdbox.cli.handlers.variable_handlers.prompt_for_value")
    @patch("cmdbox.cli.handlers.variable_handlers.prompt_for_name")
    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_updated")
    def test_run_update_variable_no_fields_edit_mode_true_multiple_edit_fields_provided(
        self,
        mock_render,
        mock_prompt_name,
        mock_prompt_value,
    ):
        mock_render.return_value = "rendered_updated"
        mock_prompt_name.return_value = "new_prompt_name"
        mock_prompt_value.return_value = "new_prompt_val"

        mock_var = MagicMock()
        mock_var.id = 1
        mock_var.name = "old_name"
        mock_var.value = "old_val"

        self.mock_var_services.get_variable.return_value = mock_var
        mock_updated_var = MagicMock()
        self.mock_var_services.get_variable_by_id.return_value = mock_updated_var

        run_update_variable(
            name="var1",
            value=None,
            new_name=None,
            set_pairs=None,
            edit_mode=True,
            edit_fields="name, value",
            get_var_services=self.get_var_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_var_services.update_variable.assert_called_with(
            "var1", value="new_prompt_val", name="new_prompt_name"
        )
        mock_render.assert_called_once_with(mock_updated_var)
        mock_prompt_name.assert_called_with(ANY, default="old_name")
        mock_prompt_value.assert_called_with(default="old_val")
        self.mock_console.print.assert_called_with("rendered_updated")

    @patch("cmdbox.cli.handlers.variable_handlers.prompt_for_value")
    @patch("cmdbox.cli.handlers.variable_handlers.prompt_for_name")
    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_updated")
    def test_run_update_variable_no_fields_edit_mode_true_edit_field_provided_as_aliases_from_settings(
        self,
        mock_render,
        mock_prompt_name,
        mock_prompt_value,
    ):
        mock_render.return_value = "rendered_updated"
        mock_prompt_name.return_value = "new_prompt_name"
        mock_prompt_value.return_value = "new_prompt_val"

        mock_var = MagicMock()
        mock_var.id = 1
        mock_var.name = "old_name"
        mock_var.value = "old_val"

        self.mock_var_services.get_variable.return_value = mock_var
        mock_updated_var = MagicMock()
        self.mock_var_services.get_variable_by_id.return_value = mock_updated_var
        self.mock_settings.field_aliases.alias_mapping = {"value": ["val"]}

        run_update_variable(
            name="var1",
            value=None,
            new_name=None,
            set_pairs=None,
            edit_mode=True,
            edit_fields="val",
            get_var_services=self.get_var_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_var_services.update_variable.assert_called_with(
            "var1", value="new_prompt_val"
        )
        mock_render.assert_called_once_with(mock_updated_var)
        mock_prompt_name.assert_not_called()
        mock_prompt_value.assert_called_with(default="old_val")
        self.mock_console.print.assert_called_with("rendered_updated")

    def test_run_update_variable_in_edit_mode_with_fields_raised_error(self):
        with self.assertRaises(typer.BadParameter):
            run_update_variable(
                name="var1",
                value="ICanStillMakeCheyenne",
                new_name=None,
                set_pairs=None,
                edit_mode=True,
                edit_fields=None,
                get_var_services=self.get_var_services,
                get_settings=self.get_settings,
                get_console=self.get_console,
            )

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_updated")
    def test_run_update_variable_does_not_update_if_updated_fields_are_unchanged(
        self, mock_render
    ):
        mock_var = MagicMock()
        mock_var.id = 1
        mock_var.name = "var1"
        mock_var.value = "old_val"
        self.mock_var_services.get_variable.return_value = mock_var
        mock_updated_var = MagicMock()
        self.mock_var_services.get_variable_by_id.return_value = mock_updated_var
        mock_render.return_value = "rendered_updated"

        run_update_variable(
            name="var1",
            value="old_val",
            new_name="var1",
            set_pairs=None,
            edit_mode=False,
            edit_fields=None,
            get_var_services=self.get_var_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
        )

        self.mock_var_services.update_variable.assert_not_called()
        mock_render.assert_not_called()
        self.mock_console.info.assert_called_with("No changes detected.")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_list")
    def test_run_list_variables(self, mock_render):
        vars_ = ["v1", "v2"]
        self.mock_var_services.list_variables.return_value = vars_
        mock_render.return_value = "rendered_list"
        run_list_variables(
            limit=10,
            order_by="name",
            tags=["t1"],
            fields=["f1"],
            get_var_services=self.get_var_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
        )
        self.mock_var_services.list_variables.assert_called_with(
            limit=10, order_by="name", tags=["t1"]
        )
        mock_render.assert_called_once_with(vars_, title="Variables", fields=["f1"])
        self.mock_console.print.assert_called_with("rendered_list")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_list")
    def test_run_list_variables_uses_correct_defaults_from_settings(self, mock_render):
        vars_ = ["v1", "v2", "v3"]
        self.mock_var_services.list_variables.return_value = vars_
        mock_render.return_value = "rendered_list"
        self.mock_settings.default_fields.variable_output = [
            "default_field_one",
            "default_field_two",
        ]
        run_list_variables(
            limit=10,
            order_by="name",
            tags=["t1"],
            fields=None,
            get_var_services=self.get_var_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
        )
        self.mock_var_services.list_variables.assert_called_with(
            limit=10, order_by="name", tags=["t1"]
        )
        mock_render.assert_called_once_with(
            vars_, title="Variables", fields=["default_field_one", "default_field_two"]
        )
        self.mock_console.print.assert_called_with("rendered_list")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_list")
    def test_run_list_variables_uses_all_keyword_correctly(self, mock_render):
        vars_ = ["v1", "v2"]
        self.mock_var_services.list_variables.return_value = vars_
        mock_render.return_value = "rendered_list"
        run_list_variables(
            limit=10,
            order_by="name",
            tags=["t1"],
            fields=["all"],
            get_var_services=self.get_var_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
        )
        self.mock_var_services.list_variables.assert_called_with(
            limit=10, order_by="name", tags=["t1"]
        )
        mock_render.assert_called_once_with(
            vars_, title="Variables", fields=self.DISPLAY_FIELDS
        )
        self.mock_console.print.assert_called_with("rendered_list")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_list")
    def test_run_search_variables(self, mock_render):
        vars_ = ["v1"]
        self.mock_var_services.search.return_value = vars_
        mock_render.return_value = "rendered_search"
        run_search_variables(
            term="term",
            limit=5,
            search_fields=["sf1"],
            fields=["f1"],
            get_var_services=self.get_var_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
            get_search_field_resolver=self.get_search_field_resolver,
        )
        self.mock_var_services.search.assert_called_with(
            "term", limit=5, fields=["sf1"]
        )
        mock_render.assert_called_once_with(
            vars_, title="Search Results", fields=["f1"]
        )
        self.mock_console.print.assert_called_with("rendered_search")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_list")
    def test_run_search_variables_uses_correct_defaults_from_settings(
        self, mock_render
    ):
        vars_ = ["v1", "v2", "v3"]
        self.mock_var_services.search.return_value = vars_
        mock_render.return_value = "rendered_search"
        self.mock_settings.default_fields.variable_output = [
            "default_field_one",
            "default_field_two",
        ]
        self.mock_settings.default_fields.variable_search = [
            "default_field_three",
            "default_field_four",
        ]
        run_search_variables(
            term="term",
            limit=5,
            search_fields=None,
            fields=None,
            get_var_services=self.get_var_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
            get_search_field_resolver=self.get_search_field_resolver,
        )
        self.mock_var_services.search.assert_called_with(
            "term",
            limit=5,
            fields=["default_field_three", "default_field_four"],  # search fields
        )
        mock_render.assert_called_once_with(
            vars_,
            title="Search Results",
            fields=["default_field_one", "default_field_two"],  # output fields
        )
        self.mock_console.print.assert_called_with("rendered_search")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_list")
    def test_run_search_variables_uses_all_keyword_correctly(self, mock_render):
        vars_ = ["v1"]
        self.mock_var_services.search.return_value = vars_
        mock_render.return_value = "rendered_search"
        run_search_variables(
            term="term",
            limit=5,
            search_fields=["all"],
            fields=["all"],
            get_var_services=self.get_var_services,
            get_settings=self.get_settings,
            get_console=self.get_console,
            get_display_field_resolver=self.get_display_field_resolver,
            get_search_field_resolver=self.get_search_field_resolver,
        )
        self.mock_var_services.search.assert_called_with(
            "term", limit=5, fields=self.SEARCH_FIELDS
        )
        mock_render.assert_called_once_with(
            vars_, title="Search Results", fields=self.DISPLAY_FIELDS
        )
        self.mock_console.print.assert_called_with("rendered_search")

    @patch("cmdbox.cli.handlers.variable_handlers.render_variable_deleted")
    def test_run_delete_variable_success(self, mock_render):
        mock_var = MagicMock()
        self.mock_var_services.get_variable.return_value = mock_var
        self.mock_var_services.delete_variable.return_value = True
        mock_render.return_value = "rendered_deleted"

        run_delete_variable(
            name="var1",
            get_var_services=self.get_var_services,
            get_console=self.get_console,
        )

        self.mock_var_services.delete_variable.assert_called_with("var1")
        mock_render.assert_called_once_with(mock_var)
        self.mock_console.print.assert_called_with("rendered_deleted")

    def test_run_delete_variable_failure(self):
        self.mock_var_services.delete_variable.return_value = False
        run_delete_variable(
            name="var1",
            get_var_services=self.get_var_services,
            get_console=self.get_console,
        )
        self.mock_console.error.assert_called()
